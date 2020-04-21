# versioned

<a href="https://github.com/greenpau/versioned/actions/" target="_blank"><img src="https://github.com/greenpau/versioned/workflows/build/badge.svg?branch=master"></a>
<a href="https://pkg.go.dev/github.com/greenpau/versioned" target="_blank"><img src="https://img.shields.io/badge/godoc-reference-blue.svg"></a>
![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)

Quickly increment major/minor/patch version in VERSION file.

## Getting Started

First, install `versioned`:

```bash
go get -u github.com/greenpau/versioned/cmd/versioned
```

### Basic Usage

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
