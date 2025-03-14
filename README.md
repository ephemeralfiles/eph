[![GitHub release](https://img.shields.io/github/release/ephemeralfiles/eph.svg)](https://github.com/ephemeralfiles/eph/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/ephemeralfiles/eph)](https://goreportcard.com/report/github.com/ephemeralfiles/eph)
![GitHub Downloads](https://img.shields.io/github/downloads/ephemeralfiles/eph/total)
[![GoDoc](https://godoc.org/github.com/ephemeralfiles/eph?status.svg)](https://godoc.org/github.com/ephemeralfiles/eph)
[![License](https://img.shields.io/github/license/ephemeralfiles/eph.svg)](LICENSE)

# eph

eph is the official client for [https://api.ephemeralfiles.com](https://api.ephemeralfiles.com).

## Getting started

### Install

Install from the release binary.

```bash
$ mkdir $HOME/.bin
$ curl -L https://github.com/ephemeralfiles/eph/releases/download/v0.2.0/eph_0.2.0_linux_amd64 --output $HOME/.bin/eph
$ chmod +x $HOME/.bin/eph
$ # add $HOME/.bin in your PATH
$ export PATH=$HOME/.bin:$PATH
$ # Add the new PATH to your .bashrc
$ echo 'export PATH=$HOME/.bin:$PATH' >> $HOME/.bashrc
```

### Configure

* First of all, you need to create an account on [https://api.ephemeralfiles.com](https://api.ephemeralfiles.com)
* Generate a token in your account
* Use the generated token to configure the client

```bash
$  eph config -t "generated-token"
Configuration saved
```

You should be able to use the client. Check with this command:

```bash
$ eph check
Token configuration:
  email: ********
  expiration Date: 2025-07-02 21:39:26
Box configuration:
  capacity: 5120 MB
  used: 0 MB
  remaining: 5120 MB
```
