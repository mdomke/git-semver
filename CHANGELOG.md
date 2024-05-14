# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

### Added
## [6.9.0] - 2024-05-13
### Added
* New flag `-target` that can be used to select to which component the version will be bumped to.
  Possible values are `dev` (default), `patch`, `minor` and `major`. E.g.

        $ git describe
        v6.8.1-16-gcf8b124
        $ git-semver -target minor
        v6.9.0
        
* Added [devbox](https://www.jetpack.io/devbox/) configuration
* Introduced additional golangci-linters.

### Changed
* Updated `git-go` to v5.11.0

### Fixed
* An error was issued when invoked from subfolder of the repository whereas `git-describe` usually
  succeeds in such cases.

## [6.8.0] - 2023-09-22
### Added
* New flag `-match` that can be used to select only specific tags matching a glob pattern into the
  calculation (e.g. `git-describe -match "v1.2.*"`)

## [6.7.0] - 2023-05-08
### Added
* New flag `-no-prefix` to exclude the prefix (e.g. "v") from the printed version.

### Changed
* Use the `-prefix`-flag to allow parsing non-standard version prefixes (e.g. "ver")

## [6.6.0] - 2023-05-08
### Changed
* Upgrade Golang to 1.20 and update dependencies.

## [6.5.0] - 2022-06-24
### Changed
* Use [distroless/static:nonroot](https://github.com/GoogleContainerTools/distroless/tree/main/base) as 
  base image for the dockerized version of `git-semver` and build with Golang 1.18.

## [6.4.0] - 2022-06-24
### Changed
* If two annotated tags point to the same commit `git-semver` will now select the one that
  was created last. E.g.

        $ git tag -a -m "Release candidate" 1.1.0-rc.1 
        $ git-semver
        1.1.0-rc.1
        $ git tag -a -m "Final release" 1.1.0
        $ git-semver
        1.1.0

  Previously the behavior was undefined. Thanks [@igor-petrik-invitae](https://github.com/igor-petrik-invitae)
  for the issue report!

## [6.3.0] - 2022-05-20
### Added
* `git-semver` can now be installed with Homebrew

        $ brew install mdomke/git-semver/git-semver

## [6.2.0] - 2022-03-10
### Added
* Also build binaries for Windows

## [6.1.1] - 2021-09-16
### Fixed
* The ability to point `git-semver` to a different repository location was broken in 6.1.0
  and has been fixed by [@masonkatz](https://github.com/masonkatz).

## [6.1.0] - 2021-08-25
### Added
* A new flag `-guard` has been introduced to avoid accidentally overwriting production
  versions with a pre-release version. Consider that we have a tag `1.2.3-rc.1` and invoke
  `git-semver` with `-no-patch` we would get the version `1.2`, which would overwrite a previous
  version that was generated from the tag `1.2.2`. The `-guard` flag will enforce that the
  pre-release identifier is always included in the output regardless of the usage of shorthand
  options like `-no-patch`, `-no-pre`, etc.

## [6.0.3] - 2021-08-23
### Fixed
* The pre-release tag was parsed incorrectly if it included another dash. E.g.: `1.2.3-pre-release.1`
  This has been fixed by [@ckoehn](https://github.com/ckoehn).

## [6.0.2] - 2021-07-13
### Changed
* Upgrade Golang to 1.16 and `go-git` to 5.4.2
* Upgrade golangci-lint to `v1.41` and fix some linting errors.
* Switch to `golang:1.16-buster` as builder-image

## [6.0.1] - 2020-12-08
### Fixed
* The default log sort-order was finding the wrong tag. Switching to commiter time
  based sorting improves the situation, but eventually the concrete algorithm from
  git describe has to be reimplemented.

## [6.0.0] - 2020-10-28
### Changed
* Remove external dependency to git with a pure Go based implemenation.

## [5.0.0] - 2020-10-08
### Changed
* The SemVer compliance for "development versions" originating from a pre-release
  tag has been improved. Previously the pre-release version has been incremented
  before attaching the `dev.X` suffix. As pointed out by @choffmeister this results
  in a not compliant version sorting since

      1.2.3-rc.2.dev.1 > 1.2.3-rc.2

  because a larger set of pre-release fields has a higher precedence than a smaller
  set, if all of the preceding identifiers are equal [1]. A development version
  originating from the tag `1.2.3-rc.1` will now result in `1.2.3-rc.1.dev.1`.
* The dev-suffix added to a version that is derived from a tagged version is now
  formatted as `dev.X`. This will enforce proper sorting since dot-separted identifiers
  are compared individually and identifiers consisting only of digits will be compared
  numerically. So that

      dev1 < dev10 < dev2

  yields the wrong order whereas

      dev.1 < dev.2 < dev.10

  works as expected. Thanks [@choffmeister](https://github.com/choffmeister).

* The size of the docker image has been reduced to 25MB by [@choffmeister](https://github.com/choffmeister).
* The commonly used prefix "v" will be automatically detected by the parser now and the
  `-prefix` option is now only used to add a prefix that was not part of the tag before.
* Moved from Travis to GitHub Actions.

## [4.0.1] - 2020-06-05
### Added
* Publish binaries upon release [@schorzz](https://github.com/schorzz).

## [4.0.0] - 2020-04-10
### Changed
* Use semantic import path versioning.
* Allow tags to have a manually configured buildmeta section. E.g.: `v4.0.2-dev6+special`.


[1]: https://semver.org/#spec-item-11
[6.9.0]: https://github.com/mdomke/git-semver/compare/v6.8.1...v6.9.0
[6.8.0]: https://github.com/mdomke/git-semver/compare/v6.7.0...v6.8.0
[6.7.0]: https://github.com/mdomke/git-semver/compare/v6.6.0...v6.7.0
[6.6.0]: https://github.com/mdomke/git-semver/compare/v6.5.0...v6.6.0
[6.5.0]: https://github.com/mdomke/git-semver/compare/v6.4.0...v6.5.0
[6.4.0]: https://github.com/mdomke/git-semver/compare/v6.3.0...v6.4.0
[6.3.0]: https://github.com/mdomke/git-semver/compare/v6.2.0...v6.3.0
[6.2.0]: https://github.com/mdomke/git-semver/compare/v6.1.1...v6.2.0
[6.1.1]: https://github.com/mdomke/git-semver/compare/v6.1.0...v6.1.1
[6.1.0]: https://github.com/mdomke/git-semver/compare/v6.0.3...v6.1.0
[6.0.3]: https://github.com/mdomke/git-semver/compare/v6.0.2...v6.0.3
[6.0.2]: https://github.com/mdomke/git-semver/compare/v6.0.1...v6.0.2
[6.0.1]: https://github.com/mdomke/git-semver/compare/v6.0.0...v6.0.1
[6.0.0]: https://github.com/mdomke/git-semver/compare/v5.0.0...v6.0.0
[5.0.0]: https://github.com/mdomke/git-semver/compare/v4.0.1...v5.0.0
[4.0.1]: https://github.com/mdomke/git-semver/compare/v4.0.0...v4.0.1
[4.0.0]: https://github.com/mdomke/git-semver/compare/v3.1.1...v4.0.0
