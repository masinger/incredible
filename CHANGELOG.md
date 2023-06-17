# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
###  Added
### Fixed
### Changed
### Removed

<!-- Unreleased Template
###  Added
### Fixed
### Changed
### Removed
--->

<!-- Version Template
## vMAJOR.MINOR.PATCH
**Date**: YYYY-MM-DD

General description

COPY FROM UNRELEASED
--->

## v0.1.0
**Date**: 2023-06-17

This release adds support for LastPass and fixes a bugin within the Bitwarden provider that caused an entry's password to be used instead of the username and vice versa.

Additionally, caching has been added, changing the behavior when mapping a file multiple times.

###  Added
1. [provider] Add support for usernames and password fetched from LastPass.
### Fixed
1. [provider] A bug within the Bitwarden provider has been fixed, where the username has been returned instead of the password and vice versa.
### Changed
1. Loaded sources are now cached, when mapped to multiple environment variables.
2. When mapping a binary source multiple times, the mapped environment variables will point to the same file.

## v0.0.1
**Date**: 2023-06-13

This is the initial release of `incredible`.

###  Added
1. JSON-Schema describing the mapping file (`incredible.yml`)
2. Initial CLI implementation
3. Support for 
   1. Bitwarden
   2. Azure Key Vault