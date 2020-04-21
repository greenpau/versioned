// Copyright 2020 Paul Greenberg (greenpau@outlook.com)

package versioned

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

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
