package main

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
	v := Version{}
	err := parse("1.2.3-4-fcf2c8f", &v)
	assert.Nil(err)
	assert.Equal(1, v.Major)
	assert.Equal(2, v.Minor)
	assert.Equal(3, v.Patch)
	assert.Equal(4, v.Commits)
	assert.Equal("fcf2c8f", v.Hash)

	v = Version{}
	err = parse("0.0.0-0-", &v)
	assert.Nil(err)
	assert.Equal(0, v.Major)
	assert.Equal(0, v.Minor)
	assert.Equal(0, v.Patch)
	assert.Equal(0, v.Commits)
	assert.Equal("", v.Hash)
}

func TestVersionString(t *testing.T) {
	assert := assert.New(t)
	v := Version{
		Major:   1,
		Minor:   2,
		Patch:   3,
		Commits: 10,
		Hash:    "fcf2c8f",
	}
	assert.Equal("1.2.4-dev10+fcf2c8f", v.String())

	v = Version{
		Major: 0,
		Minor: 3,
		Patch: 1,
	}
	assert.Equal("0.3.1", v.String())
}
