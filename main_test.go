package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/mdomke/git-semver/v6/version"
	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	for _, test := range []struct {
		args     []string
		cfg      *Config
		hasError bool
	}{
		{
			args: []string{"-prefix", "ver", "-format", "x.y.z-p", "-guard", "repo-root"},
			cfg: &Config{
				prefix:       "ver",
				format:       "x.y.z-p",
				guardRelease: true,
				args:         []string{"repo-root"},
			},
		},
		{
			args: []string{"-no-hash"},
			cfg:  &Config{excludeHash: true, args: []string{}},
		},
		{
			args: []string{"-no-meta"},
			cfg:  &Config{excludeMeta: true, args: []string{}},
		},
		{
			args: []string{"-no-pre"},
			cfg:  &Config{excludePreRelease: true, args: []string{}},
		},
		{
			args: []string{"-no-patch"},
			cfg:  &Config{excludePatch: true, args: []string{}},
		},
		{
			args: []string{"-no-minor"},
			cfg:  &Config{excludeMinor: true, args: []string{}},
		},
		{
			args: []string{"-no-prefix"},
			cfg:  &Config{excludePrefix: true, args: []string{}},
		},
		{
			args: []string{"-set-meta", "finleap"},
			cfg:  &Config{setMeta: "finleap", args: []string{}},
		},
		{
			args:     []string{"-help"},
			hasError: true,
		},
		{
			args:     []string{"-unknown"},
			hasError: true,
		},
	} {
		t.Run(strings.Join(test.args, " "), func(t *testing.T) {
			cfg, out, err := parseFlags("git-semver", test.args)
			if test.hasError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
				assert.NotEmpty(t, out)
			} else {
				test.cfg.stdout = os.Stdout
				test.cfg.stderr = os.Stderr
				assert.NoError(t, err)
				assert.Equal(t, test.cfg, cfg)
				assert.Empty(t, out)
			}
		})
	}
}

func TestSelectFormat(t *testing.T) {
	defaultVersion := version.Version{Major: 1, Minor: 2, Patch: 3, Commits: 4}
	for _, test := range []struct {
		desc   string
		cfg    Config
		v      version.Version
		format string
	}{
		{
			desc:   "Exclude minor format",
			cfg:    Config{excludeMinor: true},
			v:      defaultVersion,
			format: "x",
		},
		{
			desc:   "Exclude patch format",
			cfg:    Config{excludePatch: true},
			v:      defaultVersion,
			format: "x.y",
		},
		{
			desc:   "Exclude prerelease format",
			cfg:    Config{excludePreRelease: true},
			v:      defaultVersion,
			format: "x.y.z",
		},
		{
			desc:   "Exclude hash format",
			cfg:    Config{excludeHash: true},
			v:      defaultVersion,
			format: "x.y.z-p",
		},
		{
			desc:   "Prefere format over flags",
			cfg:    Config{format: "x.y.z", excludeMinor: true},
			v:      defaultVersion,
			format: "x.y.z",
		},
		{
			desc:   "Prefere longest strip",
			cfg:    Config{excludeHash: true, excludeMinor: true},
			v:      defaultVersion,
			format: "x",
		},
		{
			desc:   "Default to full format",
			cfg:    Config{},
			v:      defaultVersion,
			format: "x.y.z-p+m",
		},
		{
			desc:   "Use full format with pre-release version if -guard specified",
			cfg:    Config{guardRelease: true, excludePatch: true},
			v:      defaultVersion,
			format: "x.y.z-p+m",
		},
		{
			desc:   "Use full format with pre-release version if -guard and shorthand -format specified",
			cfg:    Config{guardRelease: true, format: "x.y.z"},
			v:      defaultVersion,
			format: "x.y.z-p+m",
		},
		{
			desc:   "Use custom format with pre-release version if -guard and suitable -format specified",
			cfg:    Config{guardRelease: true, format: "x.y.z-p"},
			v:      defaultVersion,
			format: "x.y.z-p",
		},
		{
			desc:   "Allow to strip meta/hash with -guard",
			cfg:    Config{guardRelease: true, excludeHash: true},
			v:      defaultVersion,
			format: "x.y.z-p",
		},
		{
			desc:   "Allow to strip patch with -guard if no pre-release version",
			cfg:    Config{guardRelease: true, excludePatch: true},
			v:      version.Version{Major: 3, Minor: 2, Patch: 1},
			format: "x.y",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			assert.Equal(t, test.format, selectFormat(&test.cfg, test.v))
		})
	}
}

func TestHandle(t *testing.T) {
	setup := func() (*Config, *bytes.Buffer) {
		var (
			cfg Config
			buf bytes.Buffer
		)
		cfg.stderr = &buf
		cfg.stdout = &buf
		return &cfg, &buf
	}
	t.Run("Path does not exist", func(t *testing.T) {
		cfg, buf := setup()
		assert.NoDirExists(t, "/sdf/")
		retval := handle(cfg, "/sdf/")
		assert.Equal(t, 1, retval)
		assert.Equal(t, "failed to open repo: repository does not exist", strings.TrimSpace(buf.String()))
	})
	t.Run("Path is not a git repo", func(t *testing.T) {
		cfg, buf := setup()
		assert.DirExists(t, "/tmp/")
		retval := handle(cfg, "/tmp/")
		assert.Equal(t, 1, retval)
		assert.Equal(t, "failed to open repo: repository does not exist", strings.TrimSpace(buf.String()))
	})
	t.Run("Meta can be set", func(t *testing.T) {
		cfg, buf := setup()
		cfg.setMeta = "finleap"
		retval := handle(cfg, "")
		assert.Equal(t, 0, retval)
		assert.True(t, strings.HasSuffix(strings.TrimSpace(buf.String()), cfg.setMeta))
	})
	t.Run("Prefix can be set", func(t *testing.T) {
		cfg, buf := setup()
		cfg.prefix = "v"
		retval := handle(cfg, "")
		fmt.Println(buf.String())
		assert.Equal(t, 0, retval)
		assert.True(t, strings.HasPrefix(strings.TrimSpace(buf.String()), cfg.prefix))
	})
	t.Run("Fails with invalid format", func(t *testing.T) {
		cfg, buf := setup()
		cfg.format = "a.b.c"
		retval := handle(cfg, "")
		assert.Equal(t, 1, retval)
		assert.Equal(t, fmt.Sprintf("invalid format: %s", cfg.format), strings.TrimSpace(buf.String()))
	})
}
