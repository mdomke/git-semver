# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [6.0.1] - 2020-12-08

### Fixed

* The default log sort-order was finding the wrong tag. Switching to commiter time
  based sorting improves the situation, but eventually the concrete algorithm from
  git describe has to be reimplemented.

## [6.0.0] - 2020-10-28

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
[6.0.0]: https://github.com/mdomke/git-semver/compare/v5.0.0...v6.0.0
[5.0.0]: https://github.com/mdomke/git-semver/compare/v4.0.1...v5.0.0
[4.0.1]: https://github.com/mdomke/git-semver/compare/v4.0.0...v4.0.1
[4.0.0]: https://github.com/mdomke/git-semver/compare/v3.1.1...v4.0.0
