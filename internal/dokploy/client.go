package dokploy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

// Client holds connection details.
type Client struct {
	BaseURL      string
	APIKey       string
	Email        string
	Password     string
	SessionToken string
	HTTPClient   *http.Client
}

func NewClient(baseURL, apiKey, email, password string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		BaseURL:  baseURL,
		APIKey:   apiKey,
		Email:    email,
		Password: password,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
	}
}

func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	url := fmt.Sprintf("%s/%s", c.BaseURL, endpoint)

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(respBytes))
	}

	return respBytes, nil
}

func (c *Client) SignIn() error {
	if c.Email == "" || c.Password == "" {
		return fmt.Errorf("email and password are required for sign in")
	}

	payload := map[string]string{
		"email":    c.Email,
		"password": c.Password,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var url string
	if strings.HasSuffix(c.BaseURL, "/api") {
		url = fmt.Sprintf("%s/auth/sign-in/email", c.BaseURL)
	} else {
		url = fmt.Sprintf("%s/api/auth/sign-in/email", c.BaseURL)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sign in failed: %s - %s", resp.Status, string(respBytes))
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return fmt.Errorf("failed to parse sign in response: %w", err)
	}

	c.SessionToken = result.Token
	return nil
}

func (c *Client) doSessionRequest(method, endpoint string, body interface{}) ([]byte, error) {
	if c.SessionToken == "" {
		if err := c.SignIn(); err != nil {
			return nil, err
		}
	}

	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	url := fmt.Sprintf("%s/trpc/%s", c.BaseURL, endpoint)

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("TRPC error: %s - %s", resp.Status, string(respBytes))
	}

	return respBytes, nil
}

// TRPCResponse generic wrapper.
type TRPCResponse[T any] struct {
	Result struct {
		Data struct {
			Json T `json:"json"`
		} `json:"data"`
	} `json:"result"`
}
