---
sidebar_position: 4
---

# Scanners

Scanners turn a continuous byte stream into discrete messages. Wherever a component accepts a stream of data (file inputs, blob storage, HTTP downloads), you choose a scanner to decide how the bytes are split.

## RAG

| Scanner | Description |
|---------|-------------|
| [RAG Chunker](/docs/components/scanners/rag-chunker) | Split text into overlapping chunks for RAG indexing, with recursive, token, or markdown strategies |

## Text

| Scanner | Description |
|---------|-------------|
| [Lines](/docs/components/scanners/lines) | One message per line. The default for most text inputs |
| [CSV](/docs/components/scanners/csv) | One message per CSV row, with header support |
| [Regex Match](/docs/components/scanners/re-match) | Split wherever a regular expression matches |

## Structured

| Scanner | Description |
|---------|-------------|
| [JSON Documents](/docs/components/scanners/json-documents) | One message per JSON document in the stream |
| [XML Documents](/docs/components/scanners/xml-documents) | One message per XML document, optionally converted to JSON |
| [Avro](/docs/components/scanners/avro) | Consume Avro Object Container Files |

## Binary

| Scanner | Description |
|---------|-------------|
| [Chunker](/docs/components/scanners/chunker) | Fixed-size byte chunks |
| [Tar](/docs/components/scanners/tar) | One message per file inside a tar archive |
| [To The End](/docs/components/scanners/to-the-end) | Read the whole stream as a single message |

## Composite

These scanners wrap another scanner, transforming the byte stream before handing it off.

| Scanner | Description |
|---------|-------------|
| [Skip BOM](/docs/components/scanners/skip-bom) | Strip a leading byte order mark, then delegate |
| [Decompress](/docs/components/scanners/decompress) | Decompress with gzip, zstd, etc., then delegate |
