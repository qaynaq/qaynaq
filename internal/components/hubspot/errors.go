package hubspot

import (
	"fmt"
	"regexp"
	"strings"
)

// statusPrefix matches an already-classified "[NNN] ..." error from doGet so we
// don't double-wrap upstream HTTP status codes.
var statusPrefix = regexp.MustCompile(`^\[\d{3}\] `)

// classifyHubSpotError normalises errors to a "[status] message" form so the
// template's catch processor can extract the status code for MCP responses.
func classifyHubSpotError(err error) error {
	msg := err.Error()

	if statusPrefix.MatchString(msg) {
		return err
	}

	if contains(msg, "is required", "must be set", "failed to interpolate", "invalid", "unsupported action") {
		return fmt.Errorf("[400] %s", msg)
	}
	if contains(msg, "401", "Unauthorized", "authentication") {
		return fmt.Errorf("[401] %s", msg)
	}
	if contains(msg, "404", "Not Found") {
		return fmt.Errorf("[404] %s", msg)
	}
	if contains(msg, "429", "rate limit", "exceeded") {
		return fmt.Errorf("[429] %s", msg)
	}

	return fmt.Errorf("[500] %s", msg)
}

func contains(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
