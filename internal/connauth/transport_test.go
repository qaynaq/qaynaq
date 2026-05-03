package connauth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/qaynaq/qaynaq/internal/vault"
)

type stubVault struct {
	tokens         []string // sequence to return on each token fetch
	calls          int32
	forceRefreshes int32
	getAccessCalls int32
}

func (s *stubVault) GetSecret(string) (string, error) { return "", nil }
func (s *stubVault) GetConnectionToken(string) (string, error) {
	return "", nil
}

func (s *stubVault) tokenAt(idx int32) string {
	if int(idx) < len(s.tokens) {
		return s.tokens[idx]
	}
	return s.tokens[len(s.tokens)-1]
}

func (s *stubVault) GetAccessToken(string) (vault.AccessToken, error) {
	idx := atomic.AddInt32(&s.getAccessCalls, 1) - 1
	atomic.AddInt32(&s.calls, 1)
	return vault.AccessToken{AccessToken: s.tokenAt(idx), ExpiresAt: time.Now().Add(1 * time.Hour)}, nil
}

func (s *stubVault) ForceRefreshAccessToken(string) (vault.AccessToken, error) {
	idx := atomic.AddInt32(&s.getAccessCalls, 1) - 1
	atomic.AddInt32(&s.calls, 1)
	atomic.AddInt32(&s.forceRefreshes, 1)
	return vault.AccessToken{AccessToken: s.tokenAt(idx), ExpiresAt: time.Now().Add(1 * time.Hour)}, nil
}

func TestRetryOn401InvalidatesAndRetriesOnce(t *testing.T) {
	var serverCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&serverCalls, 1)
		got := r.Header.Get("Authorization")
		if n == 1 {
			if got != "Bearer first" {
				t.Errorf("first call expected Bearer first, got %q", got)
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if got != "Bearer second" {
			t.Errorf("retry expected Bearer second, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	vp := &stubVault{tokens: []string{"first", "second"}}
	c := NewHTTPClient(context.Background(), vp, "conn1")

	resp, err := c.Get(srv.URL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("final status: want 200, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&vp.forceRefreshes); got != 1 {
		t.Errorf("force refreshes: want 1, got %d", got)
	}
	if got := atomic.LoadInt32(&serverCalls); got != 2 {
		t.Errorf("server calls: want 2, got %d", got)
	}
}

func TestRetryOn401NotRetriedOnNon401(t *testing.T) {
	var serverCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&serverCalls, 1)
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	vp := &stubVault{tokens: []string{"first"}}
	c := NewHTTPClient(context.Background(), vp, "conn1")

	resp, err := c.Get(srv.URL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status: want 403, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&vp.forceRefreshes); got != 0 {
		t.Errorf("force refreshes: want 0, got %d", got)
	}
	if got := atomic.LoadInt32(&serverCalls); got != 1 {
		t.Errorf("server calls: want 1, got %d", got)
	}
}

func TestRetryOn401WithReplayableBody(t *testing.T) {
	var serverCalls int32
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(body))
		n := atomic.AddInt32(&serverCalls, 1)
		if n == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	vp := &stubVault{tokens: []string{"first", "second"}}
	c := NewHTTPClient(context.Background(), vp, "conn1")

	resp, err := c.Post(srv.URL, "text/plain", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("final status: want 200, got %d", resp.StatusCode)
	}
	if len(bodies) != 2 {
		t.Fatalf("bodies: want 2, got %d", len(bodies))
	}
	for i, b := range bodies {
		if b != "hello" {
			t.Errorf("body[%d]: want 'hello', got %q", i, b)
		}
	}
}
