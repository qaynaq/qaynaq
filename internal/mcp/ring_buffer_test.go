package mcp

import (
	"strings"
	"testing"
)

func TestRingBuffer_BelowCap(t *testing.T) {
	r := newRingBuffer(64)
	_, _ = r.Write([]byte("hello"))
	if got := r.String(); got != "hello" {
		t.Fatalf("expected 'hello', got %q", got)
	}
}

func TestRingBuffer_Wrap(t *testing.T) {
	r := newRingBuffer(8)
	_, _ = r.Write([]byte("0123456789abcdef")) // 16 bytes into 8-byte buffer
	got := r.String()
	if got != "9abcdef" && got != "89abcdef" {
		t.Fatalf("expected last 8 bytes, got %q", got)
	}
	if len(got) != 8 {
		t.Fatalf("expected len 8, got %d", len(got))
	}
}

func TestRingBuffer_MultiWrite(t *testing.T) {
	r := newRingBuffer(10)
	_, _ = r.Write([]byte("abc"))
	_, _ = r.Write([]byte("defghij")) // total 10
	_, _ = r.Write([]byte("kl"))      // wraps
	got := r.String()
	if !strings.HasSuffix(got, "kl") {
		t.Fatalf("expected suffix 'kl', got %q", got)
	}
	if len(got) != 10 {
		t.Fatalf("expected full ring, len %d, got %q", len(got), got)
	}
}

func TestRingBuffer_Nil(t *testing.T) {
	var r *ringBuffer
	if _, err := r.Write([]byte("x")); err != nil {
		t.Fatalf("nil ring write should be no-op: %v", err)
	}
	if r.String() != "" {
		t.Fatal("nil ring String should be empty")
	}
}
