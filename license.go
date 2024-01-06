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

package versioned

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"

	// "log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

var (
	licenseTemplates = map[string]string{
		"apache": tmplApache,
		"asl":    tmplAsl,
		"mit":    tmplMit,
		"gpl3":   tmplGpl3,
	}

	licenseClues = map[string]string{
		"apache": "Licensed under the Apache License, Version 2.0",
		"asl":    "Licensed under the Amazon Software License",
		"mit":    "Licensed under the MIT License",
		"gpl3":   "Licensed under the GPLv3 License",
	}
)

// LicenseHeader represent license headers.
type LicenseHeader struct {
	FilePath        string
	FileExtension   string
	Year            uint64
	CopyrightHolder string
	LicenseType     string
	Action          string
	wrapChars       []string
	raw             []byte
	offset          int
	found           bool
	match           bool
	mismatchText    string
}

// NewLicenseHeader returns an instance of LicenseHeader.
func NewLicenseHeader() *LicenseHeader {
	return &LicenseHeader{
		LicenseType: "apache",
	}
}

// AddLicense adds a license header to a file.
func AddLicense(h *LicenseHeader) error {
	if err := h.build(); err != nil {
		return err
	}
	// log.Printf("header:\n%s", h.raw)
	if err := h.inspect(); err != nil {
		return err
	}
	if h.found {
		if !h.match {
			return fmt.Errorf("found license header mismatch in %q, %s", h.FilePath, h.mismatchText)
		}
		return nil
	}
	if err := h.rewrite("add"); err != nil {
		return fmt.Errorf("encountered error adding license header: %v", err)
	}
	return nil
}

// StripLicense remove a license header from a file.
func StripLicense(h *LicenseHeader) error {
	if err := h.build(); err != nil {
		return err
	}
	// log.Printf("header:\n%s", h.raw)
	if err := h.inspect(); err != nil {
		return err
	}
	if h.found {
		if err := h.rewrite("strip"); err != nil {
			return fmt.Errorf("encountered error stripping license header: %v", err)
		}
		return nil
	}
	return nil
}

func (h *LicenseHeader) inspect() error {
	h.offset = len(h.raw) + 100
	fh, err := os.Open(h.FilePath)
	if err != nil {
		return fmt.Errorf("failed opening file %q: %v", h.FilePath, err)
	}
	defer fh.Close()
	header := make([]byte, h.offset)
	if _, err := fh.Read(header); err != nil {
		if err.Error() != "EOF" {
			return fmt.Errorf("failed reading file %q: %v", h.FilePath, err)
		}
		return nil
	}

	// Remove irrelevant content from the header
	reCode1 := regexp.MustCompile(`#!/.*\n`)
	reCode2 := regexp.MustCompile(`#.*coding: .*\n`)
	// Remove code instructions
	header = reCode1.ReplaceAll(header, []byte(""))
	header = reCode2.ReplaceAll(header, []byte(""))
	header = bytes.TrimSpace(header)
	// TODO(greenpau): Remove lines that do not have copyright.
	// See h.wrapChars

	if bytes.Index(header, []byte(licenseClues[h.LicenseType])) >= 0 {
		h.found = true

		if bytes.Index(bytes.TrimSpace(header), bytes.TrimSpace(h.raw)) >= 0 {
			h.match = true
		}
		// Approximate match.
		if !h.match {
			reY := regexp.MustCompile(`\s(\d{4})\s`)
			reW := regexp.MustCompile(`\s*`)
			// Remove copyright year.
			actual := reY.ReplaceAll(header, []byte(" "))
			expected := reY.ReplaceAll(h.raw, []byte(" "))
			// Remove all whitespaces.
			actual = reW.ReplaceAll(actual, []byte(""))
			expected = reW.ReplaceAll(expected, []byte(""))
			if bytes.Index(actual, expected) >= 0 {
				h.match = true
			}
		}

		if !h.match {
			h.mismatchText = fmt.Sprintf("\n>>>got:\n%s\n>>>expected:\n%s", header, h.raw)
		}
	}
	if !h.found && bytes.Index(header, []byte("Copyright ")) >= 0 {
		h.found = true
	}
	return nil
}

func (h *LicenseHeader) rewrite(action string) error {
	var offset int
	switch action {
	case "add", "strip":
	default:
		return fmt.Errorf("unsupported action: %q", action)
	}

	b, err := ioutil.ReadFile(h.FilePath)
	if err != nil {
		if err.Error() != "EOF" {
			return fmt.Errorf("failed reading file %q: %v", h.FilePath, err)
		}
	}

	if action == "strip" {
		for _, ls := range []string{"\n\n", "\r\r", "\r\n\r\n"} {
			offset = bytes.Index(b, []byte(h.wrapChars[2]+ls))
			if offset > 0 {
				offset = offset + len([]byte(h.wrapChars[2]+ls))
				break
			}
		}
		if offset < 1 {
			return nil
		}
	}

	fi, err := os.Stat(h.FilePath)
	if err != nil {
		return fmt.Errorf("failed getting info for file %q: %v", h.FilePath, err)
	}

	fh, err := os.OpenFile(h.FilePath, os.O_WRONLY|os.O_TRUNC, fi.Mode().Perm())
	if err != nil {
		return fmt.Errorf("failed opening file %q for writing: %v", h.FilePath, err)
	}
	defer fh.Close()

	if err := fh.Truncate(0); err != nil {
		return fmt.Errorf("failed truncating file %q: %v", h.FilePath, err)
	}

	if _, err := fh.Seek(0, 0); err != nil {
		return fmt.Errorf("failed seeking the beginning of file %q: %v", h.FilePath, err)
	}

	switch action {
	case "add":
		if len(b) > 0 {

			switch h.FileExtension {
			case ".py":
				// Determine code instructions and only add the header after them.
				injectByteID := 0
				for _, ptrn := range []string{`.{0,100}#!/.*\n`, `.{0,100}#.*\scoding:.*\n`} {
					reCode := regexp.MustCompile(ptrn)
					if reCode.Match(b) {
						match := reCode.FindAllIndex(b, 1)
						if match != nil && len(match) > 0 {
							if match[0][1] > injectByteID {
								injectByteID = match[0][1]
							}
						}
					}
				}
				if injectByteID == 0 {
					if _, err := fh.Write(h.raw); err != nil {
						return fmt.Errorf("failed prepending header to file %q: %v", h.FilePath, err)
					}
					if _, err := fh.Write(b); err != nil {
						return fmt.Errorf("failed writing existing content to file %q: %v", h.FilePath, err)
					}
				} else {
					if _, err := fh.Write(b[:injectByteID]); err != nil {
						return fmt.Errorf("failed writing existing content to file %q: %v", h.FilePath, err)
					}
					fh.Write([]byte("\n"))
					if _, err := fh.Write(h.raw); err != nil {
						return fmt.Errorf("failed prepending header to file %q: %v", h.FilePath, err)
					}
					if _, err := fh.Write(b[injectByteID:]); err != nil {
						return fmt.Errorf("failed writing existing content to file %q: %v", h.FilePath, err)
					}
				}
			default:
				if _, err := fh.Write(h.raw); err != nil {
					return fmt.Errorf("failed prepending header to file %q: %v", h.FilePath, err)
				}
				if _, err := fh.Write(b); err != nil {
					return fmt.Errorf("failed writing existing content to file %q: %v", h.FilePath, err)
				}
			}
			return nil
		}
		if len(b) == 0 {
			if _, err := fh.Write(h.raw); err != nil {
				return fmt.Errorf("failed prepending header to file %q: %v", h.FilePath, err)
			}
			return nil
		}

	case "strip":
		if _, err := fh.Write(b[offset:]); err != nil {
			return fmt.Errorf("failed writing existing content to file %q: %v", h.FilePath, err)
		}
	}
	return nil
}

// AddFilePath adds the path to a file.
func (h *LicenseHeader) AddFilePath(fp string) error {
	if fp == "" {
		return fmt.Errorf("file path is empty")
	}
	h.FilePath = fp
	return nil
}

// AddCopyrightHolder adds copyright holder.
func (h *LicenseHeader) AddCopyrightHolder(s string) error {
	if s == "" {
		return fmt.Errorf("copyright holder is empty")
	}
	h.CopyrightHolder = s
	return nil
}

// AddYear adds copyright year.
func (h *LicenseHeader) AddYear(i uint64) error {
	if i == 0 {
		return fmt.Errorf("copyright year is empty")
	}
	h.Year = i
	return nil
}

// AddLicenseType adds license type.
func (h *LicenseHeader) AddLicenseType(s string) error {
	switch s {
	case "apache":
	case "asl":
	case "mit":
	case "gpl3", "gplv3":
		s = "gpl3"
	case "":
		s = "apache"
	default:
		return fmt.Errorf("license type %q is unsupported", s)
	}
	h.LicenseType = s
	return nil
}

func (h *LicenseHeader) getWrapChars() error {
	if h.FileExtension == "" {
		h.FileExtension = filepath.Ext(h.FilePath)
	}
	if h.FileExtension == "" {
		return fmt.Errorf("failed determining file extension for %q", h.FilePath)
	}

	switch h.FileExtension {
	case ".go":
		h.wrapChars = []string{"", "// ", ""}
	case ".py":
		h.wrapChars = []string{"#", "# ", "#"}
	case ".js", ".ts", ".tsx", ".mjs":
		h.wrapChars = []string{"/**", " * ", " */"}
	case ".swift":
		h.wrapChars = []string{"", "// ", ""}
	default:
		return fmt.Errorf("license header unsupported for file extension %q in %q", h.FileExtension, h.FilePath)
	}
	return nil
}

func (h *LicenseHeader) build() error {
	if err := h.getWrapChars(); err != nil {
		return err
	}
	t, err := template.New("").Parse(licenseTemplates[h.LicenseType])
	if err != nil {
		return fmt.Errorf("failed parsing template: %v", err)
	}

	var b bytes.Buffer
	if err := t.Execute(&b, h); err != nil {
		return fmt.Errorf("failed executing template: %v", err)
	}

	var header bytes.Buffer
	if h.wrapChars[0] != "" {
		fmt.Fprintln(&header, h.wrapChars[0])
	}
	r := bufio.NewScanner(&b)
	for r.Scan() {
		s := strings.TrimRightFunc(h.wrapChars[1]+r.Text(), unicode.IsSpace)
		fmt.Fprintln(&header, s)
	}
	if h.wrapChars[2] != "" {
		fmt.Fprintln(&header, h.wrapChars[2])
	}
	fmt.Fprintln(&header)
	h.raw = header.Bytes()
	return nil
}

const tmplApache = `Copyright {{.Year}} {{.CopyrightHolder}}

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.`

const tmplAsl = `Copyright {{.Year}} {{.CopyrightHolder}}. All Rights Reserved.

Licensed under the Amazon Software License (the "License").
You may not use this file except in compliance with the License.
A copy of the License is located at

    http://aws.amazon.com/asl/

or in the "license" file accompanying this file. This file is distributed
on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
express or implied. See the License for the specific language governing
permissions and limitations under the License.`

const tmplMit = `Copyright (c) {{.Year}} {{.CopyrightHolder}}

Licensed under the MIT License.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.`

const tmplGpl3 = `Copyright (C) {{.Year}} {{.CopyrightHolder}}

Licensed under the GPLv3 License.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.`
