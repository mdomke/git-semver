package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mdomke/git-semver/version"
)

var strip = flag.String("strip", "", "prefix to strip (default: none)")
var format = flag.String("format", "", "format string (e.g.: x.y.z-p+m)")
var excludeHash = flag.Bool("no-hash", false, "exclude commit hash (default: false)")
var excludePreRelease = flag.Bool("no-pre", false, "exclude pre-release version (default: false)")
var excludePatch = flag.Bool("no-patch", false, "exclude pre-release version (default: false)")
var excludeMinor = flag.Bool("no-minor", false, "exclude pre-release version (default: false)")

func selectFormat() string {
	if *format != "" {
		return *format
	}
	var format string
	switch {
	case *excludeMinor:
		format = version.NoMinorFormat
	case *excludePatch:
		format = version.NoPatchFormat
	case *excludePreRelease:
		format = version.NoPreFormat
	case *excludeHash:
		format = version.NoMetaFormat
	default:
		format = version.FullFormat
	}
	return format
}

func main() {
	flag.Parse()
	v, err := version.Derive(*strip)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	s, err := v.Format(selectFormat())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(s)
}
