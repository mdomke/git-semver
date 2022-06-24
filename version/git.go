package version

import (
	"fmt"
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

// GitDescribe looks at the git respository at path and figures
// out versioning relvant information about the head commit.
func GitDescribe(path string) (*RepoHead, error) {
	repo, err := git.PlainOpen(path)
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
	tags, err := getTagMap(repo)
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
		ref.CommitsSinceTag += 1
		return nil
	})
	return &ref, nil
}

type Tag struct {
	Name string
	When time.Time
}

func getTagMap(repo *git.Repository) (map[string]Tag, error) {
	tags, err := repo.Tags()
	if err != nil {
		return nil, err
	}
	result := make(map[string]Tag)
	if err = tags.ForEach(func(r *plumbing.Reference) error {
		tag, err := repo.TagObject(r.Hash())
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
			result[hash] = Tag{Name: tag.Name, When: tag.Tagger.When}
		case plumbing.ErrObjectNotFound:
			commit, err := repo.CommitObject(r.Hash())
			if err != nil {
				return nil
			}
			hash := commit.Hash.String()
			if c, ok := result[hash]; ok && !commit.Committer.When.After(c.When) {
				return nil
			}
			result[hash] = Tag{Name: r.Name().Short(), When: commit.Committer.When}
		default:
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	return result, nil
}
