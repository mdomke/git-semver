package version

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// RepoHead provides statistics about the head commit of a git
// repository like its commit-ash, the number of commits since
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

	commits, err := repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list commits: %w", err)
	}

	_ = commits.ForEach(func(c *object.Commit) error {
		ref.LastTag = (*tags)[c.Hash.String()]
		if ref.LastTag != "" {
			return storer.ErrStop
		}
		ref.CommitsSinceTag += 1
		return nil
	})
	return &ref, nil
}

func getTagMap(repo *git.Repository) (*map[string]string, error) {
	tags, err := repo.Tags()
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	if err = tags.ForEach(func(r *plumbing.Reference) error {
		tag, err := repo.TagObject(r.Hash())
		switch err {
		case nil:
			commit, err := tag.Commit()
			if err != nil {
				return nil
			}
			result[commit.Hash.String()] = tag.Name
		case plumbing.ErrObjectNotFound:
			result[r.Hash().String()] = r.Name().Short()
		default:
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	return &result, nil
}
