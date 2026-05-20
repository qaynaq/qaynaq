package rag_chunker

import "github.com/warpstreamlabs/bento/public/service"

const (
	fieldStrategy  = "strategy"
	fieldChunkSize = "chunk_size"
	fieldOverlap   = "overlap"
)

func configSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Beta().
		Categories("Parsing").
		Summary("Splits a stream of text into overlapping chunks suitable for RAG indexing.").
		Description(`
The ` + "`rag_chunker`" + ` scanner reads the full content of each source, splits it
into chunks using the chosen strategy, and emits one message per chunk.

### Strategies

- ` + "`recursive`" + ` - mirrors LangChain's RecursiveCharacterTextSplitter. Tries
  splitting on paragraph, line, sentence, word, then character boundaries in that
  order, preferring the coarsest boundary that keeps every chunk within the size
  limit.
- ` + "`token`" + ` - the same algorithm, with sizes interpreted as tokens using
  a ~4 character per token approximation.
- ` + "`markdown`" + ` - splits on Markdown heading boundaries (` + "`#`, `##`, ..." + `)
  and attaches the heading path as message metadata. Sections larger than the chunk
  size are sub-split with the recursive strategy.

### Metadata

Each emitted message carries these metadata fields:

- ` + "`rag_chunk_index`" + ` - the chunk's position within its source, starting at 0.
- ` + "`rag_source`" + ` - the name of the source (filename or equivalent), when available.
- For the markdown strategy, ` + "`rag_md_h1`" + ` through ` + "`rag_md_h6`" + ` carry the
  heading path leading to the chunk.
`).
		Field(service.NewStringEnumField(fieldStrategy, StrategyRecursive, StrategyToken, StrategyMarkdown).
			Description("Splitting strategy.").
			Default(StrategyRecursive)).
		Field(service.NewIntField(fieldChunkSize).
			Description("Maximum size of each chunk. Characters for `recursive` and `markdown`, tokens for `token`.").
			Default(1000)).
		Field(service.NewIntField(fieldOverlap).
			Description("Amount of content carried from the end of one chunk into the start of the next, in the same unit as chunk_size. Aligns to the nearest boundary.").
			Default(200)).
		Version("1.0.0")
}
