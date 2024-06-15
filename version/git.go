package version

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// RepoHead provides statistics about the head commit of a git
// repository like its commit-hash, the number of commits since
// the last tag and the name of the last tag.
type RepoHead struct {
	LastTag         string
	CommitsSinceTag int
	Hash            string
}

type options struct {
	matchFunc func(string) bool
}

type Option = func(*options)

func WithMatchPattern(pattern string) Option {
	return func(opts *options) {
		opts.matchFunc = func(tagName string) bool {
			if pattern == "" {
				return true
			}
			matched, err := filepath.Match(pattern, tagName)
			if err != nil {
				fmt.Printf("Ignoring invalid match pattern: %s: %s\n", pattern, err)
				return true
			}
			return matched
		}
	}
}

// GitDescribe looks at the git repository at path and figures
// out versioning relvant information about the head commit.
func GitDescribe(path string, opts ...Option) (*RepoHead, error) {
	options := options{matchFunc: func(string) bool { return true }}
	for _, apply := range opts {
		apply(&options)
	}

	openOpts := git.PlainOpenOptions{DetectDotGit: true}
	repo, err := git.PlainOpenWithOptions(path, &openOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to open repo: %w", err)
	}
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve repo head: %w", err)
	}

	ref := RepoHead{
		Hash: head.Hash().String(),
	}
	tags, err := getTagMap(repo, options.matchFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tag-list: %w", err)
	}

	if tag, found := tags[ref.Hash]; found {
		ref.LastTag = tag.Name
		return &ref, nil
	}

	commits, err := repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list commits: %w", err)
	}

	_ = commits.ForEach(func(c *object.Commit) error {
		tag, ok := tags[c.Hash.String()]
		if ok {
			ref.LastTag = tag.Name
			return storer.ErrStop
		}
		ref.CommitsSinceTag++
		return nil
	})
	return &ref, nil
}

type Tag struct {
	Name string
	When time.Time
}

func getTagMap(repo *git.Repository, match func(string) bool) (map[string]Tag, error) {
	tags, err := repo.Tags()
	if err != nil {
		return nil, err
	}
	result := make(map[string]Tag)
	if err = tags.ForEach(func(ref *plumbing.Reference) error {
		tag, err := repo.TagObject(ref.Hash())
		switch err {
		case nil:
			commit, err := tag.Commit()
			if err != nil {
				return nil
			}
			hash := commit.Hash.String()
			if t, ok := result[hash]; ok && !tag.Tagger.When.After(t.When) {
				return nil
			}
			if match(tag.Name) {
				result[hash] = Tag{Name: tag.Name, When: tag.Tagger.When}
			}
		case plumbing.ErrObjectNotFound:
			commit, err := repo.CommitObject(ref.Hash())
			if err != nil {
				return nil
			}
			tagName := ref.Name().Short()
			if !match(tagName) {
				return nil
			}
			hash := commit.Hash.String()
			c, ok := result[hash]
			if !ok {
				result[hash] = Tag{Name: tagName, When: commit.Committer.When}
				return nil
			}
			// two tags on the same commit, select the larger one.
			h0 := RepoHead{c.Name, 0, hash}
			h1 := RepoHead{tagName, 0, hash}
			v0, err0 := NewFromHead(&h0, "")
			v1, err1 := NewFromHead(&h1, "")
			if err0 != nil || (err1 == nil && v1.Compare(&v0) > 0) {
				result[hash] = Tag{Name: tagName, When: commit.Committer.When}
			}
			return nil
		default:
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	return result, nil
}
