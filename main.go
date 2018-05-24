package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mdomke/git-semver/version"
)

var prefix = flag.String("prefix", "", "the version prefix (default: none)")

func main() {
	flag.Parse()
	v, err := version.Derive()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	v.Prefix = *prefix
	fmt.Println(v)
}
