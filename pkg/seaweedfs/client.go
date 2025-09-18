package seaweedfs

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SeaweedFSClient provides direct HTTP access to SeaweedFS
type SeaweedFSClient struct {
	baseURL     string
	httpClient  *http.Client
	jwtReadKey  string
	jwtWriteKey string
}

// NewSeaweedFSClient creates a new simple HTTP client for SeaweedFS
// jwtReadKey and jwtWriteKey are optional. If provided, the client will
// attach Authorization: Bearer <jwt> on GET/HEAD (read) and PUT/DELETE (write).
func NewSeaweedFSClient(baseURL string, jwtReadKey string, jwtWriteKey string) *SeaweedFSClient {
	return &SeaweedFSClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		jwtReadKey:  jwtReadKey,
		jwtWriteKey: jwtWriteKey,
	}
}

// UploadFile uploads a file to SeaweedFS
func (c *SeaweedFSClient) UploadFile(ctx context.Context, path string, data []byte) error {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	c.attachJWT(req, true)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DownloadFile downloads a file from SeaweedFS
func (c *SeaweedFSClient) DownloadFile(ctx context.Context, path string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.attachJWT(req, false)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return data, nil
}

// attachJWT adds Authorization header when jwt keys are configured.
func (c *SeaweedFSClient) attachJWT(req *http.Request, isWrite bool) {
	var secret string
	if isWrite {
		secret = c.jwtWriteKey
	} else {
		secret = c.jwtReadKey
	}
	if secret == "" {
		return
	}

	token, err := generateHS256JWT(secret, isWrite)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
}

// generateHS256JWT creates a short-lived JWT token with minimal claims.
// It uses HS256 with the provided secret.
func generateHS256JWT(secret string, isWrite bool) (string, error) {
	header := `{"alg":"HS256","typ":"JWT"}`
	now := time.Now().Unix()
	exp := time.Now().Add(1 * time.Minute).Unix()
	scope := "read"
	if isWrite {
		scope = "write"
	}
	payload := fmt.Sprintf(`{"iat":%d,"exp":%d,"scope":"%s"}`, now, exp, scope)

	unsigned := base64URLEncode([]byte(header)) + "." + base64URLEncode([]byte(payload))
	sig := hmacSHA256(unsigned, secret)
	return unsigned + "." + base64URLEncode(sig), nil
}

func base64URLEncode(data []byte) string {
	// standard library without padding
	return base64.RawURLEncoding.EncodeToString(data)
}

func hmacSHA256(message string, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(message))
	return mac.Sum(nil)
}
func (c *SeaweedFSClient) DeleteFile(ctx context.Context, path string) error {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	c.attachJWT(req, true)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// HealthCheck checks if SeaweedFS is accessible
func (c *SeaweedFSClient) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/status", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}
