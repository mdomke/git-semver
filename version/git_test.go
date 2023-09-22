package version

import (
	"os"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

func TestGitDescribe(t *testing.T) {
	assert := assert.New(t)
	dir, _ := os.MkdirTemp("", "example")
	repo, err := git.PlainInit(dir, false)
	assert.NoError(err)

	worktree, err := repo.Worktree()
	assert.NoError(err)

	test := func(expected *RepoHead, opts ...Option) {
		actual, err := GitDescribe(dir, opts...)
		assert.NoError(err)
		assert.Equal(expected, actual)
	}

	now := time.Now().UTC()
	author := &object.Signature{
		Name:  "John Doe",
		Email: "john@doe.org",
		When:  now,
	}
	opts := git.CommitOptions{
		Author:            author,
		Committer:         author,
		AllowEmptyCommits: true,
	}

	commit1, err := worktree.Commit("first commit", &opts)
	assert.NoError(err)
	test(&RepoHead{Hash: commit1.String(), CommitsSinceTag: 1})

	tag1, err := repo.CreateTag("1.0.0", commit1, nil)
	assert.NoError(err)
	test(&RepoHead{
		LastTag:         tag1.Name().Short(),
		Hash:            commit1.String(),
		CommitsSinceTag: 0,
	})

	author.When = author.When.Add(1 * time.Hour)
	tag1Post, err := repo.CreateTag("v1.0.1", commit1, &git.CreateTagOptions{
		Tagger:  author,
		Message: "annotated tag",
	})
	assert.NoError(err)
	test(&RepoHead{
		LastTag:         tag1Post.Name().Short(),
		Hash:            commit1.String(),
		CommitsSinceTag: 0,
	})

	test(&RepoHead{
		LastTag:         tag1.Name().Short(),
		Hash:            commit1.String(),
		CommitsSinceTag: 0,
	}, WithMatchPattern("1.*.*"))

	author.When = author.When.Add(1 * time.Hour)
	commit2, err := worktree.Commit("second commit", &opts)
	assert.NoError(err)
	test(&RepoHead{
		LastTag:         tag1Post.Name().Short(),
		Hash:            commit2.String(),
		CommitsSinceTag: 1,
	})

	author.When = author.When.Add(1 * time.Second)
	tag2, err := repo.CreateTag("v2.0.0-rc.1", commit2, &git.CreateTagOptions{
		Tagger:  author,
		Message: "looks like the final release",
	})
	assert.NoError(err)
	test(&RepoHead{
		LastTag:         tag2.Name().Short(),
		Hash:            commit2.String(),
		CommitsSinceTag: 0,
	})

	author.When = author.When.Add(1 * time.Second)
	tag3, err := repo.CreateTag("v2.0.0", commit2, &git.CreateTagOptions{
		Tagger:  author,
		Message: "the final release",
	})
	assert.NoError(err)
	test(&RepoHead{
		LastTag:         tag3.Name().Short(),
		Hash:            commit2.String(),
		CommitsSinceTag: 0,
	})
}

func TestGitDescribeError(t *testing.T) {
	assert := assert.New(t)
	dir, _ := os.MkdirTemp("", "example")

	test := func(msg string) {
		head, err := GitDescribe(dir)
		assert.Nil(head)
		assert.EqualError(err, msg)
	}
	test("failed to open repo: repository does not exist")

	_, err := git.PlainInit(dir, false)
	assert.NoError(err)
	test("failed to retrieve repo head: reference not found")
}
