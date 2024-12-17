package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
}

type Request struct {
	client  *Client
	httpReq *http.Request
}

func NewClient(baseUrl string) *Client {
	return &Client{
		BaseURL: baseUrl,
	}
}

func (c *Client) buildUrl(url ...string) string {
	if len(url) == 0 {
		return c.BaseURL
	}

	return fmt.Sprintf("%s/%s", c.BaseURL, strings.Join(url, "/"))
}

func (c *Client) newRequest(runnerId string, runnerSecret string, organizationId string, method string, url string, body interface{}) (*Request, error) {
	var bodyReader io.Reader

	bodyJson, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}
	bodyReader = bytes.NewReader(bodyJson)

	fmt.Printf("Preparing %s request to %s\n", method, url)
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Runner-Id", runnerId)
	req.Header.Set("X-Runner-Secret", runnerSecret)
	req.Header.Set("X-Organization-Id", organizationId)
	req.Header.Set("Content-Type", "application/json")

	return &Request{
		client:  c,
		httpReq: req,
	}, nil
}

func (req *Request) Do() (int, []byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req.httpReq)
	if err != nil {
		fmt.Println("Request Error:", err)
		return 0, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errMsg := fmt.Sprintf("unexpected status code %d", resp.StatusCode)
		fmt.Println("API ERROR:", errMsg)
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Response body:", string(body))
		return resp.StatusCode, nil, fmt.Errorf(errMsg)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return resp.StatusCode, respBody, nil
}
