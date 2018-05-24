package version

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Prefix is the valid version prefix
const Prefix string = "v"

// Version holds the parsed components of git describe
type Version struct {
	Prefix     string
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Commits    int
	Hash       string
}

func (v Version) String() string {
	var suffix string
	if v.Commits != 0 {
		if v.PreRelease == "" {
			v.Patch++
			suffix = fmt.Sprintf("dev%d", v.Commits)
		} else {
			suffix = fmt.Sprintf("%s.dev%d", nextPreRelease(v.PreRelease), v.Commits)
		}
	} else {
		suffix = v.PreRelease
	}
	if v.Hash != "" {
		suffix += "+" + v.Hash
	}
	version := fmt.Sprintf("%s%d.%d.%d", v.Prefix, v.Major, v.Minor, v.Patch)
	if suffix != "" {
		version += "-" + suffix
	}
	return version
}

func nextPreRelease(r string) string {
	re, err := regexp.Compile(`(alpha|beta|rc)(\d+)`)
	if err != nil {
		return r
	}
	match := re.FindStringSubmatch(r)
	if match == nil {
		return r
	}
	prefix := match[1]
	var n int
	n, err = strconv.Atoi(match[2])
	if err != nil {
		return r
	}
	n++
	return fmt.Sprintf("%s%d", prefix, n)
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
	if strings.HasPrefix(s, Prefix) {
		v.Prefix = Prefix
		s = strings.TrimPrefix(s, Prefix)
	}
	if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")

		var commits string
		switch len(parts) {
		case 2:
			v.PreRelease = parts[1]
		case 3:
			commits = parts[1]
			v.Hash = parts[2]
		case 4:
			v.PreRelease = parts[1]
			commits = parts[2]
			v.Hash = parts[3]
		default:
			return fmt.Errorf("Invalid git version must be of format X.Y.Z(-<pre>)?(-n-<hash>)?: Got %s", s)
		}

		version = parts[0]
		if commits != "" {
			v.Commits, err = strconv.Atoi(commits)
			if err != nil {
				return fmt.Errorf("Failed to parse commit count from %s: %v", s, err)
			}
		}
	} else {
		version = s
	}
	err = parseVersion(version, v)
	return err
}

// Derive calculates a semantic version from the output of git describe.
// If the latest commit is not tagged, the version will have a pre-release-suffix
// appended to it (e.g.: 1.2.3-dev3+fcf2c8f). The suffix has the format dev<n>+<hash>,
// whereas n is the number of commits since the last tag and hash is the commit hash
// of the latest commit. Derive will also increment the patch-level version component
// in case it detects that the current version is a pre-release.
// If the last tag has itself a pre-release-suffix of the form (alpha|beta|rc)\d+ and the
// last commit is not tagged, Derive will increment the version of the pre-release
// instead of the patch-level version.
func Derive() (Version, error) {
	v := Version{}
	s := gitDescribe()
	err := parse(s, &v)
	return v, err
}
