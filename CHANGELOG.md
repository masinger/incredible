# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
###  Added
1. `--provider-timeout` (default `2m`) limits how long non-interactive provider CLI calls (`bw`, `az`) may take. Logins and unlocks waiting for user input aren't limited.
### Fixed
1. Provider CLI calls ignored their context, so a hanging provider blocked incredible forever. Pressing Ctrl-C while the environment is loading now stops the running provider call.
### Changed
1. Go 1.26 or newer is required to build incredible; all dependencies have been updated.
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