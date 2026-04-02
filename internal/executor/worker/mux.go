package worker

import (
	"net/http"
	"sync"
)

type safeMux struct {
	mu       sync.RWMutex
	handlers map[string]http.HandlerFunc
}

func newSafeMux() *safeMux {
	return &safeMux{
		handlers: make(map[string]http.HandlerFunc),
	}
}

func (m *safeMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[pattern] = handler
}

func (m *safeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	handler, ok := m.handlers[r.URL.Path]
	m.mu.RUnlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	handler(w, r)
}
