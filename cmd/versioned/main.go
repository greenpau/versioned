// Copyright 2020 Paul Greenberg (greenpau@outlook.com)

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
	app = versioned.NewPackageManager("versioned")
	app.Description = "Simplified package metadata management for Go packages."
	app.Documentation = "https://github.com/greenpau/versioned/"
	app.SetVersion(appVersion, "")
	app.SetGitBranch(gitBranch, "")
	app.SetGitCommit(gitCommit, "")
	app.SetBuildUser(buildUser, "")
	app.SetBuildDate(buildDate, "")
}

func main() {
	var versionFile string
	var isShowVersion bool
	var isIncrementMajor bool
	var isIncrementMinor bool
	var isIncrementPatch bool
	var isInitialize bool
	var isSilent bool
	var factor uint64

	flag.StringVar(&versionFile, "file", "VERSION", "The file with version info")
	flag.BoolVar(&isInitialize, "init", false, "initialize a new version file")
	flag.BoolVar(&isIncrementMajor, "major", false, "increment major version")
	flag.BoolVar(&isIncrementMinor, "minor", false, "increment minor version")
	flag.BoolVar(&isIncrementPatch, "patch", false, "increment patch version")
	flag.Uint64Var(&factor, "factor", 1, "increase factor")
	flag.BoolVar(&isSilent, "silent", false, "silent execution")
	flag.BoolVar(&isShowVersion, "version", false, "version information")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n%s - %s\n\n", app.Name, app.Description)
		fmt.Fprintf(os.Stderr, "Usage: %s [arguments]\n\n", app.Name)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nDocumentation: %s\n\n", app.Documentation)
	}
	flag.Parse()
	if isShowVersion {
		fmt.Fprintf(os.Stdout, "%s\n", app.Banner())
		os.Exit(0)
	}

	if isInitialize {
		if version, err := versioned.NewVersionFromFile(versionFile); err == nil {
			fmt.Fprintf(os.Stderr, "version file already exists, version: %s\n", version)
			os.Exit(0)
		}
		version, _ := versioned.NewVersion("1.0.0")
		version.FileName = versionFile
		if err := version.UpdateFile(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize new version file: %s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	version, err := versioned.NewVersionFromFile(versionFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	oldVersion := *version

	if !isIncrementMajor && !isIncrementMinor && !isIncrementPatch {
		fmt.Fprintf(os.Stdout, "%s\n", version)
		os.Exit(0)
	}

	if isIncrementMajor {
		version.IncrementMajor(factor)
		if !isSilent {
			fmt.Fprintf(os.Stderr, "increased major version by %d, current version: %s\n",
				factor, version,
			)
		}
	}

	if isIncrementMinor {
		version.IncrementMinor(factor)
		if !isSilent {
			fmt.Fprintf(os.Stderr, "increased minor version by %d, current version: %s\n",
				factor, version,
			)
		}
	}

	if isIncrementPatch {
		version.IncrementPatch(factor)
		if !isSilent {
			fmt.Fprintf(os.Stderr, "increased patch version by %d, current version: %s\n",
				factor, version,
			)
		}
	}

	if err := version.UpdateFile(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	if !isSilent {
		fmt.Fprintf(os.Stderr, "updated version: %s, previous version: %s\n",
			version, &oldVersion,
		)
	}

	os.Exit(0)
}
