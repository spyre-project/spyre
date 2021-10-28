# Spyre

![Build Status](https://github.com/spyre-project/spyre/actions/workflows/go.yml/badge.svg)

**...a simple, self-contained modular host-based IOC scanner**

_Spyre_ is a simple host-based IOC scanner built around the
[YARA](https://github.com/VirusTotal/yara) pattern matching engine and
other scan modules. The main goal of this project is easy
operationalization of YARA rules and other indicators of compromise.

Users need to bring their own rule sets. The
[awesome-yara](https://github.com/InQuest/awesome-yara) repository gives
a good overview of free yara rule sets out there.

_Spyre_ is intended to be used as an investigation tool by incident
responders. It is **not** meant to evolve into any kind of endpoint
protection service.

## Getting Started

Using _Spyre_ is easy:

1. Add YARA signatures. In its default configuration, Spyre will read
   YARA rules for file and process scanning from `filescan.yar` and
   `procscan.yar`, respectively. The following options exist for
   providing rules files to _Spyre_ (and will be tried in this order):
    1. Add the rule files to a ZIP file and append that ZIP file to
	   the binary.
    2. Add the rule files to a ZIP file whose base name is identical
       to the scanner binary's base name, i.e. if the Spyre binary is
       called `spyre` or `spyre.exe`, use `spyre.zip`.
    3. Put the rule files and the scanner binary into the same
       directory.

   ZIP file contents may be encrypted using the password `infected`
   (AV industry standard) to prevent antivirus software from scanning
   the ruleset, classifying it as malicious content and preventing the
   scan.

   YARA rule files may contain `include` statements.
2. Deploy, run the scanner
3. Collect report and evidence

## Configuration

Run-time configuration is done via an optional file `spyre.yaml`.

If a ZIP file has been appended to the _Spyre_ binary, configuration
and other files such as YARA rules are only read from this ZIP file.
Otherwise, they are read from the directory into which the binary has
been placed.

See the [example-configuration/](example-config/) subdirectory for
an example.

### Global configuration

- `hostname` / command line switch `--set-hostname`: Explicitly set
  the hostname that will be used in the log file and in the report.
  This is usually not needed.
- `max-file-size` / command line switch `--max-file-size`: Maximum
  size for files to be scanned using expensive file scanning modules
  such as YARA. Default: 32MB
- `proc-ignore-names` / command line switch `--proc-ignore`: Names of
  processes that will not be scanned using process memory scanning
  modules.
- `paths` / command line switch `--path`: Paths to be scanned using
  file scanning modules. Default: `/` (Unix) or all fixed drives
  (Windows).
- `report` / comand line switch `--report`: Set one or more report
  targets. Default: `spyre_${hostname}_${time}.log` in the current
  working directory, using the plain format. A different output format
  can be specified by appending `,format=FORMAT`.

  The following formats are currently supported:
  - `plain`, the default, a simple human-readable text format
  - `tsjson`, a JSON document that can be imported into
    [Timesketch](https://github.com/google/timesketch)

  The `hostname` and `time` variables are only expanded in the target
  filename.

  **Note:** Configuration of report targets is likely to change in one
  of the next releases.
- `high-priority` / command line switch `--high-priority`: In its
  default configuration (with this setting disabled), _Spyre_
  instructs the OS scheduler to lower the priorities of CPU time and
  I/O operations, in order to avoid disruption of normal system
  operation.
- command line switch `--loglevel=LEVEL`: Set the log level. Valid:
  trace, debug, info, notice, warn, error, quiet.

### Module-specific configuration

There are currently three areas for which scanning modules can be
implemented: System-level checks, file scans, and process scans.

Listed below are the currently implemented modules and supported
configuration parameters.

- `system`
  - `eventobj` (Windows)
	- `iocs`
  - `registry` (Windows)
	- `iocs`
  - `winkernelobj` (Windows)
    - `iocs`
  - `findwindow` (Windows)
    - `iocs`
- `file`
  - `yara`
	- `rule-files`
	- `fail-on-warnings`
- `proc`
  - `yara`
	- `rule-files`
	- `fail-on-warnings`

Please refer to the example configuration file `example-spyre.yaml`
for hints on how to describe indicators of compromise for each module.

## Notes about YARA rules

YARA is configured with default settings, plus the following explicit
switches (cf. `3rdparty.mk`):

- `--disable-magic`
- `--disable-cuckoo`
- `--enable-dotnet`
- `--enable-macho`
- `--enable-dex`

For file scans, the following variables are defined:
- `filename`,
- `filepath`,
- `extension`,
- `filetype` (not currently populated while scanning)

For process scans, the variables `pid` and `executable` are defined.

The `spyre_collect_limit` metavariable can be used to limit the number
of writes collected from matching files or to inhibit collecting files
altogether. This can be useful to limit the size of evidence packages
and to avoid collecting sensitive information.

## Building

Spyre can be built for 32bit and 64bit Linux and Windows targets.

### Debian Buster (10.x) and later

On a Debian/buster system (or a chroot) in which the following packages
have been installed:

- make
- gcc
- gcc-multilib
- gcc-mingw-w64
- autoconf
- automake
- libtool
- pkg-config
- wget
- patch
- sed
- golang-_$VERSION_-go, e.g. golang-1.8-go. The Makefile will
  automatically select the newest version unless `GOROOT` has been
  set.
- git-core
- ca-certificates
- zip

This describes the build environment that is exercised regularly via
CI.

### Fedora 30 and later

The same build has also been successfully tried on Fedora 30 with the
following packages installed:

- make
- gcc
- mingw{32,64}-gcc
- mingw{32,64}-winpthreads-static
- autoconf
- automake
- libtool
- pkgconf-pkg-config
- wget
- patch
- sed
- golang
- git-core
- ca-certificates
- zip

Once everything has been installed, just type `make`. This should
download archives for _musl-libc_, _openssl_, _yara_, build those and
then build _spyre_.

The bare _spyre_ binaries are created in `_build/<triplet>/`.

Running `make release` creates a ZIP file that contains those binaries
for all supported architectures.

### Generating binaries compatible with ancient Windows XP, Windows Server 2003

Compatibility with these systems was removed with Go 1.11, so a Go
1.10 toolchain is required. Since Go 1.10 does not support Go modules,
third-party Go dependencies have to be vendored: Use a newer Go
version do this (just run `go vendor`) and set `GOROOT` to point to
the Go 1.10 toolchain before running `make`.

### MacOSX

Currently, cross-compiling is not supported.

- GCC from Xcode
- Build-dependencies from [Homebrew](https://brew.sh/):
  - gnu-make
  - autoconf
  - automake
  - libtool
  - pkg-config
  - wget
  - gpatch
  - gnu-sed
  - gnu-tar
  - go
  - git
  - ca-certificates
  - zip

The system-supplied `make` is too old because Apple decided to be
allergic to GPLv3. `gmake` from Homebrew works fine.

## Coding

See [HACKING.md](HACKING.md)

## Copyright

Copyright 2018-2020 DCSO Deutsche Cyber-Sicherheitsorganisation GmbH

Copyright 2020-2021 Spyre Project Authors (see: AUTHORS.txt)

## License

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

See the LICENSE file for the full license text.
