package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/qaynaq/qaynaq/internal/persistence"
)

// Client ID Metadata Documents (MCP auth spec 2025-11-25): the client_id is
// an HTTPS URL pointing at a JSON document describing the client.

const (
	cimdFetchTimeout = 10 * time.Second
	cimdMaxBodySize  = 256 << 10
)

var cimdHTTPClient = &http.Client{
	Timeout: cimdFetchTimeout,
	Transport: &http.Transport{
		// The URL is client-supplied; refuse dials to internal addresses
		// after DNS resolution so it cannot be used to probe the network.
		DialContext: (&net.Dialer{
			Timeout: cimdFetchTimeout,
			Control: func(_, address string, _ syscall.RawConn) error {
				host, _, err := net.SplitHostPort(address)
				if err != nil {
					return err
				}
				ip := net.ParseIP(host)
				if ip == nil || !ip.IsGlobalUnicast() || ip.IsPrivate() {
					return fmt.Errorf("client metadata document host resolves to a non-public address")
				}
				return nil
			},
		}).DialContext,
	},
	// The document must claim the exact URL it was fetched from, so
	// following redirects can only produce an invalid document.
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func isCIMDClientID(clientID string) bool {
	if !strings.HasPrefix(clientID, "https://") {
		return false
	}
	u, err := url.Parse(clientID)
	return err == nil && u.Host != ""
}

type cimdDocument struct {
	ClientID     string   `json:"client_id"`
	ClientName   string   `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris"`
}

// resolveCIMDClient fetches and validates a Client ID Metadata Document and
// upserts it as an OAuth client so the rest of the flow (consent, tokens,
// revocation) treats it like any registered client and reconnects reuse the
// same row.
func (s *Server) resolveCIMDClient(clientID string) (*persistence.OAuthClient, error) {
	req, err := http.NewRequest(http.MethodGet, clientID, nil) //nolint:gosec // CIMD requires fetching the client-supplied URL; cimdHTTPClient refuses non-public dial targets
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := cimdHTTPClient.Do(req) //nolint:gosec // see above
	if err != nil {
		return nil, fmt.Errorf("failed to fetch client metadata document: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("client metadata document returned status %d", resp.StatusCode)
	}

	var doc cimdDocument
	if err := json.NewDecoder(io.LimitReader(resp.Body, cimdMaxBodySize)).Decode(&doc); err != nil {
		return nil, fmt.Errorf("invalid client metadata document: %w", err)
	}
	if doc.ClientID != clientID {
		return nil, fmt.Errorf("client metadata document client_id does not match its URL")
	}
	if len(doc.RedirectURIs) == 0 {
		return nil, fmt.Errorf("client metadata document has no redirect_uris")
	}

	name := strings.TrimSpace(doc.ClientName)
	if name == "" {
		if u, err := url.Parse(clientID); err == nil {
			name = u.Host
		} else {
			name = clientID
		}
	}
	client := &persistence.OAuthClient{
		ID:           clientID,
		Name:         name,
		RedirectURIs: doc.RedirectURIs,
		CreatedAt:    time.Now(),
	}
	if err := s.clientRepo.Upsert(client); err != nil {
		return nil, fmt.Errorf("failed to persist client: %w", err)
	}
	return client, nil
}
