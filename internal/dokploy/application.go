package dokploy

import (
	"encoding/json"
	"fmt"
)

type Application struct {
	ID                string   `json:"applicationId"`
	Name              string   `json:"name"`
	ProjectID         string   `json:"projectId"`
	EnvironmentID     string   `json:"environmentId"`
	RepositoryURL     string   `json:"repository"`
	Branch            string   `json:"branch"`
	BuildType         string   `json:"buildType"`
	DockerfilePath    string   `json:"dockerfile"`
	DockerContextPath string   `json:"dockerContextPath"`
	DockerBuildStage  string   `json:"dockerBuildStage"`
	Env               string   `json:"env"`
	Domains           []Domain `json:"domains"`
	AutoDeploy        bool     `json:"autoDeploy"`
	ServerID          string   `json:"serverId,omitempty"`
	// Enhanced fields
	SourceType         string `json:"sourceType"`
	CustomGitUrl       string `json:"customGitUrl"`
	CustomGitBranch    string `json:"customGitBranch"`
	CustomGitSSHKeyId  string `json:"customGitSSHKeyId"`
	CustomGitBuildPath string `json:"customGitBuildPath"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	// GitHub Provider fields
	GithubRepository string   `json:"githubRepository"`
	GithubOwner      string   `json:"owner"`
	GithubBranch     string   `json:"githubBranch"`
	GithubBuildPath  string   `json:"buildPath"`
	GithubID         string   `json:"githubId"`
	GithubWatchPaths []string `json:"watchPaths"`
	EnableSubmodules bool     `json:"enableSubmodules"`
	TriggerType      string   `json:"triggerType"`
}

func (c *Client) CreateApplication(app Application) (*Application, error) {
	// 1. Create minimal application
	createPayload := map[string]string{
		"name":          app.Name,
		"environmentId": app.EnvironmentID,
	}

	if app.ServerID != "" {
		createPayload["serverId"] = app.ServerID
	}

	resp, err := c.doRequest("POST", "application.create", createPayload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Application Application `json:"application"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, err
	}

	createdApp := wrapper.Application
	if createdApp.ID == "" {
		if err := json.Unmarshal(resp, &createdApp); err != nil {
			return nil, err
		}
	}

	// 2. Update with full configuration
	updatePayload := map[string]interface{}{
		"applicationId": createdApp.ID,
		"name":          app.Name,
		"branch":        app.Branch,
		"buildType":     app.BuildType,
		"sourceType":    app.SourceType,
		"autoDeploy":    app.AutoDeploy,
	}

	if app.RepositoryURL != "" {
		updatePayload["repository"] = app.RepositoryURL
	}
	if app.DockerfilePath != "" {
		updatePayload["dockerfile"] = app.DockerfilePath
	}
	if app.DockerContextPath != "" {
		updatePayload["dockerContextPath"] = app.DockerContextPath
	}
	if app.DockerBuildStage != "" {
		updatePayload["dockerBuildStage"] = app.DockerBuildStage
	}
	if app.CustomGitUrl != "" {
		updatePayload["customGitUrl"] = app.CustomGitUrl
	}
	if app.CustomGitBranch != "" {
		updatePayload["customGitBranch"] = app.CustomGitBranch
	}
	if app.CustomGitSSHKeyId != "" {
		updatePayload["customGitSSHKeyId"] = app.CustomGitSSHKeyId
	}
	if app.CustomGitBuildPath != "" {
		updatePayload["customGitBuildPath"] = app.CustomGitBuildPath
	}
	if app.Username != "" {
		updatePayload["username"] = app.Username
	}
	if app.Password != "" {
		updatePayload["password"] = app.Password
	}

	// Ensure defaults
	if app.SourceType == "" {
		if app.CustomGitUrl != "" {
			updatePayload["sourceType"] = "git"
		} else {
			updatePayload["sourceType"] = "github"
		}
	}

	if app.ServerID != "" {
		updatePayload["serverId"] = app.ServerID
	}

	respUpdate, err := c.doRequest("POST", "application.update", updatePayload)
	if err != nil {
		return nil, fmt.Errorf("created application %s but failed to update config: %w", createdApp.ID, err)
	}

	if string(respUpdate) == "true" {
		return c.GetApplication(createdApp.ID)
	}

	var updateResult Application
	if err := json.Unmarshal(respUpdate, &wrapper); err == nil && wrapper.Application.ID != "" {
		return &wrapper.Application, nil
	}
	if err := json.Unmarshal(respUpdate, &updateResult); err == nil {
		return &updateResult, nil
	}

	return &createdApp, nil
}

func (c *Client) GetApplication(id string) (*Application, error) {
	endpoint := fmt.Sprintf("application.one?applicationId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result Application
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateApplication(app Application) (*Application, error) {
	payload := map[string]interface{}{
		"applicationId": app.ID,
		"name":          app.Name,
		"branch":        app.Branch,
		"buildType":     app.BuildType,
		"sourceType":    app.SourceType,
		"autoDeploy":    app.AutoDeploy,
	}
	// Optional fields
	if app.RepositoryURL != "" {
		payload["repository"] = app.RepositoryURL
	}
	if app.DockerfilePath != "" {
		payload["dockerfile"] = app.DockerfilePath
	}
	if app.DockerContextPath != "" {
		payload["dockerContextPath"] = app.DockerContextPath
	}
	if app.DockerBuildStage != "" {
		payload["dockerBuildStage"] = app.DockerBuildStage
	}
	if app.CustomGitUrl != "" {
		payload["customGitUrl"] = app.CustomGitUrl
	}
	if app.CustomGitBranch != "" {
		payload["customGitBranch"] = app.CustomGitBranch
	}
	if app.CustomGitSSHKeyId != "" {
		payload["customGitSSHKeyId"] = app.CustomGitSSHKeyId
	}
	if app.CustomGitBuildPath != "" {
		payload["customGitBuildPath"] = app.CustomGitBuildPath
	}
	if app.Username != "" {
		payload["username"] = app.Username
	}
	if app.Password != "" {
		payload["password"] = app.Password
	}

	if app.EnvironmentID != "" {
		payload["environmentId"] = app.EnvironmentID
	}
	if app.ServerID != "" {
		payload["serverId"] = app.ServerID
	}

	resp, err := c.doRequest("POST", "application.update", payload)
	if err != nil {
		return nil, err
	}

	var result Application
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteApplication(id string) error {
	payload := map[string]string{
		"applicationId": id,
	}
	_, err := c.doRequest("POST", "application.remove", payload)
	return err
}

func (c *Client) SaveGithubProvider(appID string, githubConfig map[string]interface{}) error {
	payload := map[string]interface{}{
		"applicationId": appID,
	}

	// Add all github provider fields
	for key, value := range githubConfig {
		payload[key] = value
	}

	_, err := c.doRequest("POST", "application.saveGithubProvider", payload)
	return err
}

func (c *Client) DeployApplication(id string) error {
	payload := map[string]string{
		"applicationId": id,
	}
	_, err := c.doRequest("POST", "application.deploy", payload)
	return err
}
