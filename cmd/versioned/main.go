// Copyright 2020 Paul Greenberg (greenpau@outlook.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"path/filepath"
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
	app.SetVersion(appVersion, "1.0.28")
	app.SetGitBranch(gitBranch, "")
	app.SetGitCommit(gitCommit, "")
	app.SetBuildUser(buildUser, "")
	app.SetBuildDate(buildDate, "")
}

func main() {
	var versionedDir string
	var versionFile string
	var isShowVersion bool
	var isIncrementMajor bool
	var isIncrementMinor bool
	var isIncrementPatch bool
	var isInitialize bool
	var isSilent bool
	var isRelease bool
	var factor uint64
	var syncFilePath string
	var syncFileFormat string
	var isTocUpdate, isAddLicense, isStripLicense bool
	var targetFilePath string
	var licenseCopyrightHolder, licenseType string
	var licenseCopyrightYear uint64

	flag.StringVar(&versionedDir, "path", "./", "The path to data repository")
	flag.StringVar(&versionFile, "source", "VERSION", "The \"source of truth\" file with version info")
	flag.BoolVar(&isInitialize, "init", false, "initialize a new version file")
	flag.StringVar(&syncFilePath, "sync", "", "synchronize info from version file to `FILE`")
	flag.StringVar(&syncFileFormat, "format", "", "synchronize according to specific language, i.e. py, js, go, ts, etc.")
	flag.BoolVar(&isIncrementMajor, "major", false, "increment major version")
	flag.BoolVar(&isIncrementMinor, "minor", false, "increment minor version")
	flag.BoolVar(&isIncrementPatch, "patch", false, "increment patch version")

	flag.StringVar(&targetFilePath, "filepath", "", "target file path")

	// Markdown Table of Contents flags.
	flag.BoolVar(&isTocUpdate, "toc", false, "update table of contents")

	// License flags.
	flag.BoolVar(&isAddLicense, "addlicense", false, "add license header a file")
	flag.BoolVar(&isStripLicense, "striplicense", false, "strip license header from a file")
	flag.StringVar(&licenseType, "license", "apache", "license type")
	flag.StringVar(&licenseCopyrightHolder, "copyright", "", "license copyright holder")
	flag.Uint64Var(&licenseCopyrightYear, "year", 0, "copyright year")

	flag.BoolVar(&isRelease, "release", false, "omits commit version when syncing")
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
		if err := version.SetFile(versionFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize version file: %s\n", err)
			os.Exit(1)
		}
		if err := version.UpdateFile(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize new version file: %s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	switch {
	case isTocUpdate:
		if targetFilePath == "" {
			targetFilePath = "README.md"
		}
		toc := versioned.NewTableOfContents()
		toc.AddFilePath(targetFilePath)
		if err := versioned.UpdateToc(toc); err != nil {
			exitWithError(err)
		}
		os.Exit(0)
	case isAddLicense:
		lic := versioned.NewLicenseHeader()
		if err := lic.AddFilePath(targetFilePath); err != nil {
			exitWithError(err)
		}
		if err := lic.AddCopyrightHolder(licenseCopyrightHolder); err != nil {
			exitWithError(err)
		}
		if err := lic.AddYear(licenseCopyrightYear); err != nil {
			exitWithError(err)
		}
		if err := lic.AddLicenseType(licenseType); err != nil {
			exitWithError(err)
		}
		if err := versioned.AddLicense(lic); err != nil {
			exitWithError(err)
		}
		os.Exit(0)
	case isStripLicense:
		lic := versioned.NewLicenseHeader()
		if err := lic.AddFilePath(targetFilePath); err != nil {
			exitWithError(err)
		}
		if err := versioned.StripLicense(lic); err != nil {
			exitWithError(err)
		}
		os.Exit(0)
	}

	version, err := versioned.NewVersionFromFile(versionFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	oldVersion := *version

	if !isIncrementMajor && !isIncrementMinor && !isIncrementPatch && syncFilePath == "" && !isTocUpdate {
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

	if syncFilePath != "" {
		fi, err := os.Stat(syncFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		if !fi.Mode().IsRegular() {
			fmt.Fprintf(os.Stderr, "path %s is not a file\n", syncFilePath)
			os.Exit(1)
		}

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

		if isRelease {
			branch = ""
			commit = ""
		}

		pkg := versioned.NewPackageManager("")
		pkg.Version = version.String()
		pkg.Git.Branch = branch
		pkg.Git.Commit = commit

		ext := filepath.Ext(syncFilePath)
		fileDir, fileName := filepath.Split(syncFilePath)
		if ext == ".py" || syncFileFormat == "py" || syncFileFormat == "python" {
			if err := syncPythonFile(pkg, syncFilePath, fi); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
		if ext == ".go" {
			if err := syncGolangFile(pkg, syncFilePath, fi); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
		if ext == ".ts" || ext == ".js" {
			if err := syncJavascriptFile(pkg, syncFilePath, fi); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}

		fmt.Fprintf(os.Stderr, "file %s in %s directory has unsupported file extension %s\n", fileName, fileDir, ext)
		os.Exit(1)
	}

	os.Exit(0)
}

func syncJavascriptFile(pkg *versioned.PackageManager, fp string, fi os.FileInfo) error {
	var buffer bytes.Buffer
	fh, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer fh.Close()

	isVersionFound := false
	fileVersion := ""

	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Version: ") {
			isVersionFound = true
			v := strings.SplitN(line, ":", 2)[1]
			v = strings.TrimSpace(v)
			v = strings.Replace(v, ",", "", -1)
			v = strings.Replace(v, "'", "", -1)
			v = strings.Replace(v, "\"", "", -1)
			v = strings.TrimSpace(v)
			fileVersion = v
			if fileVersion != pkg.Version {
				buffer.WriteString(strings.ReplaceAll(line, fileVersion, pkg.Version) + "\n")
			} else {
				buffer.WriteString(line + "\n")
			}
			continue
		}
		buffer.WriteString(line + "\n")
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	fh.Close()
	ref := "Please see https://github.com/greenpau/versioned#nodejs-javascript-typescript"
	if !isVersionFound {
		return fmt.Errorf("version not found. %s", ref)
	}
	if pkg.Version != fileVersion {
		mode := fi.Mode()
		return ioutil.WriteFile(fp, buffer.Bytes(), mode.Perm())
	}
	return nil
}

// syncPythonFile inspects a Python file for __version__ module level
// dunder (see PEP 8) and, if necessary, updates the version to
// match the one found in VERSION file.
func syncPythonFile(pkg *versioned.PackageManager, fp string, fi os.FileInfo) error {
	var buffer bytes.Buffer
	fh, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer fh.Close()

	isVersionDunderExist := false
	fileVersion := ""
	versionDunder := "__version__"

	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		line := scanner.Text()
		line = line + "\n"
		if strings.HasPrefix(line, versionDunder) {
			isVersionDunderExist = true
			v := strings.SplitN(line, "=", 2)[1]
			v = strings.TrimSpace(v)
			v = strings.Replace(v, "'", "", -1)
			v = strings.Replace(v, "\"", "", -1)
			fileVersion = v
			if fileVersion != pkg.Version {
				buffer.WriteString("__version__ = '" + pkg.Version + "'\n")
			} else {
				buffer.WriteString(line)
			}
			continue
		}
		buffer.WriteString(line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	fh.Close()
	ref := "Please see https://github.com/greenpau/versioned#package-metadata"
	if !isVersionDunderExist {
		return fmt.Errorf("%s module level dunder not found. %s", versionDunder, ref)
	}
	if pkg.Version != fileVersion {
		mode := fi.Mode()
		return ioutil.WriteFile(fp, buffer.Bytes(), mode.Perm())
	}
	return nil
}

func syncGolangFile(pkg *versioned.PackageManager, fp string, fi os.FileInfo) error {
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

func exitWithError(err interface{}) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}
