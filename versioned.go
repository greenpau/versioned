// Copyright 2020 Paul Greenberg (greenpau@outlook.com)

package versioned

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Package stores metadata about a package.
type Package struct {
	Name          string
	Version       string
	Description   string
	Documentation string
	Git           gitMetadata
	Build         buildMetadata
}

// gitMetadata stores Git-related metadata.
type gitMetadata struct {
	Branch string
	Commit string
}

// buildInfo stores build-related metadata.
type buildMetadata struct {
	OperatingSystem string
	Architecture    string
	User            string
	Date            string
}

// Banner returns package
func (p *Package) Banner() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s %s", p.Name, p.Version))
	if p.Git.Branch != "" {
		sb.WriteString(fmt.Sprintf(", branch: %s", p.Git.Branch))
	}
	if p.Git.Commit != "" {
		sb.WriteString(fmt.Sprintf(", commit: %s", p.Git.Commit))
	}
	if p.Build.User != "" && p.Build.Date != "" {
		sb.WriteString(fmt.Sprintf(", build on %s by %s",
			p.Build.Date, p.Build.User,
		))
		if p.Build.OperatingSystem != "" && p.Build.Architecture != "" {
			sb.WriteString(
				fmt.Sprintf(" for %s/%s",
					p.Build.OperatingSystem, p.Build.Architecture,
				))
		}
		sb.WriteString(fmt.Sprintf(
			" (%s/%s %s)",
			runtime.GOOS,
			runtime.GOARCH,
			runtime.Version(),
		))
	}
	return sb.String()
}

// ShortBanner returns one-line information about a package.
func (p *Package) ShortBanner() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s %s", p.Name, p.Version))
	return sb.String()
}

// SetVersion sets Version attribute of Package.
func (p *Package) SetVersion(v, d string) {
	if v != "" {
		p.Version = v
		return
	}
	p.Version = d
}

// SetGitBranch sets Git.Branch attribute of Package.
func (p *Package) SetGitBranch(v, d string) {
	if v != "" {
		p.Git.Branch = v
		return
	}
	p.Git.Branch = d
}

// SetGitCommit sets Git.Commit attribute of Package.
func (p *Package) SetGitCommit(v, d string) {
	if v != "" {
		p.Git.Commit = v
		return
	}
	p.Git.Commit = d
}

// SetBuildUser sets Build.User attribute of Package.
func (p *Package) SetBuildUser(v, d string) {
	if v != "" {
		p.Build.User = v
		return
	}
	p.Build.User = d
}

// SetBuildDate sets Build.Date attribute of Package.
func (p *Package) SetBuildDate(v, d string) {
	if v != "" {
		p.Build.Date = v
		return
	}
	p.Build.Date = d
}

func (p *Package) String() string {
	return p.Banner()
}

// Version represents a software version.
// The version format is `major.minor.patch`.
type Version struct {
	Major    uint64
	Minor    uint64
	Patch    uint64
	FileName string
}

// NewVersion returns an instance of Version.
func NewVersion(s string) (*Version, error) {
	if s == "" {
		return nil, fmt.Errorf("empty string")
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("version must be in major.minor.patch format")
	}
	major, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse major version")
	}
	minor, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse minor version")
	}
	patch, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse patch version")
	}
	return &Version{
		Major:    major,
		Minor:    minor,
		Patch:    patch,
		FileName: "VERSION",
	}, nil
}

// String returns string representation of Version.
func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Bytes returns byte representation of Version string.
func (v *Version) Bytes() []byte {
	return []byte(v.String())
}

// IncrementMajor increments major version
func (v *Version) IncrementMajor(i uint64) {
	v.Major++
	v.Minor = 0
	v.Patch = 0
}

// IncrementMinor increments minor version
func (v *Version) IncrementMinor(i uint64) {
	v.Minor++
	v.Patch = 0
}

// IncrementPatch increments patch version
func (v *Version) IncrementPatch(i uint64) {
	v.Patch++
}

func readVersionFromFile(filePath string) (string, error) {
	var buffer bytes.Buffer
	fileHandle, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fileHandle.Close()

	scanner := bufio.NewScanner(fileHandle)
	for scanner.Scan() {
		line := scanner.Text()
		buffer.WriteString(strings.TrimSpace(line))
		break
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// NewVersionFromFile return Version instance by
// reading VERSION file in a current directory.
func NewVersionFromFile(versionFile string) (*Version, error) {
	if versionFile == "" {
		versionFile = "VERSION"
	}
	versionStr, err := readVersionFromFile(versionFile)
	if err != nil {
		return nil, fmt.Errorf("error reading %s file: %s", versionFile, err)
	}
	version, err := NewVersion(versionStr)
	if err != nil {
		return nil, err
	}
	version.FileName = versionFile
	return version, nil
}

// UpdateFile updates version information in the file associated
// with the version.
func (v *Version) UpdateFile() error {
	return ioutil.WriteFile(v.FileName, v.Bytes(), 0644)
}
