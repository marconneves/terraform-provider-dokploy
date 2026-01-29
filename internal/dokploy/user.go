package dokploy

import (
	"encoding/json"
	"fmt"
)

type User struct {
	ID             string `json:"userId"`
	Email          string `json:"email"`
	OrganizationID string `json:"organizationId"`
	Password       string `json:"password,omitempty"`
	Role           string `json:"role,omitempty"`
}

// GetUser returns the current authenticated user.
func (c *Client) GetUser() (*User, error) {
	resp, err := c.doRequest("GET", "user.get", nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		User User `json:"user"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.User.ID != "" {
		return &wrapper.User, nil
	}

	var user User
	if err := json.Unmarshal(resp, &user); err == nil && user.ID != "" {
		return &user, nil
	}

	return nil, fmt.Errorf("failed to parse user response")
}

func (c *Client) CreateUser(email, password, role string) (*User, error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
		"role":     role,
	}
	resp, err := c.doRequest("POST", "user.create", payload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		User User `json:"user"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.User.ID != "" {
		return &wrapper.User, nil
	}

	var user User
	if err := json.Unmarshal(resp, &user); err == nil {
		return &user, nil
	}
	return nil, fmt.Errorf("failed to parse create user response")
}

func (c *Client) GetUserByID(id string) (*User, error) {
	endpoint := fmt.Sprintf("user.one?userId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		User User `json:"user"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.User.ID != "" {
		return &wrapper.User, nil
	}

	var user User
	if err := json.Unmarshal(resp, &user); err == nil {
		return &user, nil
	}
	return nil, fmt.Errorf("failed to parse get user response")
}

func (c *Client) ListUsers() ([]User, error) {
	resp, err := c.doRequest("GET", "user.all", nil)
	if err != nil {
		return nil, err
	}

	var list []User
	if err := json.Unmarshal(resp, &list); err == nil {
		return list, nil
	}

	var wrapper struct {
		Users []User `json:"users"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil {
		return wrapper.Users, nil
	}

	return nil, fmt.Errorf("failed to parse user.all response")
}

func (c *Client) DeleteUser(id string) error {
	payload := map[string]string{
		"userId": id,
	}
	_, err := c.doRequest("POST", "user.remove", payload)
	return err
}
