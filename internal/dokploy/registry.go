package dokploy

import (
	"encoding/json"
	"fmt"
)

type Registry struct {
	ID            string `json:"registryId"`
	Name          string `json:"name"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	ImageRegistry string `json:"imageRegistry"`
	RegistryURL   string `json:"registryUrl,omitempty"`
}

func (c *Client) CreateRegistry(registry Registry) (*Registry, error) {
	payload := map[string]interface{}{
		"name":          registry.Name,
		"username":      registry.Username,
		"password":      registry.Password,
		"imageRegistry": registry.ImageRegistry,
	}
	if registry.RegistryURL != "" {
		payload["registryUrl"] = registry.RegistryURL
	}

	resp, err := c.doRequest("POST", "registry.create", payload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Registry Registry `json:"registry"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Registry.ID != "" {
		return &wrapper.Registry, nil
	}

	var result Registry
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetRegistry(id string) (*Registry, error) {
	endpoint := fmt.Sprintf("registry.one?registryId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Registry Registry `json:"registry"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Registry.ID != "" {
		return &wrapper.Registry, nil
	}

	var result Registry
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListRegistries() ([]Registry, error) {
	resp, err := c.doRequest("GET", "registry.all", nil)
	if err != nil {
		return nil, err
	}

	var list []Registry
	if err := json.Unmarshal(resp, &list); err == nil {
		return list, nil
	}

	var wrapper struct {
		Registries []Registry `json:"registries"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil {
		return wrapper.Registries, nil
	}

	return nil, fmt.Errorf("failed to parse registry.all response")
}

func (c *Client) UpdateRegistry(registry Registry) (*Registry, error) {
	payload := map[string]interface{}{
		"registryId":    registry.ID,
		"name":          registry.Name,
		"username":      registry.Username,
		"password":      registry.Password,
		"imageRegistry": registry.ImageRegistry,
	}
	if registry.RegistryURL != "" {
		payload["registryUrl"] = registry.RegistryURL
	}

	resp, err := c.doRequest("POST", "registry.update", payload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Registry Registry `json:"registry"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Registry.ID != "" {
		return &wrapper.Registry, nil
	}

	var result Registry
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteRegistry(id string) error {
	payload := map[string]string{
		"registryId": id,
	}
	_, err := c.doRequest("POST", "registry.remove", payload)
	return err
}
