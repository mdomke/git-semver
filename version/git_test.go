package version

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitDescribe(t *testing.T) {
	assert := assert.New(t)
	dir, err := ioutil.TempDir("", "git-semver")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Chdir(dir)
	assert.Equal("0.0.0-0-", git.Describe())
}
