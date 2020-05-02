# versioned

<a href="https://github.com/greenpau/versioned/actions/" target="_blank"><img src="https://github.com/greenpau/versioned/workflows/build/badge.svg?branch=master"></a>
<a href="https://pkg.go.dev/github.com/greenpau/versioned" target="_blank"><img src="https://img.shields.io/badge/godoc-reference-blue.svg"></a>
![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)

Simplified package metadata management for Go packages.

* [Manage package metadata as part of your code](#package-metadata)
* [Quickly increment major.minor.patch versions in `VERSION` file](#increment-majorminorpatch-versions)
* [Keep version in `VERSION` file in sync with package version metadata](#package-metadata)
* [Pass version, git branch/commit, build date/user metadata via `ldflags`](#package-metadata)

This repository contains:

* `github.com/greenpau/versioned`: Go package (library)
* `github.com/greenpau/versioned/cmd/versioned`: Command-line
  package metadata management utility

## Getting Started

First, install `versioned`:

```bash
go get -u github.com/greenpau/versioned/cmd/versioned
```

### Increment major.minor.patch Versions

Browse to a repository and initialize `VERSION` file with `versioned`:

```bash
versioned -init
```

Display current version of the repo:

```bash
versioned
```

Update patch version in `VERSION` file:

```bash
$ versioned -patch
increased patch version by 1, current version: 1.0.1
updated version: 1.0.1, previous version: 1.0.0
```

Do the same operation silently:

```bash
versioned -patch -silent
```

Update minor version in `VERSION` file:

```bash
versioned -minor
```

Update major version in `VERSION` file:

```bash
versioned -major
```

### Makefile Usage

Another way of using `versioned` is adding the following
`release` step in a `Makefile`:

```
GIT_BRANCH:=$(shell git rev-parse --abbrev-ref HEAD -- | head -1)

release:
        @echo "Making release"
        @if [ $(GIT_BRANCH) != "master" ]; then echo "cannot release to non-master branch $(GIT_BRANCH)" && false; fi
        @git diff-index --quiet HEAD -- || ( echo "git directory is dirty, commit changes first" && false )
        @versioned -patch
        @echo "Patched version"
        @git add VERSION
        @git commit -m "released v`cat VERSION | head -1`"
        @git tag -a v`cat VERSION | head -1` -m "v`cat VERSION | head -1`"
        @git push
        @git push --tags
```

## Package Metadata

For demonstration, please consider that a developer manages the following
code as part of `main.go`:

```golang
package main

import (
        "flag"
        "fmt"
        "github.com/greenpau/versioned"
        "os"
)

var (
        app        *versioned.PackageManager
        appVersion string
        gitBranch  string
        gitCommit  string
        buildUser  string
        buildDate  string
)

func init() {
    app = versioned.NewPackageManager("myapp")
    app.Description = "MyApp"
    app.Documentation = "https://github.com/me/myapp"
    app.SetVersion(appVersion, "1.0.0")
    app.SetGitBranch(gitBranch, "master")
    app.SetGitCommit(gitCommit, "v1.0.0-dirty")
    app.SetBuildUser(buildUser, "")
    app.SetBuildDate(buildDate, "")
}
```

In the `init()` function, the developer initializer `versioned` package
manager and sets default application name, description, and documentation.
These do not change at runtime.

Next, notice the `var` section, where there are a number of `string`
variable. At runtime, the values of the variables are not initialized.
The values are empty strings.

One way to initialize them is using `ldflags` at build time.
Here, the `APP_VERSION` variable is set to the value from
`VERSION` file.

Then, the variable is being user to set `main.appVersion`.
The `main` is the package in the code above and `appVersion`
is the variable that are being set at build time.

```bash
APP_VERSION:=$(shell cat VERSION | head -1)
GIT_COMMIT:=$(shell git describe --dirty --always)
GIT_BRANCH:=$(shell git rev-parse --abbrev-ref HEAD -- | head -1)
BUILD_USER:=$(shell whoami)
BUILD_DATE:=$(shell date +"%Y-%m-%d")

        @CGO_ENABLED=0 go build -o bin/myapp -v \
                -ldflags="-w -s \
                -X main.appVersion=$(APP_VERSION) \
                -X main.gitBranch=$(GIT_BRANCH) \
                -X main.gitCommit=$(GIT_COMMIT) \
                -X main.buildUser=$(BUILD_USER) \
                -X main.buildDate=$(BUILD_DATE)" \
                -gcflags="all=-trimpath=$(GOPATH)/src" \
                -asmflags="all=-trimpath $(GOPATH)/src" cmd/mypapp/*.go
```

The end result os that the following variables are being set with values.

```golang
        appVersion string
        gitBranch  string
        gitCommit  string
        buildUser  string
        buildDate  string
```

However, what happen when a user does not use `-ldflags`.

In that case, `versioned` sets a number of defaults. For example,
if `appVersion` is not being wtih `-ldflags`, then the default
`1.0.0` will be set.

```golang
    app.SetVersion(appVersion, "1.0.0")
    app.SetGitBranch(gitBranch, "master")
    app.SetGitCommit(gitCommit, "v1.0.0-dirty")
    app.SetBuildUser(buildUser, "")
    app.SetBuildDate(buildDate, "")
```

A developer, prior to releasing code, would put proper defaults.
This way when someone build the code, it would inherit a set of
default values for version, git, and build metadata.

Further, the `versioned` can be used to update the default values.

```bash
versioned -sync cmd/myapp/main.go
```
