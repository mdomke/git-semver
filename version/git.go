package version

import (
	"fmt"
	"os/exec"
	"strings"
)

type gitter interface {
	Describe() string
	CommitCount() string
}

type gitCmd struct{}

func (g gitCmd) Describe() string {
	cmd := exec.Command("git", "describe", "--tags")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("0.0.0-%s-", g.CommitCount())
	}
	return strings.TrimSpace(string(out))
}

func (g gitCmd) CommitCount() string {
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "0"
	}
	return strings.TrimSpace(string(out))
}

var git gitter = gitCmd{}
