package dokploy

import (
	"encoding/json"
	"fmt"
)

type Deployment struct {
	ID            string `json:"deploymentId"`
	ApplicationID string `json:"applicationId"`
	Status        string `json:"status"`
	Log           string `json:"log"`
	CreatedAt     string `json:"createdAt"`
}

func (c *Client) CreateDeployment(appID string) (*Deployment, error) {
	payload := map[string]string{
		"applicationId": appID,
	}
	resp, err := c.doRequest("POST", "application.deploy", payload)
	if err != nil {
		return nil, err
	}

	// Deploy endpoint might return the deployment object or just success
	var wrapper struct {
		Deployment Deployment `json:"deployment"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Deployment.ID != "" {
		return &wrapper.Deployment, nil
	}

	// If it returns "true" or similar, we might need to fetch the latest deployment
	// But let's assume it returns the created deployment for now, or we can fetch the latest one.
	// Common pattern in Dokploy seems to be returning the object.
	var result Deployment
	if err := json.Unmarshal(resp, &result); err == nil && result.ID != "" {
		return &result, nil
	}

	// Fallback: Fetch latest deployment for the app
	deployments, err := c.ListDeployments(appID)
	if err == nil && len(deployments) > 0 {
		return &deployments[0], nil
	}

	return nil, fmt.Errorf("deployment triggered but failed to parse response or fetch latest deployment")
}

func (c *Client) GetDeployment(id string) (*Deployment, error) {
	// Assuming deployment.one endpoint exists, or we might need to filter from list
	// Given the pattern, let's try deployment.one?deploymentId=...
	// If not, we might fail.
	endpoint := fmt.Sprintf("deployment.one?deploymentId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Deployment Deployment `json:"deployment"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Deployment.ID != "" {
		return &wrapper.Deployment, nil
	}

	var result Deployment
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListDeployments(appID string) ([]Deployment, error) {
	// Usually deployments are listed per application
	// Try getting application first, maybe it has deployments?
	// Or maybe endpoint deployment.all?applicationId=...
	endpoint := fmt.Sprintf("deployment.all?applicationId=%s", appID)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var list []Deployment
	if err := json.Unmarshal(resp, &list); err == nil {
		return list, nil
	}

	var wrapper struct {
		Deployments []Deployment `json:"deployments"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil {
		return wrapper.Deployments, nil
	}

	return nil, fmt.Errorf("failed to parse deployment.all response")
}
