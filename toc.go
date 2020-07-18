package versioned

import (
	"fmt"
	"strings"
)

// TableOfContents represent Markdown Table of Contents section.
type TableOfContents struct {
	entries []*tocEntry
}

type tocEntry struct {
	raw   string
	title string
	link  string
	level int
}

// NewTableOfContents return a new instance of TableOfContents.
func NewTableOfContents() *TableOfContents {
	return &TableOfContents{
		entries: []*tocEntry{},
	}
}

// AddHeading adds an entry to TableOfContents.
func (toc *TableOfContents) AddHeading(s string) error {
	t := &tocEntry{
		raw: s,
	}
	if s == "" {
		return fmt.Errorf("cannot add an empty string")
	}
	if !strings.HasPrefix(strings.TrimSpace(s), "#") {
		return fmt.Errorf("heading must start with a pound")
	}
	toc.entries = append(toc.entries, t)
	return nil
}

// ToString return string representation of TableOfContents.
func (toc *TableOfContents) ToString() string {
	return "ToC"
}
