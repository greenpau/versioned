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
	"testing"
)

func TestNewTableOfContents(t *testing.T) {
	toc := NewTableOfContents()
	if err := toc.AddHeading("## Heading 2"); err != nil {
		t.Fatal(err)
	}
	if err := toc.AddHeading("### Heading 3"); err != nil {
		t.Fatal(err)
	}
	if err := toc.AddHeading("#### Heading 4"); err != nil {
		t.Fatal(err)
	}
	if err := toc.AddHeading("##### Heading 5"); err != nil {
		t.Fatal(err)
	}
	if err := toc.AddHeading("##### Heading 6"); err != nil {
		t.Fatal(err)
	}

	if err := toc.AddHeading("## Heading 2"); err != nil {
		t.Fatal(err)
	}
	if err := toc.AddHeading("### Heading 3"); err != nil {
		t.Fatal(err)
	}
	if err := toc.AddHeading("#### Heading 4"); err != nil {
		t.Fatal(err)
	}
	if err := toc.AddHeading("##### Heading 5"); err != nil {
		t.Fatal(err)
	}
	if err := toc.AddHeading("##### Heading-5"); err != nil {
		t.Fatal(err)
	}
	if err := toc.AddHeading("##### Heading 5"); err != nil {
		t.Fatal(err)
	}
	if err := toc.AddHeading("##### Heading 6"); err != nil {
		t.Fatal(err)
	}
	if err := toc.AddHeading("##### Heading-6"); err != nil {
		t.Fatal(err)
	}

	if err := toc.AddHeading("## Heading_2"); err != nil {
		t.Fatal(err)
	}
	if err := toc.AddHeading("## Heading~2"); err != nil {
		t.Fatal(err)
	}

	if err := toc.AddHeading("#### Heading 4"); err == nil {
		t.Fatal("Expected level hopping error, but got success")
	}

	if err := toc.AddHeading(""); err == nil {
		t.Fatal("Expected empty string error, but got success")
	}
	if err := toc.AddHeading("Heading 1"); err == nil {
		t.Fatal("Expected pound start error, but got success")
	}

	t.Logf("\n\n%s\n", toc.ToString())
}
