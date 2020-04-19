package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// HealthcheckResult is a result of single healthcheck
type HealthcheckResult struct {
	Healthcheck *Healthcheck
	Time        time.Time
	OK          bool
	Message     string
}

func (r HealthcheckResult) String() string {
	if r.OK {
		return fmt.Sprintf("%s up", r.Healthcheck.Name)
	}

	return fmt.Sprintf("%s down: %s", r.Healthcheck.Name, r.Message)
}

// NewOKResult creates new successful HealthcheckResult
func NewOKResult(h *Healthcheck) *HealthcheckResult {
	return &HealthcheckResult{
		Healthcheck: h,
		Time:        time.Now().UTC(),
		OK:          true,
		Message:     "",
	}
}

// NewErrorResult creates new unsuccessful HealthcheckResult
func NewErrorResult(h *Healthcheck, str string) *HealthcheckResult {
	parts := strings.Split(str, ":")
	for i := len(parts) - 1; i >= 0; i-- {
		s := parts[i]
		s = strings.TrimSpace(s)
		if s != "" {
			str = s
			break
		}
	}

	i := strings.LastIndex(str, ":")
	if i >= 0 {
		str = str[i:]
	}

	return &HealthcheckResult{
		Healthcheck: h,
		Time:        time.Now().UTC(),
		OK:          false,
		Message:     str,
	}
}

// ExecuteHealthcheck checks specified target via HTTP
func ExecuteHealthcheck(h *Healthcheck) *HealthcheckResult {
	client := &http.Client{}
	client.Timeout = 30 * time.Second

	req := &http.Request{
		Method: http.MethodGet,
		URL:    h.URL,
	}

	resp, err := client.Do(req)
	if err != nil {
		return NewErrorResult(h, err.Error())
	}

	if resp.StatusCode >= 400 {
		return NewErrorResult(h, fmt.Sprintf("non-successful response %d %s", resp.StatusCode, resp.Status))
	}

	return NewOKResult(h)
}
