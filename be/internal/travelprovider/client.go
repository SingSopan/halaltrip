package travelprovider

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const maxResponseBytes = 2 << 20

var ErrTimeout = errors.New("travel recommendation provider timeout")
var ErrUnavailable = errors.New("travel recommendation provider unavailable")
var ErrInvalidResponse = errors.New("travel recommendation provider returned an invalid response")

type HTTPError struct{ StatusCode int }

func (e *HTTPError) Error() string {
	return fmt.Sprintf("travel recommendation provider status %d", e.StatusCode)
}

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: &http.Client{Timeout: timeout}}
}

// Do forwards one request and returns a bounded body. It never exposes provider
// response details through its errors.
func (c *Client) Do(ctx context.Context, method, path string, body []byte) (int, []byte, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return 0, nil, ErrUnavailable
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.http.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			return 0, nil, ErrTimeout
		}
		return 0, nil, ErrUnavailable
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes+1))
	if err != nil || len(bodyBytes) > maxResponseBytes {
		return 0, nil, ErrInvalidResponse
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return res.StatusCode, nil, &HTTPError{StatusCode: res.StatusCode}
	}
	if res.StatusCode != http.StatusNoContent && len(bytes.TrimSpace(bodyBytes)) == 0 {
		return res.StatusCode, nil, ErrInvalidResponse
	}
	return res.StatusCode, bodyBytes, nil
}
