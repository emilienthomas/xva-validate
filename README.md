# xva-validate

[![Build Status](https://api.travis-ci.com/emilienthomas/xva-validate.svg?branch=master)](https://travis-ci.com/emilienthomas/xva-validate)

This is a command-line tool that performs verifications on a xva file to control integrity.
You can also use the Validate function inside of your own Go application or tool.

## How to build from sources

### Prerequisites

- [go binary](https://golang.org/doc/install)

### Build

```sh
go build
```

## Usage

All configurations are defined by command line arguments.
These arguments is printed by the flag -h

```sh
$ ./xva-validate -h
Usage of xva-validate:
  -verbose uint
        Verbosity level
  -version
        Print version and exit
  -xva string
        xva file (default "backup.xva")
```

By default, nothing is printed unless the xva file is invalid. If you want more details, you can change the value of the
-v parameter:
- 0 (default): only prints "xva file is invalid" and encountered error
- 1: also prints "xva file is valid"
- 2: prints all verifications.

```sh
$ ./xva-validate --xva exportedvm.xva
2019-02-22 18:10:14 read exportedvm.xva: Invalid descriptor
```

### Exit status
- 0: xva file is valid
- 1: xva file is invalid
- 2: an error as occurred during validation.

## Deployment

Just copy the binary !
