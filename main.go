package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mdomke/git-semver/v6/version"
)

type Config struct {
	prefix            string
	format            string
	excludeHash       bool
	excludeMeta       bool
	setMeta           string
	excludePreRelease bool
	excludePatch      bool
	excludeMinor      bool
	guardRelease      bool
	args              []string
	stderr            io.Writer
	stdout            io.Writer
}

func parseFlags(progname string, args []string) (*Config, string, error) {
	var (
		buf bytes.Buffer
		cfg Config
	)

	flags := flag.NewFlagSet(progname, flag.ContinueOnError)
	flags.SetOutput(&buf)
	flags.StringVar(&cfg.prefix, "prefix", "", "prefix of version string e.g. v (default: none)")
	flags.StringVar(&cfg.format, "format", "", "format string (e.g.: x.y.z-p+m)")
	flags.BoolVar(&cfg.excludeHash, "no-hash", false, "exclude commit hash (default: false)")
	flags.BoolVar(&cfg.excludeMeta, "no-meta", false, "exclude build metadata (default: false)")
	flags.StringVar(&cfg.setMeta, "set-meta", "", "set build metadata (default: none)")
	flags.BoolVar(&cfg.excludePreRelease, "no-pre", false, "exclude pre-release version (default: false)")
	flags.BoolVar(&cfg.excludePatch, "no-patch", false, "exclude patch version (default: false)")
	flags.BoolVar(&cfg.excludeMinor, "no-minor", false, "exclude pre-release version (default: false)")
	flags.BoolVar(&cfg.guardRelease, "guard", false, "ignore shorthand options if version contains pre-release (default: false)")
	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage: %s [opts] [<repo>]\n\nOptions:\n", progname)
		flags.PrintDefaults()
	}

	err := flags.Parse(args)
	if err != nil {
		return nil, buf.String(), err
	}
	cfg.args = flags.Args()
	cfg.stderr = os.Stderr
	cfg.stdout = os.Stdout
	return &cfg, buf.String(), nil
}

func selectFormat(cfg *Config, v version.Version) string {
	var format string
	switch {
	case cfg.guardRelease && v.PreRelease() != "":
		if strings.Contains(cfg.format, version.NoMetaFormat) {
			format = cfg.format
		} else if cfg.excludeHash || cfg.excludeMeta {
			format = version.NoMetaFormat
		} else {
			format = version.FullFormat
		}
	case cfg.format != "":
		format = cfg.format
	case cfg.excludeMinor:
		format = version.NoMinorFormat
	case cfg.excludePatch:
		format = version.NoPatchFormat
	case cfg.excludePreRelease:
		format = version.NoPreFormat
	case cfg.excludeHash, cfg.excludeMeta:
		format = version.NoMetaFormat
	default:
		format = version.FullFormat
	}
	return format
}

func handle(cfg *Config, repoPath string) int {
	if repoPath == "" {
		var err error
		repoPath, err = os.Getwd()
		if err != nil {
			fmt.Fprintln(cfg.stderr, err)
			return 1
		}
	}
	v, err := version.NewFromRepo(repoPath)
	if err != nil {
		fmt.Fprintln(cfg.stderr, err)
		return 1
	}
	if cfg.setMeta != "" {
		v.Meta = cfg.setMeta
	}
	if cfg.prefix != "" {
		v.Prefix = cfg.prefix
	}
	s, err := v.Format(selectFormat(cfg, v))
	if err != nil {
		fmt.Fprintln(cfg.stderr, err)
		return 1
	}
	fmt.Fprintln(cfg.stdout, s)
	return 0
}

func main() {
	cfg, out, err := parseFlags(os.Args[0], os.Args[1:])
	if err != nil {
		fmt.Println(out)
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}

	var path string
	if len(cfg.args) > 0 {
		path = cfg.args[0]
	}
	os.Exit(handle(cfg, path))
}
