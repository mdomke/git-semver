package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVersion(t *testing.T) {
	assert := assert.New(t)
	v := Version{}
	err := parseVersion("1.2.3", &v)
	assert.Nil(err)
	assert.Equal(1, v.Major)
	assert.Equal(2, v.Minor)
	assert.Equal(3, v.Patch)
}

func TestParseInvalidVersion(t *testing.T) {
	assert := assert.New(t)
	v := Version{}
	err := parseVersion("1.2", &v)
	assert.NotNil(err)

	err = parseVersion("1.2.a", &v)
	assert.NotNil(err)
}

func TestParse(t *testing.T) {
	assert := assert.New(t)

	for _, test := range []struct {
		s string
		v Version
	}{
		{
			"1.2.3-4-fcf2c8f",
			Version{Major: 1, Minor: 2, Patch: 3, Commits: 4, Hash: "fcf2c8f"},
		},
		{
			"0.0.0-0-",
			Version{},
		},
		{
			"1.2.3-rc1",
			Version{Major: 1, Minor: 2, Patch: 3, preRelease: "rc1"},
		},
		{
			"1.2.3-rc1-2-gd92f0b2",
			Version{Major: 1, Minor: 2, Patch: 3, preRelease: "rc1", Commits: 2, Hash: "gd92f0b2"},
		},
	} {
		v := Version{}
		err := parse(test.s, &v)
		assert.Nil(err)
		assert.Equal(test.v, v)
	}
}

func TestVersionString(t *testing.T) {
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

func TestVersionFormat(t *testing.T) {
	assert := assert.New(t)
	v := Version{Major: 1, Minor: 2, Patch: 3, Commits: 10, Hash: "fcf2c8f"}
	for _, test := range []struct {
		f string
		s string
	}{
		{
			FullFormat,
			"1.2.4-dev10+fcf2c8f",
		},
		{
			NoMetaFormat,
			"1.2.4-dev10",
		},
		{
			NoPreFormat,
			"1.2.4",
		},
		{
			NoPatchFormat,
			"1.2",
		},
		{
			NoMinorFormat,
			"1",
		},
		{
			"x.y-p",
			"1.2-dev10",
		},
	} {
		s, err := v.Format(test.f)
		assert.Nil(err)
		assert.Equal(test.s, s)
	}
}

func TestNextPreRelease(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("rc2", nextPreRelease("rc1"))
	assert.Equal("alpha1", nextPreRelease("alpha0"))
	assert.Equal("beta10", nextPreRelease("beta9"))
	assert.Equal("foo", nextPreRelease("foo"))
}
