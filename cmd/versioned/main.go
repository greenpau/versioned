// Copyright 2020 Paul Greenberg (greenpau@outlook.com)

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/greenpau/versioned"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
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
	app.SetVersion(appVersion, "1.0.18")
	app.SetGitBranch(gitBranch, "master")
	app.SetGitCommit(gitCommit, "v1.0.17-1-g387a001")
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
	var syncFile string

	flag.StringVar(&versionFile, "file", "VERSION", "The file with version info")
	flag.BoolVar(&isInitialize, "init", false, "initialize a new version file")
	flag.StringVar(&syncFile, "sync", "", "synchronize info from version file to `FILE`")
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

	if !isIncrementMajor && !isIncrementMinor && !isIncrementPatch && syncFile == "" {
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

	if isIncrementMajor || isIncrementMinor || isIncrementPatch {
		if err := version.UpdateFile(); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		if !isSilent {
			fmt.Fprintf(os.Stderr, "updated version: %s, previous version: %s\n",
				version, &oldVersion,
			)
		}
	}

	if syncFile != "" {
		commit, err := executeShell([]string{"git", "describe", "--always"})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		branch, err := executeShell([]string{"git", "rev-parse", "--abbrev-ref", "HEAD", "--"})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		pkg := versioned.NewPackageManager("")
		pkg.Version = version.String()
		pkg.Git.Branch = branch
		pkg.Git.Commit = commit

		if err := parseFile(pkg, syncFile); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}

	os.Exit(0)
}

func parseFile(pkg *versioned.PackageManager, fp string) error {
	var buffer bytes.Buffer
	fh, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer fh.Close()

	isPackageIncluded := false
	isPackageInitialized := false
	foundVersionMatch := false
	rewrite := false
	pkgName := "github.com/greenpau/versioned"
	isInsideInit := 0

	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		line := scanner.Text()
		line = line + "\n"
		//fmt.Fprintf(os.Stderr, "%s", line)
		if strings.Contains(line, pkgName) {
			isPackageIncluded = true
			buffer.WriteString(line)
			continue
		}
		if strings.Contains(line, "func init() {") {
			buffer.WriteString(line)
			isInsideInit++
			continue
		}

		if isInsideInit == 1 {
			if strings.HasPrefix(line, "}") {
				buffer.WriteString(line)
				isInsideInit++
				continue
			}
		}

		if isInsideInit == 1 {
			if strings.Contains(line, "versioned.NewPackageManager") {
				isPackageInitialized = true
				buffer.WriteString(line)
				continue
			}

			if isPackageInitialized {
				verRegex := regexp.MustCompile("\\s*(\\S.*)\\.Set(\\S+)\\(\\S+, \"(.*)\"")

				if m := verRegex.FindStringSubmatch(line); len(m) > 0 {
					isRepl := false
					repl := ""
					switch v := m[2]; v {
					case "Version":
						foundVersionMatch = true
						if m[3] != pkg.Version {
							repl = pkg.Version
							isRepl = true
						}
					case "GitBranch":
						if m[3] != pkg.Git.Branch {
							repl = pkg.Git.Branch
							isRepl = true
						}
					case "GitCommit":
						if m[3] != pkg.Git.Commit {
							repl = pkg.Git.Commit
							isRepl = true
						}
					default:
						// do nothing
					}

					if isRepl {
						line = strings.Replace(line, "\""+m[3]+"\"", "\""+repl+"\"", -1)
						rewrite = true
					}
				}
			}
		}
		buffer.WriteString(line)

		//buffer.WriteString(strings.TrimSpace(line))
	}

	// fmt.Fprintf(os.Stderr, "%s\n", buffer.String())
	if err := scanner.Err(); err != nil {
		return err
	}

	fh.Close()

	ref := "Please see https://github.com/greenpau/versioned#package-metadata"

	if !isPackageIncluded {
		return fmt.Errorf("package %s not found", pkgName)
	}

	if !isPackageInitialized {
		return fmt.Errorf("package %s is not initialized. %s", pkgName, ref)
	}

	if !foundVersionMatch {
		return fmt.Errorf("package version not found. %s", ref)
	}

	if rewrite {
		return ioutil.WriteFile(fp, buffer.Bytes(), 0644)
	}
	return nil
}

func executeShell(args []string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Error executing %s: %s", args, err)
	}
	return strings.Split(stdout.String(), "\n")[0], nil
}
