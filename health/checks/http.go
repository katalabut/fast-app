package checks

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/katalabut/fast-app/health"
)

// HTTPOptions contains options for HTTP health check
type HTTPOptions struct {
	Timeout        time.Duration
	ExpectedStatus int
	ExpectedBody   string
	Method         string
	Headers        map[string]string
}

// HTTPCheck checks HTTP endpoint health
type HTTPCheck struct {
	name string
	url  string
	opts HTTPOptions
}

// NewHTTPCheck creates a new HTTP health check
func NewHTTPCheck(name, url string) *HTTPCheck {
	return &HTTPCheck{
		name: name,
		url:  url,
		opts: HTTPOptions{
			Timeout:        10 * time.Second,
			ExpectedStatus: http.StatusOK,
			Method:         http.MethodGet,
		},
	}
}

// NewHTTPCheckWithOptions creates a new HTTP health check with options
func NewHTTPCheckWithOptions(name, url string, opts HTTPOptions) *HTTPCheck {
	if opts.Timeout == 0 {
		opts.Timeout = 10 * time.Second
	}
	if opts.ExpectedStatus == 0 {
		opts.ExpectedStatus = http.StatusOK
	}
	if opts.Method == "" {
		opts.Method = http.MethodGet
	}

	return &HTTPCheck{
		name: name,
		url:  url,
		opts: opts,
	}
}

// Name returns the name of the health check
func (h *HTTPCheck) Name() string {
	return h.name
}

// Check executes the HTTP health check
func (h *HTTPCheck) Check(ctx context.Context) health.HealthResult {
	start := time.Now()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: h.opts.Timeout,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, h.opts.Method, h.url, nil)
	if err != nil {
		return health.NewUnhealthyResult("failed to create HTTP request").
			WithDetails("error", err.Error()).
			WithDetails("url", h.url)
	}

	// Add custom headers
	for key, value := range h.opts.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return health.NewUnhealthyResult("HTTP request timeout").
				WithDetails("timeout", h.opts.Timeout.String()).
				WithDetails("url", h.url).
				WithDetails("duration", duration.String()).
				WithDuration(duration)
		}

		return health.NewUnhealthyResult("HTTP request failed").
			WithDetails("error", err.Error()).
			WithDetails("url", h.url).
			WithDetails("duration", duration.String()).
			WithDuration(duration)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != h.opts.ExpectedStatus {
		return health.NewUnhealthyResult("unexpected HTTP status code").
			WithDetails("expected_status", h.opts.ExpectedStatus).
			WithDetails("actual_status", resp.StatusCode).
			WithDetails("url", h.url).
			WithDetails("duration", duration.String()).
			WithDuration(duration)
	}

	// Check response body if expected
	if h.opts.ExpectedBody != "" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return health.NewUnhealthyResult("failed to read HTTP response body").
				WithDetails("error", err.Error()).
				WithDetails("url", h.url).
				WithDetails("duration", duration.String()).
				WithDuration(duration)
		}

		if !strings.Contains(string(body), h.opts.ExpectedBody) {
			return health.NewUnhealthyResult("HTTP response body does not contain expected content").
				WithDetails("expected_body", h.opts.ExpectedBody).
				WithDetails("actual_body", string(body)).
				WithDetails("url", h.url).
				WithDetails("duration", duration.String()).
				WithDuration(duration)
		}
	}

	result := health.NewHealthyResult("HTTP endpoint is healthy").
		WithDetails("status_code", resp.StatusCode).
		WithDetails("url", h.url).
		WithDetails("duration", duration.String()).
		WithDuration(duration)

	// Check if response time is concerning (more than half of timeout)
	if duration > h.opts.Timeout/2 {
		result = health.NewDegradedResult("HTTP endpoint is slow").
			WithDetails("status_code", resp.StatusCode).
			WithDetails("url", h.url).
			WithDetails("duration", duration.String()).
			WithDetails("threshold", (h.opts.Timeout / 2).String()).
			WithDuration(duration)
	}

	return result
}
