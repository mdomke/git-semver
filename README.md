# Semantic Versioning with git tags

Software should be versioned in order to be able to identify a certain
feature set or to know when a specific bug has been fixed. It is a good
practice to use [Sementic Versioning](https://semver.org/) (SemVer) in
order to attach a meaning to a version number or the change thereof.

[git](https://git-scm.com/) allows you to conveniently reference a certain
state of your code through the usage of tags. Tags can have an arbitrary
identifier, so that it seems a naturaly choice of using them for versioning.

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

`git-semver` parses the output of `git describe` and derives a proper SemVer compliant
version from it. E.g.:

```
3.5.1-22-gbaf822dd5 -> 3.5.2-dev22+gbaf822dd5
4.2.0-rc3-5-fcf2c8f -> 4.2.0-rc4.dev5-fcf2c8f
```

It will attach a pre-release tag of the form `devN`, where `N` is the number of commits
since the last commit, and the commit hash as build-metadata. Additionally the patch level
component will be incremented in case of a pre-release-version. If the last tag itself
contains a pre-release-identifier of the form `(alpha|beta|rc)\d+`, that identifier will
be incremnted instead of the patch-level.

If you want to add a prefix to the derived version, you can use the `-prefix`-flag like so

```sh
$ git-semver -prefix v
v3.5.2-dev22+gbaf822dd5
```

## Installation

Currently `git-semver` can be installed with `go get`

```sh
$ go get gopkg.in/mdomke/git-semver.v1
```
