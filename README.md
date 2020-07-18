[![Build Status](https://travis-ci.org/spyre-project/spyre.svg?branch=master)](https://travis-ci.org/spyre-project/spyre)

# Spyre - a simple, self-contained modular host-based IOC scanner

_Spyre_ is a simple host-based IOC scanner built around the
[YARA](https://github.com/VirusTotal/yara) pattern matching engine and
other scan modules. The main goal of this project is easy
operationalization of YARA rules and other indicators of compromise.
Comprehensive rule sets are not included.

_Spyre_ is intended to be used as an investigation tool by incident
responders with an appropriate skill level. It is **not** meant to be
evolve into any kind of endpoint protection service.

## Overview

Using _Spyre_ is easy:

1. Add YARA signatures. Per default, filenames matching *.yr, *.yar,
   *.yara are recognized, see below how to change that. There are two
   options for doing this:
    - Put the rule files into the same directory as the binary
    - Add the rule files to ZIP file and append that file to the
      binary. Contents of the ZIP file may be encrypted using the
      password `infected` (AV industry standard) to provent antivirus
      software from mistaking parts of the ruleset as malicious
      content and preventing the scan.
2. Deploy, run the scanner
3. Collect report

## Configuration

Run-time options can be either passed via command line parameters or
via file that `params.txt`. Empty lines and lines starting with the
`#` character are ignored. Every line is interpreted as a single
command line argument.

If a ZIP file has been appended to the _Spyre_ binary, configuration
and other files such as YARA rules are only read from this ZIP file.
Otherwise, they are read from the directory into which the binary has
been placed.

Some options allow specifying a list of items. This can be done by
separating the items using a semicolon (`;`).

##### `--high-priority`

Normally (unless this switch is enabled), _Spyre_ instructs the OS
scheduler to lower the priorities of CPU time and I/O operations, in
order to avoid disruption of normal system operation.

##### `--set-hostname=NAME`

Explicitly set the hostname that will be used in the log file and in
the report. This is usually not needed.

##### `--loglevel=LEVEL`

Set the log level. Valid: trace, debug, info, notice, warn, error,
quiet.

##### `--report=SPEC`

Set one or more report targets, separated by a semicolon (`;`).
Default: `spyre.log` in the current working directory, using the plain
format.

A different output format can be specified by appending
`,format=FORMAT`. The following formats are currently supported:

- `plain`, the default, a simple human-readable text format
- `tsjson`, a JSON document that can be imported into
  [Timesketch](https://github.com/google/timesketch)

##### `--path=PATHLIST`

Set one or more specific filesystem paths to scan. Default: `/` (Unix)
or all fixed drives (Windows).

##### `--yara-rule-files=FILELIST`

Set explicit list of YARA rule files. Default: Use `*.yr`, `*.yar`,
*.yara` files from current working directory or appended ZIP file.

##### `--max-file-size=SIZE`

Set maximum size for files to be scanned using YARA. Default: 32MB

##### `--ioc-file=FILE`

## Notes about YARA rules

YARA is configured with default settings, plus the following explicit
switches (cf. `3rdparty.mk`):

- `--disable-magic`
- `--disable-cuckoo`
- `--enable-dotnet`

## Building

### Prerequisites

Spyre can be built on a Debian/stretch system (or a chroot) in
which the following packages have been installed:

- make
- gcc
- gcc-multilib
- gcc-mingw-w64
- autoconf
- automake
- libtool
- pkg-config
- wget
- sed
- golang-_$VERSION_-go, e.g. golang-1.8-go. The Makefile will
  automatically select the newest version unless `GOROOT` has been
  set.
- git-core
- ca-certificates
- zip

Also, go-dep from https://github.com/golang/dep is needed. `go install
github.com/golang/dep` should be sufficiant.

Once everything has been installed, just type `make`. This should
download archives for _musl-libc_, _openssl_, _yara_, build those and
then build _spyre_.

The bare _spyre_ binaries are created in `_build/<triplet>/`.

Running `make release` creates a ZIP file that contains those binaries
for all supported architectures.

### Extending

Starting with version 1.1.0, there is a module system that can be used
to add file and system scanners. File scanners, such as the YARA
module, act on every file. System scanners are run on program start
and usually consist of checks that should not be computationally or
I/O intensive.

File and system scanners need to be implemented as objects adhering to
the `FileScanner` and `SystemScanner` interfaces, respectively, and
have to be registered on startup. Packages containing those
implementations should be imported via `module_config/*.go`. See
`scanner/modules.go` for details and `scanner/yara`,
`scanner/eventobj`, `scanner/registry` for concrete implementations.

## Potentially interesting sub-packages

- _appendedzip_, code that tries to find a zip file appended to
  another file such as the main executable
- _zipfs_, which has been incorporated into
  [spf13/afero](https://github.com/spf13/afero), originated here and
  has since been extended with support for encrypted ZIP files.

## Copyright

Copyright 2018-2020 Deutsche Cyber-Sicherheitsorganisation GmbH

Copyright 2020      Spyre Project Authors (see: AUTHORS.txt)

## License

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

See the LICENSE file for the full license text.
