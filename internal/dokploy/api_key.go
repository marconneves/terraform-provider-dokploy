package dokploy

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type ApiKey struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	Prefix      string `json:"prefix"`
	Description string `json:"description"`
	UserId      string `json:"userId"`
	CreatedAt   string `json:"createdAt"`
}

func (c *Client) CreateApiKey(name, description, organizationID string) (*ApiKey, error) {
	// payload: {"0":{"json":{"name":"...","expiresIn":null,"prefix":"...","metadata":{"organizationId":"..."},"rateLimitEnabled":false...}}}
	payload := map[string]interface{}{
		"0": map[string]interface{}{
			"json": map[string]interface{}{
				"name":        name,
				"description": description,
				"expiresIn":   nil,
				"prefix":      "dk", // default prefix?
				"metadata": map[string]string{
					"organizationId": organizationID,
				},
				"rateLimitEnabled": false,
			},
			"meta": map[string]interface{}{
				"values": map[string]interface{}{
					"expiresIn": []string{"undefined"},
				},
			},
		},
	}

	resp, err := c.doSessionRequest("POST", "user.createApiKey?batch=1", payload)
	if err != nil {
		return nil, err
	}

	var response []TRPCResponse[ApiKey]
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("empty response from user.createApiKey")
	}

	return &response[0].Result.Data.Json, nil
}

func (c *Client) DeleteApiKey(id string) error {
	payload := map[string]interface{}{
		"0": map[string]interface{}{
			"json": map[string]interface{}{
				"id": id,
			},
		},
	}
	_, err := c.doSessionRequest("POST", "user.removeApiKey?batch=1", payload)
	return err
}

func (c *Client) GetApiKey(id string) (*ApiKey, error) {
	input := map[string]interface{}{
		"0": map[string]interface{}{
			"json": nil,
		},
	}

	jsonBytes, _ := json.Marshal(input)
	encodedInput := url.QueryEscape(string(jsonBytes))

	endpoint := fmt.Sprintf("user.getApiKeys?batch=1&input=%s", encodedInput)
	resp, err := c.doSessionRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response []TRPCResponse[[]ApiKey]
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("empty response from user.getApiKeys")
	}

	keys := response[0].Result.Data.Json
	for _, key := range keys {
		if key.ID == id {
			return &key, nil
		}
	}

	return nil, nil
}
