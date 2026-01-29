package dokploy

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type Organization struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Logo      string `json:"logo"`
	OwnerId   string `json:"ownerId"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"createdAt"`
}

func (c *Client) CreateOrganization(name string, logo string) (*Organization, error) {
	jsonPayload := map[string]interface{}{
		"name":           name,
		"organizationId": "",
	}
	if logo != "" {
		jsonPayload["logo"] = logo
	}

	payload := map[string]interface{}{
		"0": map[string]interface{}{
			"json": jsonPayload,
		},
	}

	resp, err := c.doSessionRequest("POST", "organization.create?batch=1", payload)
	if err != nil {
		return nil, err
	}

	var response []TRPCResponse[Organization]
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("empty response from organization.create")
	}

	return &response[0].Result.Data.Json, nil
}

func (c *Client) GetOrganization(id string) (*Organization, error) {
	input := map[string]interface{}{
		"0": map[string]interface{}{
			"json": map[string]interface{}{
				"organizationId": id,
			},
		},
	}
	jsonBytes, _ := json.Marshal(input)
	encodedInput := url.QueryEscape(string(jsonBytes))

	endpoint := fmt.Sprintf("organization.get?batch=1&input=%s", encodedInput)
	resp, err := c.doSessionRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response []TRPCResponse[Organization]
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, nil
	}

	org := response[0].Result.Data.Json
	if org.ID == "" {
		return nil, nil
	}

	return &org, nil
}

func (c *Client) DeleteOrganization(id string) error {
	payload := map[string]interface{}{
		"0": map[string]interface{}{
			"json": map[string]interface{}{
				"organizationId": id,
			},
		},
	}
	_, err := c.doSessionRequest("POST", "organization.remove?batch=1", payload)
	return err
}
