package google_drive

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/api/googleapi"
)

func classifyError(err error) error {
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		return fmt.Errorf("[%d] %s", apiErr.Code, err.Error())
	}

	msg := err.Error()
	if strings.Contains(msg, "is required") ||
		strings.Contains(msg, "must be set") ||
		strings.Contains(msg, "failed to parse") ||
		strings.Contains(msg, "failed to interpolate") ||
		strings.Contains(msg, "invalid") {
		return fmt.Errorf("[400] %s", msg)
	}

	return fmt.Errorf("[500] %s", msg)
}
