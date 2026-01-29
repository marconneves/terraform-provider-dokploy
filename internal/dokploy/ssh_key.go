package dokploy

import (
	"encoding/json"
	"fmt"
)

type SSHKey struct {
	ID          string `json:"sshKeyId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PrivateKey  string `json:"privateKey"`
	PublicKey   string `json:"publicKey"`
}

func (c *Client) CreateSSHKey(name, description, privateKey, publicKey string) (*SSHKey, error) {
	// Fetch user to get Organization ID
	user, err := c.GetUser()
	if err != nil {
		return nil, fmt.Errorf("failed to get user for organization ID: %w", err)
	}

	payload := map[string]string{
		"name":           name,
		"description":    description,
		"privateKey":     privateKey,
		"publicKey":      publicKey,
		"organizationId": user.OrganizationID,
	}

	resp, err := c.doRequest("POST", "sshKey.create", payload)
	if err != nil {
		return nil, err
	}

	// Handle empty response or boolean by fetching list
	if len(resp) == 0 || string(resp) == "true" {
		return c.findSSHKeyByName(name)
	}

	var wrapper struct {
		SSHKey SSHKey `json:"sshKey"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.SSHKey.ID != "" {
		return &wrapper.SSHKey, nil
	}

	var result SSHKey
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return c.findSSHKeyByName(name)
	}

	// Fallback to list lookup if unmarshal failed to produce ID
	return &result, nil
}

func (c *Client) ListSSHKeys() ([]SSHKey, error) {
	resp, err := c.doRequest("GET", "sshKey.all", nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		SSHKeys []SSHKey `json:"sshKeys"` // Guessing wrapper
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.SSHKeys != nil {
		return wrapper.SSHKeys, nil
	}

	var list []SSHKey
	if err := json.Unmarshal(resp, &list); err == nil {
		return list, nil
	}

	return nil, fmt.Errorf("failed to parse sshKey.all response")
}

func (c *Client) findSSHKeyByName(name string) (*SSHKey, error) {
	keys, err := c.ListSSHKeys()
	if err != nil {
		return nil, fmt.Errorf("ssh key created but failed to list keys: %w", err)
	}
	for _, key := range keys {
		if key.Name == name {
			return &key, nil
		}
	}
	return nil, fmt.Errorf("ssh key created but not found in list by name: %s", name)
}

func (c *Client) GetSSHKey(id string) (*SSHKey, error) {
	endpoint := fmt.Sprintf("sshKey.one?sshKeyId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var result SSHKey
	if err := json.Unmarshal(resp, &result); err != nil {
		// Try wrapper?
		var wrapper struct {
			SSHKey SSHKey `json:"sshKey"`
		}
		if err2 := json.Unmarshal(resp, &wrapper); err2 == nil {
			return &wrapper.SSHKey, nil
		}
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteSSHKey(id string) error {
	payload := map[string]string{
		"sshKeyId": id,
	}
	_, err := c.doRequest("POST", "sshKey.remove", payload)
	return err
}
