# Example configuration for using _Spyre_

This directory contains a simple _Spyre_ configuration that references a few simple YARA rules (some of them from Florian Roth's [signature-base](https://github.com/Neo23x0/signature-base)). It is intended to serve as an example only.

While it is possible to use `spyre.yaml` and `*.yara` as-is for testing, it may be wise to put production rulesets into an encrypted ZIP file to hide signatures from overzealous antivirus products. Running `make` or `make spyre.zip` produces such an encrypted ZIP file, `spyre.zip` that can be placed into the same directory as the `spyre` or `spyre.exe` binary. (It has to have the smae basename as the _Spyre_ binary.)

``` console
$ make spyre.zip
zip spyre.zip -Pinfected spyre.yaml file-rules.yara proc-rules.yara common.yara
  adding: spyre.yaml (deflated 49%)
  adding: file-rules.yara (deflated 51%)
  adding: proc-rules.yara (deflated 45%)
  adding: common.yara (deflated 30%)
```

Running `make self-contained` requires the _Spyre_ binaries to be present in their original build directories and it will produce self-contained Windows (32bit x86) and Linux (x86_64) binaries that contain the configuration file and the YARA rules.

``` console
$ make self-contained
zip spyre.zip -Pinfected spyre.yaml file-rules.yara proc-rules.yara common.yara
  adding: spyre.yaml (deflated 49%)
  adding: file-rules.yara (deflated 51%)
  adding: proc-rules.yara (deflated 45%)
  adding: common.yara (deflated 30%)
cat ../_build/i686-w64-mingw32/spyre.exe spyre.zip > spyre-self-contained.exe.t
mv spyre-self-contained.exe.t spyre-self-contained.exe
cat ../_build/x86_64-linux-musl/spyre spyre.zip > spyre-self-contained.t
chmod 755 spyre-self-contained.t
mv spyre-self-contained.t spyre-self-contained
```

Note: Both the YARA rulesets `file-rules.yara` and `proc-rules.yara` use YARA's "include" feature and reference a `common.yara` file.
