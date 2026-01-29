package dokploy

import (
	"encoding/json"
	"fmt"
)

type Server struct {
	ID          string `json:"serverId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IPAddress   string `json:"ipAddress"`
	Port        int64  `json:"port"`
	Username    string `json:"username"`
	SSHKeyID    string `json:"sshKeyId"`
}

func (c *Client) CreateServer(server Server) (*Server, error) {
	payload := map[string]interface{}{
		"name":        server.Name,
		"description": server.Description,
		"ipAddress":   server.IPAddress,
		"port":        server.Port,
		"username":    server.Username,
		"sshKeyId":    server.SSHKeyID,
	}
	resp, err := c.doRequest("POST", "server.create", payload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Server Server `json:"server"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Server.ID != "" {
		return &wrapper.Server, nil
	}

	var result Server
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetServer(id string) (*Server, error) {
	endpoint := fmt.Sprintf("server.one?serverId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Server Server `json:"server"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Server.ID != "" {
		return &wrapper.Server, nil
	}

	var result Server
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListServers() ([]Server, error) {
	resp, err := c.doRequest("GET", "server.all", nil)
	if err != nil {
		return nil, err
	}

	var list []Server
	if err := json.Unmarshal(resp, &list); err == nil {
		return list, nil
	}

	// Maybe wrapped?
	var wrapper struct {
		Servers []Server `json:"servers"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil {
		return wrapper.Servers, nil
	}

	return nil, fmt.Errorf("failed to parse server.all response")
}

func (c *Client) UpdateServer(server Server) (*Server, error) {
	payload := map[string]interface{}{
		"serverId":    server.ID,
		"name":        server.Name,
		"description": server.Description,
		"ipAddress":   server.IPAddress,
		"port":        server.Port,
		"username":    server.Username,
		"sshKeyId":    server.SSHKeyID,
	}

	resp, err := c.doRequest("POST", "server.update", payload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Server Server `json:"server"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Server.ID != "" {
		return &wrapper.Server, nil
	}

	var result Server
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteServer(id string) error {
	payload := map[string]string{
		"serverId": id,
	}
	_, err := c.doRequest("POST", "server.remove", payload)
	return err
}
