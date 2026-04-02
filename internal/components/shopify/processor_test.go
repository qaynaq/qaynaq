package shopify

import (
	"errors"
	"strings"
	"testing"
)

func TestClassifyShopifyError(t *testing.T) {
	tests := []struct {
		name    string
		err     string
		wantPfx string
	}{
		{"required field", "order_id is required for get_order action", "[400]"},
		{"interpolation error", "failed to interpolate limit: bad value", "[400]"},
		{"unsupported action", "unsupported action: foo", "[400]"},
		{"auth error", "401 Unauthorized", "[401]"},
		{"not found", "404 Not Found", "[404]"},
		{"rate limit", "429 rate limit exceeded", "[429]"},
		{"unknown error", "something went wrong", "[500]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := classifyShopifyError(errors.New(tt.err))
			if !strings.HasPrefix(err.Error(), tt.wantPfx) {
				t.Errorf("classifyShopifyError(%q) = %q, want prefix %q", tt.err, err.Error(), tt.wantPfx)
			}
		})
	}
}
