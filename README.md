[![Build Status](https://travis-ci.org/DCSO/spyre.svg?branch=master)](https://travis-ci.org/DCSO/spyre)

# Spyre - a simple, self-contained YARA-based file IOC scanner

_Spyre_ is a simple YARA scanner, the main goal is easy
operationalization of YARA rules. Comprehensive rule sets are not
included.

_Spyre_ is intended to be used as an investigation tool by incident
responders with an appropriate skill level. It is **not** meant to be
used as any kind of endpoint protection service.

## Overview

Using _Spyre_ is easy:

1. Add YARA signatures. Per default, filenames matching *.yr, *.yar,
   *.yara are recognized, see below how to change that. There are two
   options for doing this:
    - Put the rule files into the same directory as the binary
    - Add the rule files to ZIP file and append that file to the
      binary.
2. Deploy, run the scanner
3. Collect report

## Command line options

- `--path`: Set specific filesystem path to scan. Default: Unix: `/`,
  Windows: all fixed drives.
- `--report`: Specify one or more report targets. Default:
  `spyre.log` in the current working directory, using the plain format.
  A special output format can be specified by appending
  ;format=$FORMAT`. The following formats are currently supported:
    - `plain`, the default, a simple human-readable text format
    - `tsjson`, a JSON document that can be imported into
      [Timesketch](https://github.com/google/timesketch)
- `--max-file-size`: Set maximum size for files to be scanned.
  Default: 32MB.
- `--yara-rule-files`: Set explicit list of YARA rule files. Default:
  Use `*.yr`, `*.yar`, `*.yara` files from current working directory
  or appended ZIP file
- `--loglevel`: Set log level. Default: `info`.
- `--high-priority`: Normally (unless this switch is enabled), _Spyre_
  will instruct the OS scheduler to lower the priorities of CPU time
  and I/O operations, in order to avoid disruption of normal system
  operation.

## Notes about YARA rules

YARA is configured with default settings, plus the following explicit
switches (cf. `3rdparty.mk`):

- `--disable-magic`
- `--disable-cuckoo`
- `--enable-dotnet`

Please do not use the `import` statement as proper support for it is
lacking so far.

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

Also, go-dep from https://github.com/golang/dep is needed. `go install
github.com/golang/dep` should be sufficiant.

Once everything has been installed, just type `make`. This should
download archives for _musl-libc_, _openssl_, _yara_, build those and
then build _spyre_.

The _spyre_ bare binaries are created in `_build/<triplet>/`.

Running `make release` creates a ZIP file that contains binaries for
all supported architectures.

## Potentially interesting sub-packages

- _appendedzip_, code that tries to find a zip file appended to
  another file such as the main executable
- _zipfs_, a read-only filesystem provider for
  [spf13/afero](https://github.com/spf13/afero), see also
  [afero PR #146](https://github.com/spf13/afero/pull/146)

## Author

Hilko Bengen <hilko.bengen@dcso.de>

## Copyright

Copyright 2018 Deutsche Cyber-Sicherheitsorganisation GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
