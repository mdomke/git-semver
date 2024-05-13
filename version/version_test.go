package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVersion(t *testing.T) {
	assert := assert.New(t)
	v, err := NewFromHead(&RepoHead{LastTag: "1.2.3"}, "")
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
		_, err := NewFromHead(&RepoHead{LastTag: s}, "")
		assert.Error(err)
	}
}

func TestParse(t *testing.T) {
	assert := assert.New(t)

	for _, test := range []struct {
		ref    RepoHead
		v      Version
		prefix string
	}{
		{
			ref: RepoHead{LastTag: "1.2.3", CommitsSinceTag: 4, Hash: "fcf2c8fa"},
			v:   Version{Major: 1, Minor: 2, Patch: 3, Commits: 4, Meta: "fcf2c8fa"},
		},
		{
			ref: RepoHead{},
			v:   Version{},
		},
		{
			ref: RepoHead{LastTag: "1.2.3-rc.1"},
			v:   Version{Major: 1, Minor: 2, Patch: 3, preRelease: "rc.1"},
		},
		{
			ref: RepoHead{LastTag: "1.2.3-rc.1", CommitsSinceTag: 2, Hash: "gd92f0b2"},
			v:   Version{Major: 1, Minor: 2, Patch: 3, preRelease: "rc.1", Commits: 2, Meta: "gd92f0b2"},
		},
		{
			ref: RepoHead{LastTag: "3.2.1"},
			v:   Version{Major: 3, Minor: 2, Patch: 1},
		},
		{
			ref: RepoHead{LastTag: "v3.2.1"},
			v:   Version{Prefix: "v", Major: 3, Minor: 2, Patch: 1},
		},
		{
			ref:    RepoHead{LastTag: "ver3.2.1"},
			v:      Version{Prefix: "ver", Major: 3, Minor: 2, Patch: 1},
			prefix: "ver",
		},
		{
			ref: RepoHead{LastTag: "3.2.1-liftoff.alpha.1", CommitsSinceTag: 3, Hash: "fcf2c8fa"},
			v:   Version{Major: 3, Minor: 2, Patch: 1, preRelease: "liftoff.alpha.1", Commits: 3, Meta: "fcf2c8fa"},
		},
		{
			ref: RepoHead{LastTag: "3.5.0-liftoff-alpha.1"},
			v:   Version{Major: 3, Minor: 5, Patch: 0, preRelease: "liftoff-alpha.1"},
		},
		{
			ref: RepoHead{LastTag: "3.2.1+special"},
			v:   Version{Major: 3, Minor: 2, Patch: 1, Meta: "special"},
		},
		{
			ref: RepoHead{LastTag: "3.2.1-rc.2+special"},
			v:   Version{Major: 3, Minor: 2, Patch: 1, preRelease: "rc.2", Meta: "special"},
		},
		{
			ref: RepoHead{LastTag: "3.2.1-rc.2+special", CommitsSinceTag: 3, Hash: "gd92f0b2"},
			v:   Version{Major: 3, Minor: 2, Patch: 1, preRelease: "rc.2", Commits: 3, Meta: "special"},
		},
	} {
		v, err := NewFromHead(&test.ref, test.prefix)
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
			"1.2.3-dev.10+fcf2c8f",
		},
		{
			Version{Major: 0, Minor: 3, Patch: 1},
			"0.3.1",
		},
		{
			Version{Prefix: "v", Major: 0, Minor: 3, Patch: 1},
			"v0.3.1",
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
			"1.2.3-dev.10+fcf2c8f",
		},
		{
			NoMetaFormat,
			"",
			"1.2.3-dev.10",
		},
		{
			NoPreFormat,
			"",
			"1.2.3",
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
		{
			FullFormat,
			"",
			"1.2.3-dev.10+fcf2c8f",
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

func TestReleaseToString(t *testing.T) {
	assert.PanicsWithError(t, "unexpected target component 8", func() {
		v := Target(8)
		_ = v.String()
	})

	v := Devel
	assert.Equal(t, "dev", v.String())

	v = Patch
	assert.Equal(t, "patch", v.String())

	v = Minor
	assert.Equal(t, "minor", v.String())

	v = Major
	assert.Equal(t, "major", v.String())
}

func TestParseRelease(t *testing.T) {
	var v Target
	assert.EqualError(t, v.Set("foo"), "parse error")

	assert.NoError(t, v.Set("dev"))
	assert.Equal(t, Devel, v)

	assert.NoError(t, v.Set("patch"))
	assert.Equal(t, Patch, v)

	assert.NoError(t, v.Set("minor"))
	assert.Equal(t, Minor, v)

	assert.NoError(t, v.Set("major"))
	assert.Equal(t, Major, v)
}
