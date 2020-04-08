package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Predefined format strings to be used with the Format function
const (
	FullFormat    = "x.y.z-p+m"
	NoMetaFormat  = "x.y.z-p"
	NoPreFormat   = "x.y.z"
	NoPatchFormat = "x.y"
	NoMinorFormat = "x"
)

type buffer []byte

func (b *buffer) AppendInt(i int, sep byte) {
	b.AppendString(strconv.FormatInt(int64(i), 10), sep)
}

func (b *buffer) AppendString(s string, sep byte) {
	if len(s) > 0 && len(*b) > 0 {
		*b = append(*b, sep)
	}
	*b = append(*b, s...)
}

// Version holds the parsed components of git describe
type Version struct {
	Major      int
	Minor      int
	Patch      int
	preRelease string
	Commits    int
	Hash       string
}

// Format returns a string representation of the version including the parts
// defined in the format string. The format can have the following components:
// * x -> major version
// * y -> minor version
// * z -> patch version
// * p -> pre-release
// * m -> metadata
// x, y and z are separated by a dot. p is seprated by a hyphen and m by a plus sing.
// E.g.: x.y.z-p+m or x.y
func (v Version) Format(format string, prefix string) (string, error) {
	re := regexp.MustCompile(
		`(?P<major>x)(?P<minor>\.y)?(?P<patch>\.z)?(?P<pre>-p)?(?P<meta>\+m)?`)

	matches := re.FindStringSubmatch(format)
	if matches == nil {
		return "", fmt.Errorf("invalid format: %s", format)
	}

	var buf buffer

	names := re.SubexpNames()
	for i := 0; i < len(matches); i++ {
		if len(matches[i]) == 0 {
			continue
		}
		switch names[i] {
		case "major":
			buf.AppendInt(v.Major, '.')
		case "minor":
			buf.AppendInt(v.Minor, '.')
		case "patch":
			patch := v.Patch
			if v.Commits > 0 && v.preRelease == "" {
				patch++
			}
			buf.AppendInt(patch, '.')
		case "pre":
			buf.AppendString(v.PreRelease(), '-')
		case "meta":
			buf.AppendString(v.Hash, '+')
		}
	}
	return prefix + string(buf), nil
}

func (v Version) String() string {
	result, err := v.Format(FullFormat, "")
	if err != nil {
		return ""
	}
	return result
}

// PreRelease formats the pre-release version depending on the number n of commits since the
// last tag. If n is zero it returns the parsed pre-release version. If n is greater than zero
// it will append the string "dev<n>" to the pre-release version.
func (v Version) PreRelease() string {
	if v.Commits == 0 {
		return v.preRelease
	}
	if v.preRelease == "" {
		return fmt.Sprintf("dev%d", v.Commits)
	}
	return fmt.Sprintf("%s.dev%d", nextPreRelease(v.preRelease), v.Commits)
}

func nextPreRelease(r string) string {
	re, err := regexp.Compile(`(.*?)(alpha|beta|rc)(\d+)`)
	if err != nil {
		return r
	}
	match := re.FindStringSubmatch(r)
	if match == nil {
		return r
	}
	prefix := match[1]
	preRelease := match[2]
	var n int
	n, err = strconv.Atoi(match[3])
	if err != nil {
		return r
	}
	n++
	return fmt.Sprintf("%s%s%d", prefix, preRelease, n)
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

func parse(s string, v *Version, prefix ...string) error {
	var version string
	for _, p := range prefix {
		s = strings.TrimPrefix(s, p)
	}
	if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")

		var commits string
		switch len(parts) {
		case 2:
			v.preRelease = parts[1]
		case 3:
			commits = parts[1]
			v.Hash = parts[2]
		case 4:
			v.preRelease = parts[1]
			commits = parts[2]
			v.Hash = parts[3]
		default:
			return fmt.Errorf("invalid git version must be of format X.Y.Z(-<pre>)?(-n-<hash>)?: Got %s", s)
		}

		version = parts[0]
		if commits != "" {
			var err error
			v.Commits, err = strconv.Atoi(commits)
			if err != nil {
				return fmt.Errorf("failed to parse commit count from %s: %v", s, err)
			}
		}
	} else {
		version = s
	}
	return parseVersion(version, v)
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
// Specify prefix in case the git-tag has a non SemVer commpliant prefix which should be
// stripped by the parser.
func Derive(prefix ...string) (Version, error) {
	v := Version{}
	s := git.Describe()
	err := parse(s, &v, prefix...)
	return v, err
}
