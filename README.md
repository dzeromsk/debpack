# debpack (tar2deb) - package debs the easy way

## Overview

`tar2deb` is a tool that takes a tar and outputs an deb. `debpack` is a golang library to create debs. Both are written in pure go, without using debuild (dpkg-buildpackage, fakeroot, lintian). API documentation for `debpack` can be found in [![GoDoc](https://godoc.org/github.com/dzeromsk/debpack?status.svg)](https://godoc.org/github.com/dzeromsk/debpack).

## Installation

```bash
$ go get -u github.com/dzeromsk/debpack/...
```

This will make the `tar2deb` tool available in `${GOPATH}/bin`, which by default means `~/go/bin`.

## Usage

`tar2deb` takes a `tar` file (from `stdin` or a specified filename), and outputs an `deb`.

```
Usage:
  tar2deb [OPTION] [FILE]
Options:
  -arch string
    	the deb architecture (default "all")
  -description string
    	the package description
  -file FILE
    	write deb to FILE instead of stdout
  -maintainer string
    	the package maintainer (default "unknown")
  -name string
    	the package name
  -version string
    	the package version
```

## Example

Prepare workspace
```bash
$ mkdir -p testdata/usr/bin/
$ cp ~/bin/example testdata/usr/bin/
```

Create archive
```bash
$ tar --mtime='1970-01-01 00:00:00Z' --owner=0 --group=0 --numeric-owner -Ctestdata -cf testdata/package.tar usr
$ tar tvf testdata/package.tar 
drwxr-xr-x 0/0               0 1970-01-01 01:00 usr/
drwxr-xr-x 0/0               0 1970-01-01 01:00 usr/bin/
-rwxr-xr-x 0/0         3677513 1970-01-01 01:00 usr/bin/example
```

Create deb package
```bash 
$ tar2deb -name package -version 1 -description "example package" -file testdata/package.deb testdata/package.tar 
```

Inspect and install
```bash 
$ dpkg --info testdata/package.deb 
 new Debian package, version 2.0.
 size 1881056 bytes: control archive=253 bytes.
     116 bytes,     6 lines      control              
      57 bytes,     1 lines      md5sums              
 Package: package
 Version: 1
 Architecture: all
 Installed-Size: 3591
 Maintainer: unknown
 Description: example package
$ dpkg -c testdata/package.deb 
drwxr-xr-x 0/0               0 1970-01-01 01:00 usr
drwxr-xr-x 0/0               0 1970-01-01 01:00 usr/bin
-rwxr-xr-x 0/0         3677513 1970-01-01 01:00 usr/bin/example
$ sudo dpkg -i testdata/package.deb 
Selecting previously unselected package package.
(Reading database ... 1234567 files and directories currently installed.)
Preparing to unpack testdata/package.deb ...
Unpacking package (1) ...
Setting up package (1) ...

```


## Features

 - Simple.
 - No config files.
 - You put files into the deb, so that deb/apt will install them on a host.
 - Does not build anything.
 - Does not try to auto-detect dependencies.
 - Does not try to magically deduce on which computer architecture you run.
 - Does not require any deb database or other state, and does not use the
   filesystem.

## Downsides

 - Many features are missing.
 - All of the artifacts are stored in memory, sometimes more than once.
 - Less backwards compatible than `debpack`.

## Philosophy

`tar2deb` is influenced heavily in style and interface from the https://github.com/google/rpmpack package.


