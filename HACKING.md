# Hacking on _Spyre_

## Source code layout

### Core code

- `cmd/spyre`: The main program
- `config`: Run-time configuration via command line switches.
- `log`: Human-readable output.
- `report`: Methods that produce output from scanners' match
  outputs. Output produced by the report subsystem will (hopefully) be
  processed by other programs.
- `platform`: platform-specific logic
- `platform/sys`: low-level syscall interface, magic constants, some
  of it generated code
- `scanner`, `scanner/*`: Scan modules
- `module_config`: Compile-time module configuration

### Utility packages

- `appendedzip`: Locate a ZIP archive appended to another file (such
  as the Spyre binary).
- `zipfs`: An [spf13/afero](https://github.com/spf13/afero) module
  that supports encrypted ZIP files.

### Scratch space

- `_3rdparty`: Build space for third-party dependencies not written in
  Go; see below.
- `_build`: Target directory for Spyre binaries
- `_gopath`: Local GOPATH

These files are prefixed with an underscore because the Go toolchain
won't look at those.

## Modules

All scanning functionality is implemented within modules. Currently,
three interface types are defined: `SystemScanner`, `FileScanner`,
`ProcScanner`, `EvtxScanner`. Including and initialization of those modules happens
through `import` statements in `module_config/*.go`.

`SystemScanner`s are run on program start and should consist of checks
that are not computationally or I/O intensive.

`FileScanner`s are run for every file, `ProcScanner`s are run for
every process id, but the individual scanner implementations can
choose to skip specific files or processes.

Refer to `scanner/yara` for a concrete `FileScanner` / `ProcScanner` / `EvtxScanner`
implementation and to `scanner/netscan` and `scanner/registry` for a
`SystemScanner` implementation.

## Third-party-dependencies

The rules in `3rdparty.mk` build C dependencies that are linked in
using the CGO foreign function interface. Artifacts are installed into
`_3rdparty/tgt/<triplet>`.
