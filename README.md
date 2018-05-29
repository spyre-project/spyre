# Spyre - a simple, self-contained YARA-based file IOC scanner

This program is intended to be used as an investigation tool by
incident response groups with an appropriate skill level. It is
**not** meant to be used as any kind of endpoint protection service.

## Overview

Get up and running in 3 easy steps:

1. Add YARA signatures to a ZIP archive
2. Append ZIP archive to the executable. Sign the executable if
   appropriate.
3. Deploy

## Command line options

- `--path`: Set specific filesystem path to scan. Default: Unix: `/`,
  Windows: all fixed drives.
- `--report`: Specify one or more report targets. Default:
  `spyre.log` in the current working directory, using the plain format.
  A special output format can be specified by appending
  `,format=<fmt>`. Currently, `plain` (default) and `tsjson` are supported.
- `--max-file-size`: Set maximum size for files to be scanned.
  Default: 32MB.
- `--loglevel`: Set log level. Default: `info`.

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
- go-dep from https://github.com/golang/dep

Once everything has been installed, just type `make`. This should
download archives for musl-libc, openssl, yara, build those and then
build Spyre.

The bare binaries are created in `_build/<triplet>/`.

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
