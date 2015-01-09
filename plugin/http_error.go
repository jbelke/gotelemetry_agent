// Package plugin contains all the plugins known to the Telemetry agent.
//
// Currently, the following are implemented:
//
// * com.telemetryapp.intercom
// * com.telemetryapp.sql
// * com.telemetryapp.process
package plugin

import (
	"fmt"
	"net/http"
)

type HTTPError struct {
	StatusCode int
	Status     string
}

func NewHTTPError(res *http.Response) error {
	if res.StatusCode > 399 {
		return &HTTPError{
			StatusCode: res.StatusCode,
			Status:     res.Status,
		}
	}

	return nil
}

func (h *HTTPError) Error() string {
	return fmt.Sprintf("HTTP Error %d: %s", h.StatusCode, h.Status)
}
