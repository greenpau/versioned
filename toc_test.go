// Copyright 2020 Paul Greenberg (greenpau@outlook.com)

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
