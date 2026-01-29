package dokploy

import (
	"encoding/json"
	"fmt"
)

type Compose struct {
	ID                string   `json:"composeId"`
	Name              string   `json:"name"`
	ProjectID         string   `json:"projectId"`
	EnvironmentID     string   `json:"environmentId"`
	ComposeFile       string   `json:"composeFile"`
	SourceType        string   `json:"sourceType"`
	CustomGitUrl      string   `json:"customGitUrl"`
	CustomGitBranch   string   `json:"customGitBranch"`
	CustomGitSSHKeyId string   `json:"customGitSSHKeyId"`
	ComposePath       string   `json:"composePath"`
	AutoDeploy        bool     `json:"autoDeploy"`
	ServerID          string   `json:"serverId,omitempty"`
	Domains           []Domain `json:"domains"`
}

func (c *Client) CreateCompose(comp Compose) (*Compose, error) {
	// 1. Create minimal compose
	payload := map[string]string{
		"environmentId": comp.EnvironmentID,
		"name":          comp.Name,
		"composeType":   "docker-compose",
		"appName":       comp.Name,
	}

	if comp.ServerID != "" {
		payload["serverId"] = comp.ServerID
	}

	// If raw content provided, include it
	if comp.ComposeFile != "" {
		payload["composeFile"] = comp.ComposeFile
	}

	resp, err := c.doRequest("POST", "compose.create", payload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Compose Compose `json:"compose"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Compose.ID != "" {
		return &wrapper.Compose, nil
	}

	createdComp := wrapper.Compose
	if createdComp.ID == "" {
		if err := json.Unmarshal(resp, &createdComp); err != nil {
			return nil, err
		}
	}

	// 2. Update with Git configuration if necessary
	updatePayload := map[string]interface{}{
		"composeId":  createdComp.ID,
		"name":       comp.Name,
		"sourceType": comp.SourceType,
		"autoDeploy": comp.AutoDeploy,
	}

	if comp.CustomGitUrl != "" {
		updatePayload["customGitUrl"] = comp.CustomGitUrl
	}
	if comp.CustomGitBranch != "" {
		updatePayload["customGitBranch"] = comp.CustomGitBranch
	}
	if comp.CustomGitSSHKeyId != "" {
		updatePayload["customGitSSHKeyId"] = comp.CustomGitSSHKeyId
	}
	if comp.ComposePath != "" {
		updatePayload["composePath"] = comp.ComposePath
	}
	if comp.ComposeFile != "" {
		updatePayload["composeFile"] = comp.ComposeFile
	}

	if comp.ServerID != "" {
		updatePayload["serverId"] = comp.ServerID
	}

	if comp.SourceType == "" {
		if comp.CustomGitUrl != "" {
			updatePayload["sourceType"] = "git"
		} else if comp.ComposeFile != "" {
			updatePayload["sourceType"] = "raw"
		} else {
			updatePayload["sourceType"] = "github"
		}
	}

	respUpdate, err := c.doRequest("POST", "compose.update", updatePayload)
	if err != nil {
		return nil, fmt.Errorf("created compose %s but failed to update config: %w", createdComp.ID, err)
	}

	if string(respUpdate) == "true" {
		return c.GetCompose(createdComp.ID)
	}

	var updateResult Compose
	if err := json.Unmarshal(respUpdate, &wrapper); err == nil && wrapper.Compose.ID != "" {
		return &wrapper.Compose, nil
	}
	if err := json.Unmarshal(respUpdate, &updateResult); err == nil {
		return &updateResult, nil
	}

	return &createdComp, nil
}

func (c *Client) GetCompose(id string) (*Compose, error) {
	endpoint := fmt.Sprintf("compose.one?composeId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var result Compose
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateCompose(comp Compose) (*Compose, error) {
	payload := map[string]interface{}{
		"composeId":  comp.ID,
		"name":       comp.Name,
		"sourceType": comp.SourceType,
		"autoDeploy": comp.AutoDeploy,
	}

	if comp.CustomGitUrl != "" {
		payload["customGitUrl"] = comp.CustomGitUrl
	}
	if comp.CustomGitBranch != "" {
		payload["customGitBranch"] = comp.CustomGitBranch
	}
	if comp.CustomGitSSHKeyId != "" {
		payload["customGitSSHKeyId"] = comp.CustomGitSSHKeyId
	}
	if comp.ComposePath != "" {
		payload["composePath"] = comp.ComposePath
	}
	if comp.ComposeFile != "" {
		payload["composeFile"] = comp.ComposeFile
	}

	if comp.EnvironmentID != "" {
		payload["environmentId"] = comp.EnvironmentID
	}
	if comp.ServerID != "" {
		payload["serverId"] = comp.ServerID
	}

	resp, err := c.doRequest("POST", "compose.update", payload)
	if err != nil {
		return nil, err
	}

	var result Compose
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteCompose(id string) error {
	payload := map[string]string{
		"composeId": id,
	}
	_, err := c.doRequest("POST", "compose.remove", payload)
	return err
}

func (c *Client) DeployCompose(id string) error {
	payload := map[string]string{
		"composeId": id,
	}
	_, err := c.doRequest("POST", "compose.deploy", payload)
	return err
}
