package version

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DefaultPrefix that is recognized and ignored by the parser.
const DefaultPrefix = "v"

// Predefined format strings to be used with the Format function.
const (
	FullFormat    = "x.y.z-p+m"
	NoMetaFormat  = "x.y.z-p"
	NoPreFormat   = "x.y.z"
	NoPatchFormat = "x.y"
	NoMinorFormat = "x"
)

// Target specifies a component of a semantic version that should be updated to.
type Target int

const (
	Devel Target = iota // updates to the next development version (e.g. updating dev.N)
	Patch               // updates to the next patch level
	Minor               // updates to the next minor version
	Major               // updates to the next major version
)

// The DefaultTarget when calculating the next version.
const DefaultTarget = Devel

func (t *Target) String() string {
	switch *t {
	case Devel:
		return "dev"
	case Patch:
		return "patch"
	case Minor:
		return "minor"
	case Major:
		return "major"
	default:
		panic(fmt.Errorf("unexpected target component %v", *t))
	}
}

func (t *Target) Set(value string) error {
	switch value {
	case "dev":
		*t = Devel
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

// Version holds the parsed components of git describe.
type Version struct {
	Prefix     string
	Major      int
	Minor      int
	Patch      int
	preRelease string
	Commits    int
	Meta       string
}

// Return 0 if both versions are equal
// Compares two versions. It returns
//   - 0 if both versions are equal
//   - some positive number if v > o
//   - some negative number if v < o
func (v Version) Compare(o *Version) int {
	if v.Major != o.Major {
		return v.Major - o.Major
	}
	if v.Minor != o.Minor {
		return v.Minor - o.Minor
	}
	if v.Patch != o.Patch {
		return v.Patch - o.Patch
	}
	if v.Commits != o.Commits {
		return v.Commits - o.Commits
	}
	if v.Meta != o.Meta {
		return strings.Compare(v.Meta, o.Meta)
	}
	// alphabetic order
	return strings.Compare(v.String(), o.String())
}

// BumpTo increases the version to the next patch/minor/major version. The version components with
// lower priority than the update target will be reset to zero.
//
//   - If target devel -> x.y.z-dev.(n+1)
//   - If target patch -> x.y.(z+1)
//   - If target minor -> x.(y+1).0
//   - If target major -> (x+1).0.0
func (v Version) BumpTo(target Target) Version {
	resetSuffix := func() {
		v.Commits = 0
		v.preRelease = ""
		v.Meta = ""
	}
	switch target {
	case Devel:
		if v.Commits > 0 && v.preRelease == "" {
			v.Patch++
		}
	case Patch:
		resetSuffix()
		v.Patch++
	case Minor:
		resetSuffix()
		v.Patch = 0
		v.Minor++
	case Major:
		resetSuffix()
		v.Patch = 0
		v.Minor = 0
		v.Major++
	}
	return v
}

// Format returns a string representation of the version including the parts
// defined in the format string. The format can have the following components:
//
//   - x -> major version
//   - y -> minor version
//   - z -> patch version
//   - p -> pre-release
//   - m -> metadata
//
// x, y and z are separated by a dot. p is seprated by a hyphen and m by a plus sign.
// E.g.: x.y.z-p+m or x.y .
func (v Version) Format(format string) (string, error) {
	re := regexp.MustCompile( // nolint: varnamelen
		`(?P<major>x)(?P<minor>\.y)?(?P<patch>\.z)?(?P<pre>-p)?(?P<meta>\+m)?`)

	matches := re.FindStringSubmatch(format)
	if matches == nil {
		return "", fmt.Errorf("invalid format: %s", format)
	}

	var (
		buf   buffer
		names = re.SubexpNames()
	)
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
			buf.AppendInt(v.Patch, '.')
		case "pre":
			buf.AppendString(v.PreRelease(), '-')
		case "meta":
			buf.AppendString(v.Meta, '+')
		}
	}
	return v.Prefix + string(buf), nil
}

func (v Version) String() string {
	result, err := v.Format(FullFormat)
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

// NewFromHead creates a new [Version] based on the given head revision, which can be created with
// [GitDescribe].
//
// The prefix is an arbitrary string that is prepended to the version number. The not SemVer
// commpliant but commonly used prefix v will be automatically detected.
func NewFromHead(head *RepoHead, prefix string) (Version, error) {
	result := Version{Commits: head.CommitsSinceTag}
	if prefix == "" {
		prefix = DefaultPrefix
	}
	if strings.HasPrefix(head.LastTag, prefix) {
		result.Prefix = prefix
	}
	version := strings.TrimPrefix(head.LastTag, result.Prefix)
	if strings.Contains(version, "+") {
		parts := strings.Split(version, "+")
		version = parts[0]
		result.Meta = parts[1]
	} else if head.CommitsSinceTag > 0 {
		result.Meta = head.Hash[:8]
	}
	if strings.Contains(version, "-") {
		parts := strings.SplitN(version, "-", 2)
		version = parts[0]
		result.preRelease = parts[1]
	}

	if version == "" {
		result.Major = 0
		result.Minor = 0
		result.Patch = 0
		return result, nil
	}

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return result, fmt.Errorf("git version tag must contain 3 components: X.Y.Z: Got %s", version)
	}
	var err error
	result.Major, err = strconv.Atoi(parts[0])
	if err != nil {
		return result, fmt.Errorf("failed to parse major version: %w", err)
	}
	result.Minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return result, fmt.Errorf("failed to parse minor version: %w", err)
	}
	result.Patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return result, fmt.Errorf("failed to parse patch version: %w", err)
	}
	return result, nil
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
