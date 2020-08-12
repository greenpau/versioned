package versioned

import (
	"bytes"
	"fmt"
	"strings"
)

const allowedLinkChars = "0123456789abcdefghijklmnopqrstuvwxyz-"

// TableOfContents represent Markdown Table of Contents section.
type TableOfContents struct {
	entries   []*tocEntry
	maxDepth  int
	minDepth  int
	lastDepth int
	sep       string
	linkRef   map[string]int
}

type tocEntry struct {
	title string
	link  string
	depth int
}

// NewTableOfContents return a new instance of TableOfContents.
func NewTableOfContents() *TableOfContents {
	return &TableOfContents{
		entries:  []*tocEntry{},
		minDepth: 1000,
		maxDepth: 0,
		sep:      "*",
		linkRef:  make(map[string]int),
	}
}

// AddHeading adds an entry to TableOfContents.
func (toc *TableOfContents) AddHeading(s string) error {
	if s == "" {
		return fmt.Errorf("cannot add an empty string")
	}
	if !strings.HasPrefix(strings.TrimSpace(s), "#") {
		return fmt.Errorf("heading must start with a pound")
	}
	arr := strings.SplitN(s, " ", 2)
	h := &tocEntry{
		depth: len(arr[0]),
		title: strings.TrimSpace(arr[1]),
	}
	if h.depth > toc.maxDepth {
		toc.maxDepth = h.depth
	}
	if h.depth < toc.minDepth {
		toc.minDepth = h.depth
	}
	depthDiff := h.depth - toc.lastDepth
	if (depthDiff) > 1 && toc.lastDepth > 0 {
		return fmt.Errorf(
			"heading hopped more than one level: %d, %d (current) vs. %d (previous)",
			depthDiff, h.depth, toc.lastDepth,
		)
	}
	toc.lastDepth = h.depth
	toc.entries = append(toc.entries, h)
	return nil
}

func (toc *TableOfContents) getLink(s string) string {
	s = strings.ToLower(s)
	link := "#"
	for _, c := range s {
		if string(c) == " " {
			link += "-"
			continue
		}
		if !strings.Contains(allowedLinkChars, string(c)) {
			continue
		}
		link += string(c)
	}
	i, exists := toc.linkRef[link]
	if exists {
		toc.linkRef[link]++
		link = fmt.Sprintf("%s-%d", link, i)
	} else {
		toc.linkRef[link] = 1
	}

	return link
}

// ToString return string representation of TableOfContents.
func (toc *TableOfContents) ToString() string {
	var tocBuffer bytes.Buffer
	for _, h := range toc.entries {
		offsetDepth := h.depth - toc.minDepth
		tocBuffer.WriteString(strings.Repeat("  ", offsetDepth))
		tocBuffer.WriteString(fmt.Sprintf("%s [%s](%s)", toc.sep, h.title, toc.getLink(h.title)))
		tocBuffer.WriteString("\n")
	}
	return tocBuffer.String()
}
