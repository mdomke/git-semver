package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVersion(t *testing.T) {
	assert := assert.New(t)
	v := Version{}
	err := parseVersion("1.2.3", &v)
	assert.NoError(err)
	assert.Equal(Version{Major: 1, Minor: 2, Patch: 3}, v)
}

func TestParseVersionInvalid(t *testing.T) {
	assert := assert.New(t)
	for _, s := range []string{
		"1.2",
		"1.2.a",
		"1.a.3",
		"a.2.3",
	} {
		v := Version{}
		err := parseVersion(s, &v)
		assert.Error(err)
	}
}

func TestParse(t *testing.T) {
	assert := assert.New(t)

	for _, test := range []struct {
		s     string
		v     Version
		strip string
	}{
		{
			"1.2.3-4-fcf2c8f",
			Version{Major: 1, Minor: 2, Patch: 3, Commits: 4, Hash: "fcf2c8f"},
			"",
		},
		{
			"0.0.0-0-",
			Version{},
			"",
		},
		{
			"1.2.3-rc1",
			Version{Major: 1, Minor: 2, Patch: 3, preRelease: "rc1"},
			"",
		},
		{
			"1.2.3-rc1-2-gd92f0b2",
			Version{Major: 1, Minor: 2, Patch: 3, preRelease: "rc1", Commits: 2, Hash: "gd92f0b2"},
			"",
		},
		{
			"3.2.1",
			Version{Major: 3, Minor: 2, Patch: 1},
			"",
		},
		{
			"v3.2.1",
			Version{Major: 3, Minor: 2, Patch: 1},
			"v",
		},
	} {
		v := Version{}
		err := parse(test.s, &v, test.strip)
		assert.NoError(err)
		assert.Equal(test.v, v)
	}
}

func TestParseInvalid(t *testing.T) {
	for _, s := range []string{
		"1.2.3-rc1-14-gd92f0b2-foo", // too many parts
		"1.2.3-rc1-foo-gd92f0b2",    // invalid commit count
	} {
		v := Version{}
		err := parse(s, &v)
		assert.Error(t, err)
	}
}

func TestString(t *testing.T) {
	assert := assert.New(t)
	for _, test := range []struct {
		v Version
		s string
	}{
		{
			Version{Major: 1, Minor: 2, Patch: 3, Commits: 10, Hash: "fcf2c8f"},
			"1.2.4-dev10+fcf2c8f",
		},
		{
			Version{Major: 0, Minor: 3, Patch: 1},
			"0.3.1",
		},
		{
			Version{Major: 1, Minor: 3, Patch: 0, preRelease: "rc3"},
			"1.3.0-rc3",
		},
		{
			Version{Major: 2, Minor: 5, Patch: 0, preRelease: "rc3", Commits: 3},
			"2.5.0-rc4.dev3",
		},
	} {
		assert.Equal(test.s, test.v.String())
	}

}

func TestFormat(t *testing.T) {
	assert := assert.New(t)
	v := Version{Major: 1, Minor: 2, Patch: 3, Commits: 10, Hash: "fcf2c8f"}
	for _, test := range []struct {
		f string
		p string
		s string
	}{
		{
			FullFormat,
			"",
			"1.2.4-dev10+fcf2c8f",
		},
		{
			NoMetaFormat,
			"",
			"1.2.4-dev10",
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
			"v1.2-dev10",
		},
	} {
		s, err := v.Format(test.f, test.p)
		assert.NoError(err)
		assert.Equal(test.s, s)
	}
}

func TestInvalidFormat(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	s, err := v.Format("q", "")
	assert.EqualError(t, err, "invalid format: q")
	assert.Equal(t, "", s)
}

func TestNextPreRelease(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("rc2", nextPreRelease("rc1"))
	assert.Equal("alpha1", nextPreRelease("alpha0"))
	assert.Equal("beta10", nextPreRelease("beta9"))
	assert.Equal("foo", nextPreRelease("foo"))
}

type gitFaker struct {
	s string
}

func (g gitFaker) Describe() string    { return g.s }
func (g gitFaker) CommitCount() string { return "" }

func TestDerive(t *testing.T) {
	for _, test := range []struct {
		s string
		v Version
	}{
		{
			"3.2.1-rc3-10-ge6c3c44",
			Version{Major: 3, Minor: 2, Patch: 1, preRelease: "rc3", Commits: 10, Hash: "ge6c3c44"},
		},
	} {
		git = gitFaker{test.s}
		Derive()
	}
}
