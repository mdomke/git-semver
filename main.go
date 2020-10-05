package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mdomke/git-semver/v5/version"
)

var prefix = flag.String("prefix", "", "prefix of version string e.g. v (default: none)")
var format = flag.String("format", "", "format string (e.g.: x.y.z-p+m)")
var excludeHash = flag.Bool("no-hash", false, "exclude commit hash (default: false)")
var excludeMeta = flag.Bool("no-meta", false, "exclude build metadata (default: false)")
var setMeta = flag.String("set-meta", "", "set build metadata (default: none)")
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
	case *excludeHash, *excludeMeta:
		format = version.NoMetaFormat
	default:
		format = version.FullFormat
	}
	return format
}

func main() {
	flag.Parse()
	v, err := version.Derive(*prefix)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *setMeta != "" {
		v.Meta = *setMeta
	}
	s, err := v.Format(selectFormat(), *prefix)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(s)
}
