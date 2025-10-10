[![GitHub release](https://img.shields.io/github/release/ephemeralfiles/eph.svg)](https://github.com/ephemeralfiles/eph/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/ephemeralfiles/eph)](https://goreportcard.com/report/github.com/ephemeralfiles/eph)
![coverage](https://raw.githubusercontent.com/wiki/ephemeralfiles/eph/coverage-badge.svg)
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

## Working with Organizations

Organizations allow teams to share storage and collaborate on files. The `eph org` command provides comprehensive organization management.

### Listing Organizations

View all organizations you have access to:

```bash
$ eph org list
ID                                    NAME    STORAGE         SUBSCRIPTION  ROLE
abc-123-def                          eph1    2.5GB / 10GB    Active        Admin
xyz-789-uvw                          eph2    500MB / 5GB     Active        Member
```

### Setting Default Organization

Set a default organization to avoid specifying it for every command:

```bash
$ eph org use eph1
Default organization set to 'eph1'

# Clear default organization
$ eph org use --clear
Default organization cleared
```

### Organization Information

View detailed organization information:

```bash
$ eph org info
Organization: eph1
ID: abc-123-def
Storage: 2.5GB / 10GB (25.0%)
Subscription: Active
Members: 5
Files: 147 (142 active, 5 expired)
```

### Uploading Files to Organizations

Upload files to your organization with optional tags:

```bash
# Upload to default organization
$ eph org up -i quarterly-report.pdf

# Upload with tags for better organization
$ eph org up -i invoice.pdf --tags "invoice,q1,2024"

# Override default organization
$ eph org up -i file.pdf --org eph2
```

### Listing Organization Files

List files in an organization with various filtering options:

```bash
# List all files
$ eph org ls

# Filter by tags
$ eph org ls --tags "invoice,2024"

# Show recent files
$ eph org ls --recent --limit 10

# Show expired files
$ eph org ls --expired

# JSON output
$ eph org ls --format json
```

### Downloading Organization Files

Download files from an organization:

```bash
# Download with original filename
$ eph org dl -i file-uuid-123

# Download with custom filename
$ eph org dl -i file-uuid-123 -o custom-name.pdf
```

### Deleting Organization Files

Delete files from an organization:

```bash
# With confirmation prompt
$ eph org rm -i file-uuid-123

# Skip confirmation
$ eph org rm -i file-uuid-123 --force
```

### Storage Management

Monitor organization storage usage:

```bash
$ eph org storage
Organization: eph1
Storage Limit:    10.00 GB
Used:             2.50 GB
Available:        7.50 GB
Usage:            25%
Status:           Normal
```

### Organization Statistics

View detailed organization statistics:

```bash
$ eph org stats
Organization ID:  abc-123-def
Files:            147
Active:           142
Expired:          5
Total Size:       2.5 GB
Members:          5
```

### Popular Tags

View most-used tags in an organization:

```bash
$ eph org tags --limit 10
TAG           COUNT
invoice       45
2024          38
report        25
contract      18
```

### Multi-Organization Workflow

Work with multiple organizations efficiently:

```bash
# Set default organization
$ eph org use eph1
$ eph org up -i file1.pdf

# Temporarily use different organization
$ eph org up -i file2.pdf --org eph2

# Check storage across organizations
$ eph org storage --org eph1
$ eph org storage --org eph2
```
