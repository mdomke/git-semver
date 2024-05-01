package version

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DefaultPrefix that is recognized and ignored by the parser
const DefaultPrefix = "v"

// Predefined format strings to be used with the Format function
const (
	FullFormat    = "x.y.z-p+m"
	NoMetaFormat  = "x.y.z-p"
	NoPreFormat   = "x.y.z"
	NoPatchFormat = "x.y"
	NoMinorFormat = "x"
)

// Enum that specifies which version component should be bumped
type VersionComp int

const (
	Patch VersionComp = iota
	Minor
	Major
)

func (t *VersionComp) String() string {
	switch *t {
	case Patch:
		return "patch"
	case Minor:
		return "minor"
	case Major:
		return "major"
	default:
		panic(fmt.Errorf("unexpected TargetRevision value %v", *t))
	}
}

func (t *VersionComp) Set(value string) error {
	switch value {
	case "patch":
		*t = Patch
	case "minor":
		*t = Minor
	case "major":
		*t = Major
	default:
		return errors.New(`parse error`)
	}
	return nil
}

const DefaultVersionComp = Patch

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
	Prefix     string
	Major      int
	Minor      int
	Patch      int
	preRelease string
	Commits    int
	Meta       string
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
func (v Version) Format(format string, target VersionComp) (string, error) {
	re := regexp.MustCompile(
		`(?P<major>x)(?P<minor>\.y)?(?P<patch>\.z)?(?P<pre>-p)?(?P<meta>\+m)?`)

	matches := re.FindStringSubmatch(format)
	if matches == nil {
		return "", fmt.Errorf("invalid format: %s", format)
	}

	var (
		buf   buffer
		major = v.Major
		minor = v.Minor
		patch = v.Patch
	)

	if v.Commits > 0 && v.preRelease == "" {
		switch target {
		case Major:
			major++
			minor = 0
			patch = 0
		case Minor:
			minor++
			patch = 0
		case Patch:
			patch++
		}
	}

	names := re.SubexpNames()
	for i := 0; i < len(matches); i++ {
		if len(matches[i]) == 0 {
			continue
		}
		switch names[i] {
		case "major":
			buf.AppendInt(major, '.')
		case "minor":
			buf.AppendInt(minor, '.')
		case "patch":
			buf.AppendInt(patch, '.')
		case "pre":
			buf.AppendString(v.PreRelease(), '-')
		case "meta":
			buf.AppendString(v.Meta, '+')
		}
	}

	return v.Prefix + string(buf), nil
}

func (v Version) String() string {
	result, err := v.Format(FullFormat, DefaultVersionComp)
	if err != nil {
		return ""
	}
	return result
}

// PreRelease formats the pre-release version depending on the number n of commits since the
// last tag. If n is zero it returns the parsed pre-release version. If n is greater than zero
// it will append the string "dev.<n>" to the pre-release version.
func (v Version) PreRelease() string {
	if v.Commits == 0 {
		return v.preRelease
	}
	if v.preRelease == "" {
		return fmt.Sprintf("dev.%d", v.Commits)
	}
	return fmt.Sprintf("%s.dev.%d", v.preRelease, v.Commits)
}

func NewFromHead(head *RepoHead, prefix string) (Version, error) {
	v := Version{Commits: head.CommitsSinceTag}
	if prefix == "" {
		prefix = DefaultPrefix
	}
	if strings.HasPrefix(head.LastTag, prefix) {
		v.Prefix = prefix
	}
	version := strings.TrimPrefix(head.LastTag, v.Prefix)
	if strings.Contains(version, "+") {
		parts := strings.Split(version, "+")
		version = parts[0]
		v.Meta = parts[1]
	} else if head.CommitsSinceTag > 0 {
		v.Meta = head.Hash[:8]
	}
	if strings.Contains(version, "-") {
		parts := strings.SplitN(version, "-", 2)
		version = parts[0]
		v.preRelease = parts[1]
	}

	if version == "" {
		v.Major = 0
		v.Minor = 0
		v.Patch = 0
		return v, nil
	}

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return v, fmt.Errorf("git version tag must contain 3 components: X.Y.Z: Got %s", version)
	}
	var err error
	v.Major, err = strconv.Atoi(parts[0])
	if err != nil {
		return v, fmt.Errorf("failed to parse major version: %v", err)
	}
	v.Minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return v, fmt.Errorf("failed to parse minor version: %v", err)
	}
	v.Patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return v, fmt.Errorf("failed to parse patch version: %v", err)
	}
	return v, nil
}

// NewFromRepo calculates a semantic version for the head commit of the repo at path.
// If the latest commit is not tagged, the version will have a pre-release-suffix
// appended to it (e.g.: 1.2.3-dev.3+fcf2c8f). The suffix has the format dev.<n>+<hash>,
// whereas n is the number of commits since the last tag and hash is the commit hash
// of the latest commit. NewFromRepo will also increment the patch-level version component
// in case it detects that the current version is a pre-release.
// If the last tag has itself a pre-release-identifier and the last commit is not tagged,
// NewFromRepo will not increment the patch-level version.
// The prefix is an arbitrary string that is prepended to the version number. The not SemVer
// commpliant but commonly used prefix v will be automatically detected.
// The glob pattern can be used to limit the tags that are being considered in the calculation. The
// pattern allows the syntax described for filepath.Match.
func NewFromRepo(path, prefix, pattern string) (Version, error) {
	head, err := GitDescribe(path, WithMatchPattern(pattern))
	if err != nil {
		return Version{}, err
	}
	v, err := NewFromHead(head, prefix)
	return v, err
}
