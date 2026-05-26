# RAG Chunker

Splits text into overlapping chunks suitable for retrieval-augmented generation (RAG) indexing. Read a document with a file or blob input, scan it with the RAG Chunker, generate embeddings on each chunk, and write the result to a vector database.

For a full end-to-end example, see the [RAG Knowledge Base for AI Assistants](/playbooks/rag-knowledge-base) playbook.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Strategy | enum | `Recursive` | How the text is broken up. See below |
| Chunk Size | integer | `1000` | Maximum size of each chunk. Characters for Recursive and Markdown, tokens for Token |
| Overlap | integer | `200` | How much of the previous chunk is carried into the start of the next. Aligned to the nearest boundary. A typical value is around 20% of the chunk size |

## Strategies

### Recursive

Mirrors LangChain's `RecursiveCharacterTextSplitter`. Splits on the coarsest boundary that still keeps every chunk within the size limit, in this order:

1. Paragraphs (`\n\n`)
2. Lines (`\n`)
3. Sentences (`. `)
4. Words (` `)
5. Characters

Overlap is aligned to the nearest boundary, so the carried context never starts in the middle of a word.

### Token

The same recursive algorithm, but with Chunk Size and Overlap interpreted as tokens. Tokens are estimated as 4 characters each, which works well for English and most embedding models. No external tokenizer is required.

### Markdown

Splits on Markdown heading boundaries (`#`, `##`, ..., `######`). Each chunk carries its heading path as metadata - the body text itself is left untouched. Sections that overflow Chunk Size are sub-split using the recursive strategy, with the same heading path attached to every sub-chunk.

## Metadata

Every emitted chunk carries:

| Field | Description |
|-------|-------------|
| `rag_chunk_index` | The chunk's position in its source, starting at 0 |
| `rag_source` | The name of the source (filename or equivalent) when available |

With the Markdown strategy, the active heading path is also attached:

| Field | Description |
|-------|-------------|
| `rag_md_h1` | Active top-level heading |
| `rag_md_h2` through `rag_md_h6` | Lower-level headings, present only when set |
