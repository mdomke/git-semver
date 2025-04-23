package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVersion(t *testing.T) {
	v, err := NewFromHead(&RepoHead{LastTag: "1.2.3"}, "")
	require.NoError(t, err)
	assert.Equal(t, Version{Major: 1, Minor: 2, Patch: 3}, v)
}

func TestNewVersionInvalid(t *testing.T) {
	for _, tagName := range []string{
		"1.2",
		"1.2.a",
		"1.a.3",
		"a.2.3",
	} {
		_, err := NewFromHead(&RepoHead{LastTag: tagName}, "")
		require.Error(t, err)
	}
}

func TestParse(t *testing.T) {
	for _, test := range []struct {
		ref    RepoHead
		ver    Version
		prefix string
	}{
		{
			ref: RepoHead{LastTag: "1.2.3", CommitsSinceTag: 4, Hash: "fcf2c8fa"},
			ver: Version{Major: 1, Minor: 2, Patch: 3, Commits: 4, Meta: "fcf2c8fa"},
		},
		{
			ref: RepoHead{},
			ver: Version{},
		},
		{
			ref: RepoHead{LastTag: "1.2.3-rc.1"},
			ver: Version{Major: 1, Minor: 2, Patch: 3, preRelease: "rc.1"},
		},
		{
			ref: RepoHead{LastTag: "1.2.3-rc.1", CommitsSinceTag: 2, Hash: "gd92f0b2"},
			ver: Version{Major: 1, Minor: 2, Patch: 3, preRelease: "rc.1", Commits: 2, Meta: "gd92f0b2"},
		},
		{
			ref: RepoHead{LastTag: "3.2.1"},
			ver: Version{Major: 3, Minor: 2, Patch: 1},
		},
		{
			ref: RepoHead{LastTag: "v3.2.1"},
			ver: Version{Prefix: "v", Major: 3, Minor: 2, Patch: 1},
		},
		{
			ref:    RepoHead{LastTag: "ver3.2.1"},
			ver:    Version{Prefix: "ver", Major: 3, Minor: 2, Patch: 1},
			prefix: "ver",
		},
		{
			ref: RepoHead{LastTag: "3.2.1-liftoff.alpha.1", CommitsSinceTag: 3, Hash: "fcf2c8fa"},
			ver: Version{Major: 3, Minor: 2, Patch: 1, preRelease: "liftoff.alpha.1", Commits: 3, Meta: "fcf2c8fa"},
		},
		{
			ref: RepoHead{LastTag: "3.5.0-liftoff-alpha.1"},
			ver: Version{Major: 3, Minor: 5, Patch: 0, preRelease: "liftoff-alpha.1"},
		},
		{
			ref: RepoHead{LastTag: "3.2.1+special"},
			ver: Version{Major: 3, Minor: 2, Patch: 1, Meta: "special"},
		},
		{
			ref: RepoHead{LastTag: "3.2.1-rc.2+special"},
			ver: Version{Major: 3, Minor: 2, Patch: 1, preRelease: "rc.2", Meta: "special"},
		},
		{
			ref: RepoHead{LastTag: "3.2.1-rc.2+special", CommitsSinceTag: 3, Hash: "gd92f0b2"},
			ver: Version{Major: 3, Minor: 2, Patch: 1, preRelease: "rc.2", Commits: 3, Meta: "special"},
		},
	} {
		v, err := NewFromHead(&test.ref, test.prefix)
		require.NoError(t, err)
		assert.Equal(t, test.ver, v)
	}
}

func TestString(t *testing.T) {
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
		assert.Equal(t, test.s, test.v.String())
	}

}

func TestFormat(t *testing.T) {
	ver := Version{Major: 1, Minor: 2, Patch: 3, Commits: 10, Meta: "fcf2c8f"}
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
		ver.Prefix = test.p
		s, err := ver.Format(test.f)
		require.NoError(t, err)
		assert.Equal(t, test.s, s)
	}
}

func TestInvalidFormat(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	s, err := v.Format("q")
	require.EqualError(t, err, "invalid format: q")
	assert.Equal(t, "", s)
}

func TestReleaseToString(t *testing.T) {
	assert.PanicsWithError(t, "unexpected target component 8", func() {
		target := Target(8)
		_ = target.String()
	})

	target := Devel
	assert.Equal(t, "dev", target.String())

	target = Patch
	assert.Equal(t, "patch", target.String())

	target = Minor
	assert.Equal(t, "minor", target.String())

	target = Major
	assert.Equal(t, "major", target.String())
}

func TestParseRelease(t *testing.T) {
	var target Target
	require.EqualError(t, target.Set("foo"), "parse error")

	require.NoError(t, target.Set("dev"))
	assert.Equal(t, Devel, target)

	require.NoError(t, target.Set("patch"))
	assert.Equal(t, Patch, target)

	require.NoError(t, target.Set("minor"))
	assert.Equal(t, Minor, target)

	require.NoError(t, target.Set("major"))
	assert.Equal(t, Major, target)
}

func TestVersionCompare(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	assert.Equal(t, v.Compare(&Version{Major: 1, Minor: 3, Patch: 0}), -1)
	assert.Equal(t, v.Compare(&Version{Major: 1, Minor: 0, Patch: 0}), 2)
	assert.Equal(t, v.Compare(&Version{Major: 1, Minor: 2, Patch: 3}), 0)
	assert.Equal(t, v.Compare(&Version{Major: 1, Minor: 2, Patch: 3}), 0)

	v = Version{Major: 1, Minor: 2, Patch: 3, Meta: "dev.2"}
	assert.Equal(t, v.Compare(&Version{Major: 1, Minor: 2, Patch: 3, Meta: "dev.1"}), 1)
	assert.Equal(t, v.Compare(&Version{Major: 1, Minor: 2, Patch: 3, Meta: "dev.5"}), -1)
	assert.Equal(t, v.Compare(&Version{Major: 1, Minor: 2, Patch: 3, Meta: "dev.2"}), 0)
}
