// Copyright 2020 Paul Greenberg (greenpau@outlook.com)

package versioned

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type action struct {
	operation   string
	incrementBy uint64
}

func TestVersionedCalculus(t *testing.T) {
	testFailed := 0

	for i, test := range []struct {
		description string
		input       string
		output      string
		actions     []action
		shouldFail  bool // Whether test should result in a failure
		shouldErr   bool // Whether parsing of a response should result in error
		errMessage  string
	}{
		{
			input:  "1.0.0",
			output: "2.0.0",
			actions: []action{
				action{
					operation:   "increment_major",
					incrementBy: 1,
				},
			},
			shouldFail: false,
			shouldErr:  false,
			errMessage: "",
		},
		{
			input:  "1.0.0",
			output: "1.1.0",
			actions: []action{
				action{
					operation:   "increment_minor",
					incrementBy: 1,
				},
			},
			shouldFail: false,
			shouldErr:  false,
			errMessage: "",
		},
		{
			input:  "1.0.0",
			output: "1.0.1",
			actions: []action{
				action{
					operation:   "increment_patch",
					incrementBy: 1,
				},
			},
			shouldFail: false,
			shouldErr:  false,
			errMessage: "",
		},
		{
			input:  "1.0.0",
			output: "2.1.1",
			actions: []action{
				action{
					operation:   "increment_major",
					incrementBy: 1,
				},
				action{
					operation:   "increment_minor",
					incrementBy: 1,
				},
				action{
					operation:   "increment_patch",
					incrementBy: 1,
				},
			},
			shouldFail: false,
			shouldErr:  false,
			errMessage: "",
		},
		{
			description: "When minor version increments, patch should reset to zero",
			input:       "1.0.1",
			output:      "1.1.0",
			actions: []action{
				action{
					operation:   "increment_minor",
					incrementBy: 1,
				},
			},
			shouldFail: false,
			shouldErr:  false,
			errMessage: "",
		},
		{
			description: "when major version increments, both minor and patch should reset to zero",
			input:       "1.1.1",
			output:      "2.0.0",
			actions: []action{
				action{
					operation:   "increment_major",
					incrementBy: 1,
				},
			},
			shouldFail: false,
			shouldErr:  false,
			errMessage: "",
		},
		{input: "", output: "", actions: []action{}, shouldFail: false, shouldErr: true, errMessage: "empty string"},
		{
			input:      "1.1.1.1",
			output:     "",
			actions:    []action{},
			shouldFail: false,
			shouldErr:  true,
			errMessage: "version must be in major.minor.patch format",
		},
		{
			input:      "1aZ.1.1",
			output:     "",
			actions:    []action{},
			shouldFail: false,
			shouldErr:  true,
			errMessage: "failed to parse major version",
		},
		{
			input:      "1.1aZ.1",
			output:     "",
			actions:    []action{},
			shouldFail: false,
			shouldErr:  true,
			errMessage: "failed to parse minor version",
		},
		{
			input:      "1.1.1aZ",
			output:     "",
			actions:    []action{},
			shouldFail: false,
			shouldErr:  true,
			errMessage: "failed to parse patch version",
		},
	} {
		version, err := NewVersion(test.input)
		if err != nil {
			if !test.shouldErr {
				t.Logf("FAIL: Test %d: input: %s, expected output: %s, error: %s", i, test.input, test.output, err)
				testFailed++
			} else {
				if test.errMessage != err.Error() {
					t.Logf("FAIL: Test %d: input: %s, error: %s (expected) vs. %s (received)", i, test.input, test.errMessage, err)
					testFailed++
				} else {
					t.Logf("PASS: Test %d: input: %s, error: %s", i, test.input, test.errMessage)
				}
			}
			continue
		}
		for _, action := range test.actions {
			switch action.operation {
			case "increment_major":
				version.IncrementMajor(action.incrementBy)
			case "increment_minor":
				version.IncrementMinor(action.incrementBy)
			case "increment_patch":
				version.IncrementPatch(action.incrementBy)
			default:
				t.Fatalf("FAIL: Test %d: input: %s, expected output: %s, error: unsupported test action", i, test.input, test.output)
			}
		}

		failedTest := false

		if version.String() != test.output {
			if !test.shouldFail {
				failedTest = true
			}
		} else {
			if test.shouldFail {
				failedTest = true
			}
		}

		if failedTest {
			t.Logf("FAIL: Test %d: input: '%s', expected output: '%s'", i, test.input, test.output)
			testFailed++
		} else {
			t.Logf("PASS: Test %d: input: '%s', expected output: '%s'", i, test.input, test.output)
		}

	}
	if testFailed > 0 {
		t.Fatalf("Failed %d tests", testFailed)
	}
}

func TestVersionedFileOperations(t *testing.T) {
	testFailed := 0
	tempDirName := ".tmp"
	if _, err := os.Stat(tempDirName); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(".tmp", 0755)
			if err != nil {
				t.Fatalf("Failed creating temporary directory %s: %s", tempDirName, err)
			}
		} else {
			t.Fatalf("Error with temporary directory %s: %s", tempDirName, err)
		}
	}

	testInputFileNames := []string{"VERSION", "PROP_VERSION"}
	versionStr := "1.2.3"
	versionBytes := []byte(versionStr)

	for _, name := range testInputFileNames {
		tempFileName := filepath.Join(tempDirName, name)
		if err := ioutil.WriteFile(tempFileName, versionBytes, 0666); err != nil {
			t.Fatalf("Error writing to %s: %s", tempFileName, err)
		}
	}

	testInputFileNames = append(testInputFileNames, "")

	os.Chdir(tempDirName)
	for i, name := range testInputFileNames {
		version, err := NewVersionFromFile(name)
		if err != nil {
			t.Logf("FAIL: Test %d: input: '%s', error: %s", i, name, err)
			testFailed++
			continue
		}
		if version.String() != versionStr {
			t.Logf("FAIL: Test %d: input: '%s', output: %s (expected) vs. %s (received)",
				i, name, versionStr, version)
			testFailed++
			continue
		}
		if err := version.UpdateFile(); err != nil {
			t.Logf("FAIL: Test %d: input: '%s', error: %s", i, name, err)
			testFailed++
			continue
		}

		t.Logf("PASS: Test %d: input: %s, version: %s %v", i, name, version, version.Bytes())
	}

	if testFailed > 0 {
		t.Fatalf("Failed %d tests", testFailed)
	}

}

func TestAppPackage(t *testing.T) {
	app := &Package{
		Name:          "versioned",
		Description:   "Quickly increment major/manor/patch in VERSION file",
		Documentation: "https://github.com/greenpau/versioned/",
	}
	app.SetVersion("", "1.0.1")
	app.SetGitBranch("", "master")
	app.SetGitCommit("", "v1.0.1-g0c85fbc")
	app.SetBuildUser("", "greenpau")
	app.SetBuildDate("", "2020-04-25")

	app.SetVersion("1.0.1", "")
	app.SetGitBranch("master", "")
	app.SetGitCommit("v1.0.1-g0c85fbc", "")
	app.SetBuildUser("greenpau", "")
	app.SetBuildDate("2020-04-25", "")

	vers := app.Banner()
	vers = strings.Split(vers, "(")[0]
	expVers := "versioned 1.0.1, branch: master, commit: v1.0.1-g0c85fbc, build on 2020-04-25 by greenpau "
	if vers != expVers {
		t.Fatalf("FAIL: Version mismatch: %s (expected) vs. %s (received)", expVers, vers)
	}
	vers = app.ShortBanner()
	expVers = "versioned 1.0.1"

	if vers != expVers {
		t.Fatalf("FAIL: Version mismatch: %s (expected) vs. %s (received)", expVers, vers)
	}
	t.Logf("%s", app)
}
