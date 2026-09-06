package hubspot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/qaynaq/qaynaq/internal/connauth"
)

// httpDoer is the subset of *http.Client the processor uses, so tests can
// inject a fake without a live HubSpot connection.
type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func (p *Processor) client(ctx context.Context) (httpDoer, error) {
	p.clientOnce.Do(func() {
		vp := connauth.Provider()
		if vp == nil {
			p.clientErr = fmt.Errorf("vault provider not initialised")
			return
		}
		p.httpClient = connauth.NewHTTPClient(ctx, vp, p.oauthConnection)
	})
	return p.httpClient, p.clientErr
}

func (p *Processor) doGet(ctx context.Context, endpoint string) (map[string]any, error) {
	return p.do(ctx, http.MethodGet, endpoint, nil)
}

func (p *Processor) doPost(ctx context.Context, endpoint string, payload any) (map[string]any, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}
	return p.do(ctx, http.MethodPost, endpoint, encoded)
}

func (p *Processor) doPatch(ctx context.Context, endpoint string, payload any) (map[string]any, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}
	return p.do(ctx, http.MethodPatch, endpoint, encoded)
}

// doDelete issues a DELETE and treats any 2xx (including 204 No Content) as
// success. HubSpot's delete archives the record and returns no body.
func (p *Processor) doDelete(ctx context.Context, endpoint string) error {
	client, err := p.client(ctx)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("[%d] %s", resp.StatusCode, truncate(string(body), 500))
	}
	return nil
}

func (p *Processor) do(ctx context.Context, method, endpoint string, payload []byte) (map[string]any, error) {
	client, err := p.client(ctx)
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if payload != nil {
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("[%d] %s", resp.StatusCode, truncate(string(body), 500))
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return result, nil
}

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit]
}
