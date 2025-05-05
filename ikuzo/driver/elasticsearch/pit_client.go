package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type PITClient struct {
	URLs            []string
	Username        string
	Password        string
	HTTPClient      *http.Client
	currentURLIndex int
	mu              sync.RWMutex
}

type PITResponse struct {
	ID string `json:"id"`
}

type SearchResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			ID     string                 `json:"_id"`
			Score  float64                `json:"_score"`
			Source map[string]interface{} `json:"_source"`
			Sort   []interface{}          `json:"sort"`
		} `json:"hits"`
	} `json:"hits"`
}

func NewPITClient(urls []string, username, password string) (*PITClient, error) {
	if len(urls) == 0 {
		return nil, errors.New("at least one URL must be provided")
	}

	cleanURLs := make([]string, len(urls))
	for i, url := range urls {
		cleanURLs[i] = strings.TrimSuffix(url, "/")
	}

	return &PITClient{
		URLs:       cleanURLs,
		Username:   username,
		Password:   password,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *PITClient) getCurrentURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.URLs[c.currentURLIndex]
}

func (c *PITClient) rotateURL() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentURLIndex = (c.currentURLIndex + 1) % len(c.URLs)
	return c.URLs[c.currentURLIndex]
}

func (c *PITClient) executeRequest(req *http.Request) (*http.Response, error) {
	var lastErr error
	tried := make(map[string]bool)

	for i := 0; i < len(c.URLs); i++ {
		currentURL := c.getCurrentURL()
		if tried[currentURL] {
			c.rotateURL()
			continue
		}

		reqURL := strings.Replace(req.URL.String(), req.URL.Scheme+"://"+req.URL.Host, currentURL, 1)
		newReq, err := http.NewRequestWithContext(req.Context(), req.Method, reqURL, req.Body)
		if err != nil {
			lastErr = err
			c.rotateURL()
			tried[currentURL] = true
			continue
		}
		newReq.Header = req.Header
		if c.Username != "" {
			newReq.SetBasicAuth(c.Username, c.Password)
		}

		resp, err := c.HTTPClient.Do(newReq)
		if err != nil {
			lastErr = err
			c.rotateURL()
			tried[currentURL] = true
			continue
		}

		return resp, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all servers failed, last error: %w", lastErr)
	}
	return nil, errors.New("all servers failed")
}

func (c *PITClient) CreatePIT(ctx context.Context, indexName string, keepAlive string) (string, error) {
	if indexName == "" {
		return "", fmt.Errorf("index name is required for Elasticsearch PIT creation")
	}

	urlPath := fmt.Sprintf("/%s/_pit?keep_alive=%s", indexName, keepAlive)
	baseURL := c.getCurrentURL()
	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+urlPath, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Username != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.executeRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create PIT: %s - %s", resp.Status, string(body))
	}

	var pitResp PITResponse
	if err := json.NewDecoder(resp.Body).Decode(&pitResp); err != nil {
		return "", err
	}
	return pitResp.ID, nil
}

func (c *PITClient) SearchWithPIT(
	ctx context.Context,
	pitID string,
	keepAlive string,
	query map[string]interface{},
	size int,
	sortFields []map[string]string,
	searchAfter []interface{},
) (*SearchResponse, error) {
	urlPath := "/_search"

	requestBody := map[string]interface{}{
		"size":  size,
		"query": query,
		"pit": map[string]interface{}{
			"id":         pitID,
			"keep_alive": keepAlive,
		},
		"sort": sortFields,
	}

	if len(searchAfter) > 0 {
		requestBody["search_after"] = searchAfter
	}

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	baseURL := c.getCurrentURL()
	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+urlPath, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Username != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.executeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: %s - %s", resp.Status, string(body))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	return &searchResp, nil
}

func (c *PITClient) ClosePIT(ctx context.Context, pitID string) error {
	urlPath := "/_pit"
	requestBody := map[string]string{
		"pit_id": pitID,
	}

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	baseURL := c.getCurrentURL()
	req, err := http.NewRequestWithContext(ctx, "DELETE", baseURL+urlPath, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Username != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.executeRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to close PIT: %s - %s", resp.Status, string(body))
	}

	return nil
}
