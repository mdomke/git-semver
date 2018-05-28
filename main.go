package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mdomke/git-semver/version"
)

var prefix = flag.String("prefix", "", "the version prefix (default: none)")
var excludeHash = flag.Bool("no-hash", false, "exclude commit hash (default: false)")

func main() {
	flag.Parse()
	v, err := version.Derive()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	v.Prefix = *prefix
	if *excludeHash {
		v.Hash = ""
	}
	fmt.Println(v)
}
