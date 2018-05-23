package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Version struct {
	Major   int
	Minor   int
	Patch   int
	Commits int
	Hash    string
}

func (v Version) String() string {
	var pre, meta string
	if v.Commits != 0 {
		v.Patch++
		pre = fmt.Sprintf("-dev%d", v.Commits)
	}
	if v.Hash != "" {
		meta = fmt.Sprintf("+%s", v.Hash)
	}
	return fmt.Sprintf("%d.%d.%d%s%s", v.Major, v.Minor, v.Patch, pre, meta)
}

func gitDescribe() string {
	cmd := exec.Command("git", "describe", "--tags")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("0.0.0-%s-", gitCommitCount())
	}
	return strings.TrimSpace(string(out))
}

func gitCommitCount() string {
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "0"
	}
	return strings.TrimSpace(string(out))
}

func parseVersion(s string, v *Version) error {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return fmt.Errorf("git version tag must contain 3 components: X.Y.Z: Got %s", s)
	}
	var err error
	v.Major, err = strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("Failed to parse major version: %v", err)
	}
	v.Minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("Failed to parse minor version: %v", err)
	}
	v.Patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("Failed to parse patch version: %v", err)
	}
	return nil
}

func parse(s string, v *Version) (err error) {
	var version string
	if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		if len(parts) != 3 {
			return fmt.Errorf("git version must be of format X.Y.Z-n-<hash>: Got %s", s)
		}
		version = parts[0]
		v.Commits, err = strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("Failed to parse commit count from %s: %v", s, err)
		}
		v.Hash = parts[2]
	} else {
		version = s
	}
	err = parseVersion(version, v)
	return err
}

func GitVersion() (Version, error) {
	v := Version{}
	s := gitDescribe()
	err := parse(s, &v)
	return v, err
}
