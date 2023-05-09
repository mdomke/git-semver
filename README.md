[![GitHub Actions](https://img.shields.io/github/actions/workflow/status/mdomke/git-semver/lint-and-test.yaml?branch=master)](https://github.com/mdomke/git-semver/actions?query=workflow%3Alint-and-test)
[![Codecov](https://codecov.io/gh/mdomke/git-semver/branch/master/graph/badge.svg)](https://codecov.io/gh/mdomke/git-semver)
![License](https://img.shields.io/github/license/mdomke/git-semver.svg)
![Tag](https://img.shields.io/github/tag/mdomke/git-semver.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/mdomke/git-semver)](https://goreportcard.com/report/github.com/mdomke/git-semver)

# Semantic Versioning with git tags

Software should be versioned in order to be able to identify a certain
feature set or to know when a specific bug has been fixed. It is a good
practice to use [Semantic Versioning](https://semver.org/) (SemVer) in
order to attach a meaning to a version number or the change thereof.

[git](https://git-scm.com/) allows you to conveniently reference a certain
state of your code through the usage of tags. Tags can have an arbitrary
identifier, so that it seems a natural choice to use them for versioning.

---

* [Version tags](#version-tags)
* [Usage](#git-semver)
   * [Formatting](#formatting)
   * [Command line options](#command-line-options)
   * [Release safeguard](#release-safeguard)
* [Installation](#installation)
* [Docker usage](#docker-usage)


## Version tags

A semantic version consists of three dot-separated parts `<major>.<minor>.<patch>`
and this should be the name that you give to a tag. Optionally you can prepend
the letter `v` if your language specific tooling requires it. It is also possible
to attach a pre-release identifier to a version e.g. for a release candidate. This
identifier is separated with hyphen from the core version component. A valid version
tag would be, e.g. `1.2.3`, `v2.3.0`, `1.1.0-rc3`.

```sh
$ git tag v2.0.0-rc1
```

So for a tagged commit we would know which version to assign to our software, but
which version should we use for not tagged commits? We can use `git describe` to
get a unique identifier based on the last tagged commit.

```sh
$ git describe --tags
3.5.1-22-gbaf822dd5
```

This is the 22nd commit after the tag `3.5.1` with the abbreviated commit hash `gbaf822dd5`.
Sadly this identifier has two drawbacks.

1. It's not compliant to SemVer, because there are multiple hyphens after the core version.
   See the [BNF specifiction](https://github.com/semver/semver/blob/master/semver.md#backusnaur-form-grammar-for-valid-semver-versions)

2. It doesn't allow proper sorting of versions, because the pre-release identifier would
   make the version smaller than the tagged version, even though it has several commits build
   on top of that version.

## git-semver

`git-semver` collects information about the head commit of a repo similar to how
`git describe` would do it and derives a SemVer compliant version from it. E.g.:

| `git describe`          | `git-semver`                |
| ---                     | ---                         |
| `3.5.1-22-gbaf822d`     | `3.5.2-dev.22+baf822dd`     |
| `4.2.0-rc.3-5-gfcf2c8f` | `4.2.0-rc.3.dev.5+fcf2c8fd` |
| `1.0.1`                 | `1.0.1`                     |

It will attach a pre-release tag of the form `dev.N`, where `N` is the number of commits
since the last commit, and the commit hash as build-metadata. Additionally the patch level
component will be incremented in case of a pre-release-version. If the last tag itself
contains a pre-release-identifier the `dev.N` suffix will be appended but all other parts
will be left untouched. This complies with the [precedence rules](https://semver.org/#spec-item-11)
defined in the SemVer spec. So that

```
0.9.9 < 1.0.0-rc.1 < 1.0.0-rc1.dev.3+fcf2c8fd < 1.0.0-rc.2 < 1.0.0
```

### Formatting

The output of `git-semver` can be controlled with the `-format` option or one of it shorthand
companions as described [here](#command-line-options). The format string can include the following
characters

| Format char | Description         |
| ---         | ---                 |
| `x`         | Major version       |
| `y`         | Minor version       |
| `z`         | Patch version       |
| `p`         | Pre-release version |
| `m`         | Metadata            |

The format chars `x`, `y` and `z` are separted with a dot, `p` with a hyphen and `m` with a
plus character. A valid format string is e.g.: `x.y+m`

### Command line options

The output and parsing of `git-semver` can be controlled with the following options.

| Name                  | Description                                              |
| ---                   | ---                                                      |
| `-format`             | Format string as described [here](#formatting)           |
| `-no-minor`           | Exclude minor version and all following components       |
| `-no-patch`           | Exclude patch version and all following components       |
| `-no-pre`             | Exclude pre-release version and all following components |
| `-no-meta`/`-no-hash` | Exclude build metadata                                   |
| `-prefix`             | Prefix string for version e.g.: v                        |
| `-set-meta`           | Set buildmeta to this value                              |
| `-guard`              | Ignore shorthand formats for pre-release versions        |


#### Examples

```sh
$ git-semver
3.5.2-dev.22+8eaec5d3

# Exclude build metadata
$ git-semver -no-meta
3.5.2-dev.22

# Only major and minor version
$ git-semver -no-patch
3.5

$ git-semver -prefix v -no-hash
v3.5.2

$ git-semver -set-meta custom
3.5.2+custom
```

### Release safeguard

If you use `git-semver` to automatically derive versions for your application and you
want to provide convenient shorthand versions (e.g. `1.2`), so that it is easier to follow
non-breaking updates, you might run into the problem that a pre-release version accidentally 
overwrites a production version. This is because

```sh
# tag of HEAD commit: 1.2.2
$ git-semver -no-patch
1.2

# tag of HEAD commit: 1.2.3-dev.1"
$ git-semver -no-patch
1.2
```

result in the same shorthand version. To mitigate this problem you can use the `-guard` option
that will ignore any output format that doesn't contain the pre-release identifier if the current
version is a pre-release version. E.g.

```sh
# tag of HEAD commit: 1.2.3-dev.1"
$ git-semver -guard -no-patch
1.2.3-dev.1+8eaec5d3
```

### Caveats

If you create multiple annotated tags on the same commit (e.g. you want to promote a release candidate
to be the final release without adding any further commits), `git-semver` will pick the tag that was
created last, which is usually what you want. E.g.

```sh
$ git tag -a -m "Release candidate" 1.1.0-rc.1 
$ git-semver
1.1.0-rc.1
$ git tag -a -m "Final release" 1.1.0
$ git-semver
1.1.0
```

## Installation

Currently `git-semver` can be installed with `go install`

```sh
$ go install github.com/mdomke/git-semver/v6@latest
```

There is also a [Homebrew](https://brew.sh/) formula that can be installed with

```sh
$ brew install mdomke/git-semver/git-semver
```

## Docker usage

You can also use `git-semver` as a docker-container. Images are available from [DockerHub][1] and
[GitHub Container Registry][2]

```sh
docker run --rm -v `pwd`:/git-semver mdomke/git-semver
```
or
```sh
docker run --rm -v `pwd`:/git-semver ghcr.io/mdomke/git-semver
```

[1]: https://hub.docker.com/r/mdomke/git-semver
[2]: https://github.com/mdomke/git-semver/pkgs/container/git-semver
