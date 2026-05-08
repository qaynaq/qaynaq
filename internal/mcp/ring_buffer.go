package mcp

import (
	"io"
	"sync"
)

// ringBuffer keeps the last N bytes written to it. Used to capture the tail
// of stderr from supervised child processes. Goroutine-safe.
type ringBuffer struct {
	mu   sync.Mutex
	buf  []byte
	cap  int
	head int
	full bool
}

func newRingBuffer(capacity int) *ringBuffer {
	if capacity <= 0 {
		capacity = 4 * 1024
	}
	return &ringBuffer{
		buf: make([]byte, capacity),
		cap: capacity,
	}
}

func (r *ringBuffer) Write(p []byte) (int, error) {
	if r == nil {
		return len(p), nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	n := len(p)
	if n >= r.cap {
		copy(r.buf, p[n-r.cap:])
		r.head = 0
		r.full = true
		return n, nil
	}
	end := r.head + n
	if end <= r.cap {
		copy(r.buf[r.head:end], p)
	} else {
		first := r.cap - r.head
		copy(r.buf[r.head:], p[:first])
		copy(r.buf[:n-first], p[first:])
		r.full = true
	}
	r.head = (r.head + n) % r.cap
	if end >= r.cap {
		r.full = true
	}
	return n, nil
}

func (r *ringBuffer) String() string {
	if r == nil {
		return ""
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.full {
		return string(r.buf[:r.head])
	}
	out := make([]byte, 0, r.cap)
	out = append(out, r.buf[r.head:]...)
	out = append(out, r.buf[:r.head]...)
	return string(out)
}

var _ io.Writer = (*ringBuffer)(nil)
