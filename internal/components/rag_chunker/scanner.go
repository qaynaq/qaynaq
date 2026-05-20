package rag_chunker

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/warpstreamlabs/bento/public/service"
)

func init() {
	err := service.RegisterBatchScannerCreator("rag_chunker", configSpec(),
		func(conf *service.ParsedConfig, _ *service.Resources) (service.BatchScannerCreator, error) {
			return newCreatorFromParsed(conf)
		})
	if err != nil {
		panic(err)
	}
}

func newCreatorFromParsed(conf *service.ParsedConfig) (*creator, error) {
	strategy, err := conf.FieldString(fieldStrategy)
	if err != nil {
		return nil, err
	}
	chunkSize, err := conf.FieldInt(fieldChunkSize)
	if err != nil {
		return nil, err
	}
	overlap, err := conf.FieldInt(fieldOverlap)
	if err != nil {
		return nil, err
	}

	ch, err := New(Config{
		Strategy:  strategy,
		ChunkSize: chunkSize,
		Overlap:   overlap,
	})
	if err != nil {
		return nil, err
	}
	return &creator{chunker: ch}, nil
}

type creator struct {
	chunker Chunker
}

func (c *creator) Create(rdr io.ReadCloser, aFn service.AckFunc, details *service.ScannerSourceDetails) (service.BatchScanner, error) {
	s := &scanner{
		r:       rdr,
		chunker: c.chunker,
	}
	if details != nil {
		s.source = details.Name()
	}
	return service.AutoAggregateBatchScannerAcks(s, aFn), nil
}

func (c *creator) Close(context.Context) error {
	return nil
}

type scanner struct {
	r       io.ReadCloser
	chunker Chunker
	source  string

	chunks   []Chunk
	consumed bool
	nextIdx  int
}

func (s *scanner) NextBatch(ctx context.Context) (service.MessageBatch, error) {
	if !s.consumed {
		if err := s.readAndChunk(); err != nil {
			return nil, err
		}
	}
	if s.nextIdx >= len(s.chunks) {
		return nil, io.EOF
	}

	ch := s.chunks[s.nextIdx]
	msg := service.NewMessage([]byte(ch.Content))
	msg.MetaSet("rag_chunk_index", strconv.Itoa(s.nextIdx))
	if s.source != "" {
		msg.MetaSet("rag_source", s.source)
	}
	for k, v := range ch.Metadata {
		if str, ok := v.(string); ok {
			msg.MetaSet(metaKeyFromHeader(k), str)
		}
	}

	s.nextIdx++
	return service.MessageBatch{msg}, nil
}

func (s *scanner) readAndChunk() error {
	raw, err := io.ReadAll(s.r)
	if err != nil {
		return fmt.Errorf("rag_chunker: read source: %w", err)
	}
	s.consumed = true
	if len(raw) == 0 {
		return nil
	}
	chunks, err := s.chunker.Split(string(raw))
	if err != nil {
		return fmt.Errorf("rag_chunker: split: %w", err)
	}
	s.chunks = chunks
	return nil
}

func (s *scanner) Close(ctx context.Context) error {
	if s.r == nil {
		return nil
	}
	return s.r.Close()
}

// Bloblang metadata keys cannot contain spaces, so "Header 1" becomes "rag_md_h1".
func metaKeyFromHeader(header string) string {
	switch header {
	case "Header 1":
		return "rag_md_h1"
	case "Header 2":
		return "rag_md_h2"
	case "Header 3":
		return "rag_md_h3"
	case "Header 4":
		return "rag_md_h4"
	case "Header 5":
		return "rag_md_h5"
	case "Header 6":
		return "rag_md_h6"
	}
	return header
}
