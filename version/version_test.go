package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVersion(t *testing.T) {
	assert := assert.New(t)
	v, err := NewFromHead(&RepoHead{LastTag: "1.2.3"})
	assert.NoError(err)
	assert.Equal(Version{Major: 1, Minor: 2, Patch: 3}, v)
}

func TestNewVersionInvalid(t *testing.T) {
	assert := assert.New(t)
	for _, s := range []string{
		"1.2",
		"1.2.a",
		"1.a.3",
		"a.2.3",
	} {
		_, err := NewFromHead(&RepoHead{LastTag: s})
		assert.Error(err)
	}
}

func TestParse(t *testing.T) {
	assert := assert.New(t)

	for _, test := range []struct {
		ref RepoHead
		v   Version
	}{
		{
			RepoHead{LastTag: "1.2.3", CommitsSinceTag: 4, Hash: "fcf2c8fa"},
			Version{Major: 1, Minor: 2, Patch: 3, Commits: 4, Meta: "fcf2c8fa"},
		},
		{
			RepoHead{},
			Version{},
		},
		{
			RepoHead{LastTag: "1.2.3-rc.1"},
			Version{Major: 1, Minor: 2, Patch: 3, preRelease: "rc.1"},
		},
		{
			RepoHead{LastTag: "1.2.3-rc.1", CommitsSinceTag: 2, Hash: "gd92f0b2"},
			Version{Major: 1, Minor: 2, Patch: 3, preRelease: "rc.1", Commits: 2, Meta: "gd92f0b2"},
		},
		{
			RepoHead{LastTag: "3.2.1"},
			Version{Major: 3, Minor: 2, Patch: 1},
		},
		{
			RepoHead{LastTag: "v3.2.1"},
			Version{Prefix: "v", Major: 3, Minor: 2, Patch: 1},
		},
		{
			RepoHead{LastTag: "3.2.1-liftoff.alpha.1", CommitsSinceTag: 3, Hash: "fcf2c8fa"},
			Version{Major: 3, Minor: 2, Patch: 1, preRelease: "liftoff.alpha.1", Commits: 3, Meta: "fcf2c8fa"},
		},
		{
			RepoHead{LastTag: "3.2.1+special"},
			Version{Major: 3, Minor: 2, Patch: 1, Meta: "special"},
		},
		{
			RepoHead{LastTag: "3.2.1-rc.2+special"},
			Version{Major: 3, Minor: 2, Patch: 1, preRelease: "rc.2", Meta: "special"},
		},
		{
			RepoHead{LastTag: "3.2.1-rc.2+special", CommitsSinceTag: 3, Hash: "gd92f0b2"},
			Version{Major: 3, Minor: 2, Patch: 1, preRelease: "rc.2", Commits: 3, Meta: "special"},
		},
	} {
		v, err := NewFromHead(&test.ref)
		assert.NoError(err)
		assert.Equal(test.v, v)
	}
}

func TestString(t *testing.T) {
	assert := assert.New(t)
	for _, test := range []struct {
		v Version
		s string
	}{
		{
			Version{Major: 1, Minor: 2, Patch: 3, Commits: 10, Meta: "fcf2c8f"},
			"1.2.4-dev.10+fcf2c8f",
		},
		{
			Version{Major: 0, Minor: 3, Patch: 1},
			"0.3.1",
		},
		{
			Version{Major: 1, Minor: 3, Patch: 0, preRelease: "rc.3"},
			"1.3.0-rc.3",
		},
		{
			Version{Major: 2, Minor: 5, Patch: 0, preRelease: "rc.3", Commits: 3},
			"2.5.0-rc.3.dev.3",
		},
	} {
		assert.Equal(test.s, test.v.String())
	}

}

func TestFormat(t *testing.T) {
	assert := assert.New(t)
	v := Version{Major: 1, Minor: 2, Patch: 3, Commits: 10, Meta: "fcf2c8f"}
	for _, test := range []struct {
		f string
		p string
		s string
	}{
		{
			FullFormat,
			"",
			"1.2.4-dev.10+fcf2c8f",
		},
		{
			NoMetaFormat,
			"",
			"1.2.4-dev.10",
		},
		{
			NoPreFormat,
			"",
			"1.2.4",
		},
		{
			NoPatchFormat,
			"",
			"1.2",
		},
		{
			NoMinorFormat,
			"v",
			"v1",
		},
		{
			"x.y-p",
			"v",
			"v1.2-dev.10",
		},
	} {
		v.Prefix = test.p
		s, err := v.Format(test.f)
		assert.NoError(err)
		assert.Equal(test.s, s)
	}
}

func TestInvalidFormat(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	s, err := v.Format("q")
	assert.EqualError(t, err, "invalid format: q")
	assert.Equal(t, "", s)
}
