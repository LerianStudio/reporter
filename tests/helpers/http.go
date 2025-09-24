package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPClient struct {
	base   string
	client *http.Client
}

func NewHTTPClient(base string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{base: base, client: &http.Client{Timeout: timeout}}
}

func (c *HTTPClient) RequestFull(ctx context.Context, method, path string, headers map[string]string, body any) (int, []byte, http.Header, error) {
	var rdr io.Reader

	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, nil, nil, fmt.Errorf("marshal body: %w", err)
		}

		rdr = bytes.NewReader(b)

		if headers == nil {
			headers = map[string]string{}
		}

		if _, ok := headers["Content-Type"]; !ok {
			headers["Content-Type"] = "application/json"
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.base+path, rdr)
	if err != nil {
		return 0, nil, nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, resp.Header, err
	}

	return resp.StatusCode, data, resp.Header, nil
}

func (c *HTTPClient) Request(ctx context.Context, method, path string, headers map[string]string, body any) (int, []byte, error) {
	code, b, _, err := c.RequestFull(ctx, method, path, headers, body)
	return code, b, err
}

func (c *HTTPClient) RequestWithHeaderValues(ctx context.Context, method, path string, headers map[string][]string, body any) (int, []byte, http.Header, error) {
	var rdr io.Reader

	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, nil, nil, fmt.Errorf("marshal body: %w", err)
		}

		rdr = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.base+path, rdr)
	if err != nil {
		return 0, nil, nil, err
	}

	for k, vals := range headers {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, resp.Header, err
	}

	return resp.StatusCode, data, resp.Header, nil
}

func (c *HTTPClient) RequestFullWithRetry(ctx context.Context, method, path string, headers map[string]string, body any, attempts int, baseBackoff time.Duration) (int, []byte, http.Header, error) {
	if attempts <= 0 {
		attempts = 1
	}

	if baseBackoff <= 0 {
		baseBackoff = 200 * time.Millisecond
	}

	var (
		lastCode int
		lastBody []byte
		lastHdr  http.Header
		lastErr  error
	)

	for i := 0; i < attempts; i++ {
		code, b, hdr, err := c.RequestFull(ctx, method, path, headers, body)

		lastCode, lastBody, lastHdr, lastErr = code, b, hdr, err
		if err == nil && code != 429 && code != 502 && code != 503 && code != 504 {
			return code, b, hdr, nil
		}

		time.Sleep(time.Duration(1<<i) * baseBackoff)
	}

	return lastCode, lastBody, lastHdr, lastErr
}
