package rag_chunker

import (
	"strings"
	"testing"
)

func TestNew_InvalidStrategy(t *testing.T) {
	_, err := New(Config{Strategy: "bogus", ChunkSize: 100, Overlap: 10})
	if err == nil {
		t.Fatal("expected error for unknown strategy")
	}
}

func TestNew_InvalidSizes(t *testing.T) {
	cases := []Config{
		{Strategy: "recursive", ChunkSize: 0, Overlap: 0},
		{Strategy: "recursive", ChunkSize: 100, Overlap: -1},
		{Strategy: "recursive", ChunkSize: 100, Overlap: 100},
		{Strategy: "recursive", ChunkSize: 100, Overlap: 150},
	}
	for _, c := range cases {
		if _, err := New(c); err == nil {
			t.Fatalf("expected error for %+v", c)
		}
	}
}

func TestRecursiveSplit_ShortText(t *testing.T) {
	c, err := New(Config{Strategy: "recursive", ChunkSize: 100, Overlap: 10})
	if err != nil {
		t.Fatal(err)
	}
	chunks, err := c.Split("hello world")
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 1 || chunks[0].Content != "hello world" {
		t.Fatalf("expected single passthrough chunk, got %#v", chunks)
	}
}

func TestRecursiveSplit_Empty(t *testing.T) {
	c, _ := New(Config{Strategy: "recursive", ChunkSize: 100, Overlap: 10})
	chunks, err := c.Split("")
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 0 {
		t.Fatalf("expected no chunks for empty input, got %d", len(chunks))
	}
}

func TestRecursiveSplit_RespectsChunkSize(t *testing.T) {
	text := strings.Repeat("paragraph one.\n\n", 50) // 16 * 50 = 800 chars
	c, _ := New(Config{Strategy: "recursive", ChunkSize: 100, Overlap: 20})
	chunks, err := c.Split(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
	// Each chunk should be within the upper bound. We allow chunkSize + overlap
	// slack since the merge step packs splits up to chunkSize and then carries
	// trailing splits forward.
	max := 100 + 20
	for i, ch := range chunks {
		if len(ch.Content) > max {
			t.Fatalf("chunk %d exceeds bound: %d > %d", i, len(ch.Content), max)
		}
	}
}

func TestRecursiveSplit_OverlapBoundaryAligned(t *testing.T) {
	// Build text from three paragraphs each smaller than chunkSize but together
	// larger. Overlap should be a whole paragraph carried into the next chunk,
	// not an arbitrary character slice.
	p1 := strings.Repeat("a", 40)
	p2 := strings.Repeat("b", 40)
	p3 := strings.Repeat("c", 40)
	text := p1 + "\n\n" + p2 + "\n\n" + p3
	c, _ := New(Config{Strategy: "recursive", ChunkSize: 90, Overlap: 50})
	chunks, err := c.Split(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}
	// The second chunk should start with one of the paragraph bodies, never
	// the middle of one (no "aa..." truncated to "a"*15).
	second := chunks[1].Content
	if !strings.HasPrefix(second, p1) &&
		!strings.HasPrefix(second, "\n\n"+p2) &&
		!strings.HasPrefix(second, p2) &&
		!strings.HasPrefix(second, "\n\n"+p3) &&
		!strings.HasPrefix(second, p3) {
		t.Fatalf("second chunk does not start at a paragraph boundary: %q", second)
	}
}

func TestRecursiveSplit_NoBoundaryFallback(t *testing.T) {
	// Single very long run with no separators at all: must fall through to the
	// character-level fallback and still respect chunkSize.
	text := strings.Repeat("x", 250)
	c, _ := New(Config{Strategy: "recursive", ChunkSize: 100, Overlap: 0})
	chunks, _ := c.Split(text)
	if len(chunks) < 3 {
		t.Fatalf("expected at least 3 chunks for 250 chars of no-boundary text, got %d", len(chunks))
	}
	for i, ch := range chunks {
		if len(ch.Content) > 100 {
			t.Fatalf("chunk %d exceeds chunkSize: %d", i, len(ch.Content))
		}
	}
}

func TestTokenSplit_UsesCharBudget(t *testing.T) {
	// token strategy uses chunkSize*4 chars internally. A 200-token chunk size
	// should comfortably hold a 500-char run as a single chunk.
	c, _ := New(Config{Strategy: "token", ChunkSize: 200, Overlap: 20})
	chunks, _ := c.Split(strings.Repeat("a", 500))
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk for 500 chars under 200-token budget, got %d", len(chunks))
	}
}

func TestMarkdownSplit_AttachesHeaderPath(t *testing.T) {
	text := `# Top

Intro text.

## Section A

Body of A.

## Section B

Body of B.
`
	c, _ := New(Config{Strategy: "markdown", ChunkSize: 1000, Overlap: 0})
	chunks, err := c.Split(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(chunks))
	}
	if got := chunks[0].Metadata["Header 1"]; got != "Top" {
		t.Errorf("chunk 0 Header 1 = %v, want Top", got)
	}
	if got := chunks[1].Metadata["Header 2"]; got != "Section A" {
		t.Errorf("chunk 1 Header 2 = %v, want Section A", got)
	}
	if got := chunks[1].Metadata["Header 1"]; got != "Top" {
		t.Errorf("chunk 1 Header 1 = %v, want Top (inherited)", got)
	}
	if got := chunks[2].Metadata["Header 2"]; got != "Section B" {
		t.Errorf("chunk 2 Header 2 = %v, want Section B", got)
	}
}

func TestMarkdownSplit_DoesNotPrependHeadingsToContent(t *testing.T) {
	text := "# Top\n\nBody under top."
	c, _ := New(Config{Strategy: "markdown", ChunkSize: 1000, Overlap: 0})
	chunks, _ := c.Split(text)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if strings.Contains(chunks[0].Content, "#") {
		t.Errorf("content should not contain heading marker, got %q", chunks[0].Content)
	}
	if chunks[0].Content != "Body under top." {
		t.Errorf("content = %q, want %q", chunks[0].Content, "Body under top.")
	}
}

func TestMarkdownSplit_OversizedSectionGetsSubSplit(t *testing.T) {
	body := strings.Repeat("filler line.\n", 30) // ~390 chars
	text := "# Big\n\n" + body
	c, _ := New(Config{Strategy: "markdown", ChunkSize: 100, Overlap: 10})
	chunks, _ := c.Split(text)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple sub-chunks, got %d", len(chunks))
	}
	for i, ch := range chunks {
		if got := ch.Metadata["Header 1"]; got != "Big" {
			t.Errorf("sub-chunk %d Header 1 = %v, want Big", i, got)
		}
	}
}

func TestMarkdownSplit_HeaderResetsLowerLevels(t *testing.T) {
	text := `# A

a body

## B1

b body

# C

c body
`
	c, _ := New(Config{Strategy: "markdown", ChunkSize: 1000, Overlap: 0})
	chunks, _ := c.Split(text)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	if _, has := chunks[2].Metadata["Header 2"]; has {
		t.Errorf("chunk under # C should not inherit Header 2 from earlier branch")
	}
	if got := chunks[2].Metadata["Header 1"]; got != "C" {
		t.Errorf("chunk 2 Header 1 = %v, want C", got)
	}
}

func TestParseHeading(t *testing.T) {
	cases := []struct {
		line   string
		level  int
		title  string
		ok     bool
	}{
		{"# Hello", 1, "Hello", true},
		{"### Nested  ", 3, "Nested", true},
		{"###### Six", 6, "Six", true},
		{"####### Seven", 0, "", false},
		{"#no space", 0, "", false},
		{"Not a heading", 0, "", false},
		{"  ## Indented", 2, "Indented", true},
	}
	for _, tc := range cases {
		level, title, ok := parseHeading(tc.line)
		if ok != tc.ok || level != tc.level || title != tc.title {
			t.Errorf("parseHeading(%q) = (%d, %q, %v), want (%d, %q, %v)",
				tc.line, level, title, ok, tc.level, tc.title, tc.ok)
		}
	}
}
