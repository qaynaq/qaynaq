package rag_chunker

import (
	"fmt"
	"strings"
)

const (
	StrategyRecursive = "recursive"
	StrategyToken     = "token"
	StrategyMarkdown  = "markdown"
)

type Config struct {
	Strategy  string
	ChunkSize int
	Overlap   int
}

type Chunk struct {
	Content  string
	Metadata map[string]any
}

type Chunker interface {
	Split(text string) ([]Chunk, error)
}

func New(cfg Config) (Chunker, error) {
	if cfg.ChunkSize <= 0 {
		return nil, fmt.Errorf("chunk_size must be positive, got %d", cfg.ChunkSize)
	}
	if cfg.Overlap < 0 {
		return nil, fmt.Errorf("overlap must be non-negative, got %d", cfg.Overlap)
	}
	if cfg.Overlap >= cfg.ChunkSize {
		return nil, fmt.Errorf("overlap (%d) must be smaller than chunk_size (%d)", cfg.Overlap, cfg.ChunkSize)
	}

	switch cfg.Strategy {
	case "", StrategyRecursive:
		return &recursiveChunker{chunkSize: cfg.ChunkSize, overlap: cfg.Overlap}, nil
	case StrategyToken:
		// 1 token ~= 4 characters.
		return &recursiveChunker{chunkSize: cfg.ChunkSize * 4, overlap: cfg.Overlap * 4}, nil
	case StrategyMarkdown:
		return &markdownChunker{
			recursive: &recursiveChunker{chunkSize: cfg.ChunkSize, overlap: cfg.Overlap},
			chunkSize: cfg.ChunkSize,
		}, nil
	default:
		return nil, fmt.Errorf("unknown strategy %q", cfg.Strategy)
	}
}

// Coarsest to finest, matching LangChain's RecursiveCharacterTextSplitter.
var recursiveSeparators = []string{"\n\n", "\n", ". ", " ", ""}

type recursiveChunker struct {
	chunkSize int
	overlap   int
}

func (r *recursiveChunker) Split(text string) ([]Chunk, error) {
	if text == "" {
		return nil, nil
	}
	parts := r.split(text, recursiveSeparators)
	out := make([]Chunk, len(parts))
	for i, p := range parts {
		out[i] = Chunk{Content: p}
	}
	return out, nil
}

func (r *recursiveChunker) split(text string, seps []string) []string {
	if len(text) <= r.chunkSize {
		return []string{text}
	}

	separator := seps[len(seps)-1]
	nextSeps := []string{}
	for i, s := range seps {
		if s == "" {
			separator = s
			break
		}
		if strings.Contains(text, s) {
			separator = s
			nextSeps = seps[i+1:]
			break
		}
	}

	splits := splitKeepingSeparator(text, separator)

	var (
		result []string
		good   []string
	)
	flush := func() {
		if len(good) == 0 {
			return
		}
		result = append(result, r.mergeSplits(good)...)
		good = good[:0]
	}
	for _, s := range splits {
		if len(s) <= r.chunkSize {
			good = append(good, s)
			continue
		}
		flush()
		if len(nextSeps) > 0 {
			result = append(result, r.split(s, nextSeps)...)
		} else {
			result = append(result, s)
		}
	}
	flush()
	return result
}

// Each separator stays attached to the chunk that follows it, so concatenating
// the result reproduces the input.
func splitKeepingSeparator(text, sep string) []string {
	if sep == "" {
		out := make([]string, 0, len(text))
		for _, r := range text {
			out = append(out, string(r))
		}
		return out
	}
	pieces := strings.Split(text, sep)
	out := make([]string, 0, len(pieces))
	for i, p := range pieces {
		if i == 0 {
			out = append(out, p)
		} else {
			out = append(out, sep+p)
		}
	}
	filtered := out[:0]
	for _, s := range out {
		if s != "" {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// Overlap is carried as whole splits, never mid-word.
func (r *recursiveChunker) mergeSplits(splits []string) []string {
	var (
		result  []string
		current []string
		total   int
	)
	for _, s := range splits {
		if total+len(s) > r.chunkSize && len(current) > 0 {
			result = append(result, strings.Join(current, ""))
			current, total = r.applyOverlap(current)
		}
		current = append(current, s)
		total += len(s)
	}
	if len(current) > 0 {
		result = append(result, strings.Join(current, ""))
	}
	return result
}

func (r *recursiveChunker) applyOverlap(current []string) ([]string, int) {
	if r.overlap == 0 {
		return current[:0], 0
	}
	var (
		carried []string
		size    int
	)
	for i := len(current) - 1; i >= 0; i-- {
		piece := current[i]
		if size+len(piece) > r.overlap && len(carried) > 0 {
			break
		}
		carried = append([]string{piece}, carried...)
		size += len(piece)
	}
	return carried, size
}

type markdownChunker struct {
	recursive *recursiveChunker
	chunkSize int
}

type mdSection struct {
	headers map[string]string
	body    strings.Builder
}

func (m *markdownChunker) Split(text string) ([]Chunk, error) {
	if text == "" {
		return nil, nil
	}

	sections := m.parseSections(text)

	var out []Chunk
	for _, sec := range sections {
		body := strings.TrimSpace(sec.body.String())
		if body == "" {
			continue
		}
		if len(body) <= m.chunkSize {
			out = append(out, Chunk{Content: body, Metadata: copyHeaders(sec.headers)})
			continue
		}
		subs, err := m.recursive.Split(body)
		if err != nil {
			return nil, err
		}
		for _, sub := range subs {
			out = append(out, Chunk{Content: sub.Content, Metadata: copyHeaders(sec.headers)})
		}
	}
	return out, nil
}

func copyHeaders(src map[string]string) map[string]any {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]any, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func (m *markdownChunker) parseSections(text string) []mdSection {
	lines := strings.Split(text, "\n")

	var (
		sections []mdSection
		current  mdSection
		headers  = map[string]string{}
	)

	flush := func() {
		if current.body.Len() == 0 && len(current.headers) == 0 {
			return
		}
		sections = append(sections, current)
		current = mdSection{}
	}

	for _, line := range lines {
		if level, title, ok := parseHeading(line); ok {
			flush()
			for l := level; l <= 6; l++ {
				delete(headers, headerKey(l))
			}
			headers[headerKey(level)] = title
			current.headers = copyHeaderStrings(headers)
			continue
		}
		if current.headers == nil {
			current.headers = copyHeaderStrings(headers)
		}
		if current.body.Len() > 0 {
			current.body.WriteByte('\n')
		}
		current.body.WriteString(line)
	}
	flush()
	return sections
}

func parseHeading(line string) (level int, title string, ok bool) {
	trimmed := strings.TrimLeft(line, " ")
	if !strings.HasPrefix(trimmed, "#") {
		return 0, "", false
	}
	i := 0
	for i < len(trimmed) && trimmed[i] == '#' {
		i++
	}
	if i == 0 || i > 6 {
		return 0, "", false
	}
	if i < len(trimmed) && trimmed[i] != ' ' {
		return 0, "", false
	}
	return i, strings.TrimSpace(trimmed[i:]), true
}

func headerKey(level int) string {
	return fmt.Sprintf("Header %d", level)
}

func copyHeaderStrings(src map[string]string) map[string]string {
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}
