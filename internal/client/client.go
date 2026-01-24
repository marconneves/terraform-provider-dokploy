package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DokployClient holds connection details.
type DokployClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewDokployClient(baseURL, apiKey string) *DokployClient {
	return &DokployClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *DokployClient) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
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

	// fmt.Fprintf(os.Stderr, "DEBUG RESPONSE [%s]: %s\n", endpoint, string(respBytes))

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(respBytes))
	}

	return respBytes, nil
}

// --- User ---

type User struct {
	ID             string `json:"userId"`
	Email          string `json:"email"`
	OrganizationID string `json:"organizationId"`
	Password       string `json:"password,omitempty"`
	Role           string `json:"role,omitempty"`
}

// GetUser returns the current authenticated user.
func (c *DokployClient) GetUser() (*User, error) {
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

func (c *DokployClient) CreateUser(email, password, role string) (*User, error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
		"role":     role,
	}
	// Assuming auth.create or user.create
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

func (c *DokployClient) GetUserByID(id string) (*User, error) {
	// Assuming user.one?userId=...
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

func (c *DokployClient) ListUsers() ([]User, error) {
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

func (c *DokployClient) DeleteUser(id string) error {
	payload := map[string]string{
		"userId": id,
	}
	_, err := c.doRequest("POST", "user.remove", payload)
	return err
}

// --- Project ---

type Project struct {
	ID           string        `json:"projectId"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Environments []Environment `json:"environments"`
}

type projectResponse struct {
	Project Project `json:"project"`
}

func (c *DokployClient) CreateProject(name, description, env string) (*Project, error) {
	payload := map[string]string{
		"name":        name,
		"description": description,
		"env":         env,
	}
	resp, err := c.doRequest("POST", "project.create", payload)
	if err != nil {
		return nil, err
	}

	var result projectResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result.Project, nil
}

func (c *DokployClient) GetProject(id string) (*Project, error) {
	endpoint := fmt.Sprintf("project.one?projectId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result Project
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) ListProjects() ([]Project, error) {
	resp, err := c.doRequest("GET", "project.all", nil)
	if err != nil {
		return nil, err
	}

	var list []Project
	if err := json.Unmarshal(resp, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *DokployClient) DeleteProject(id string) error {
	payload := map[string]string{
		"projectId": id,
	}
	_, err := c.doRequest("POST", "project.remove", payload)
	return err
}

func (c *DokployClient) UpdateProject(id, name, description string) (*Project, error) {
	payload := map[string]string{
		"projectId":   id,
		"name":        name,
		"description": description,
	}
	resp, err := c.doRequest("POST", "project.update", payload)
	if err != nil {
		return nil, err
	}

	var result Project
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// --- Environment ---

type Environment struct {
	ID           string        `json:"environmentId"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	ProjectID    string        `json:"projectId"`
	Postgres     []Database    `json:"postgres"`
	Mysql        []Database    `json:"mysql"`
	Mariadb      []Database    `json:"mariadb"`
	Mongo        []Database    `json:"mongo"`
	Redis        []Database    `json:"redis"`
	Applications []Application `json:"applications"`
	Composes     []Compose     `json:"composes"`
}

func (c *DokployClient) GetEnvironment(id string) (*Environment, error) {
	endpoint := fmt.Sprintf("environment.one?environmentId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Environment Environment `json:"environment"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Environment.ID != "" {
		return &wrapper.Environment, nil
	}

	var result Environment
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) CreateEnvironment(projectID, name, description string) (*Environment, error) {
	payload := map[string]string{
		"projectId":   projectID,
		"name":        name,
		"description": description,
	}
	resp, err := c.doRequest("POST", "environment.create", payload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Environment Environment `json:"environment"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Environment.ID != "" {
		return &wrapper.Environment, nil
	}

	var result Environment
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) UpdateEnvironment(env Environment) (*Environment, error) {
	payload := map[string]interface{}{
		"environmentId": env.ID,
		"name":          env.Name,
		"description":   env.Description,
		"projectId":     env.ProjectID,
	}
	resp, err := c.doRequest("POST", "environment.update", payload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Environment Environment `json:"environment"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Environment.ID != "" {
		return &wrapper.Environment, nil
	}

	var result Environment
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) DeleteEnvironment(id string) error {
	payload := map[string]string{
		"environmentId": id,
	}
	_, err := c.doRequest("POST", "environment.remove", payload)
	return err
}

// --- Application ---

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

func (c *DokployClient) CreateApplication(app Application) (*Application, error) {
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

func (c *DokployClient) GetApplication(id string) (*Application, error) {
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

func (c *DokployClient) UpdateApplication(app Application) (*Application, error) {
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

func (c *DokployClient) DeleteApplication(id string) error {
	payload := map[string]string{
		"applicationId": id,
	}
	_, err := c.doRequest("POST", "application.remove", payload)
	return err
}

func (c *DokployClient) SaveGithubProvider(appID string, githubConfig map[string]interface{}) error {
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

func (c *DokployClient) DeployApplication(id string) error {
	payload := map[string]string{
		"applicationId": id,
	}
	_, err := c.doRequest("POST", "application.deploy", payload)
	return err
}

// --- Compose ---

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

func (c *DokployClient) CreateCompose(comp Compose) (*Compose, error) {
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

func (c *DokployClient) GetCompose(id string) (*Compose, error) {
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

func (c *DokployClient) UpdateCompose(comp Compose) (*Compose, error) {
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

func (c *DokployClient) DeleteCompose(id string) error {
	payload := map[string]string{
		"composeId": id,
	}
	_, err := c.doRequest("POST", "compose.remove", payload)
	return err
}

func (c *DokployClient) DeployCompose(id string) error {
	payload := map[string]string{
		"composeId": id,
	}
	_, err := c.doRequest("POST", "compose.deploy", payload)
	return err
}

// --- Database ---

type Database struct {
	ID            string `json:"databaseId"`
	Name          string `json:"name"`
	AppName       string `json:"appName"`
	Type          string `json:"type"`
	ProjectID     string `json:"projectId"`
	EnvironmentID string `json:"environmentId"`
	Version       string `json:"version"`
	DockerImage   string `json:"dockerImage"`
	ExternalPort  int64  `json:"externalPort"`
	InternalPort  int64  `json:"internalPort"`
	Password      string `json:"password"`
	PostgresID    string `json:"postgresId"`
	MysqlID       string `json:"mysqlId"`
	MariadbID     string `json:"mariadbId"`
	MongoID       string `json:"mongoId"`
	RedisID       string `json:"redisId"`
}

func (c *DokployClient) CreateDatabase(projectID, environmentID, name, dbType, password, dockerImage string) (*Database, error) {
	var endpoint string
	payload := map[string]string{
		"environmentId":    environmentID,
		"name":             name,
		"appName":          name,
		"databaseName":     name,
		"databasePassword": password,
		"dockerImage":      dockerImage,
	}

	switch dbType {
	case "postgres":
		endpoint = "postgres.create"
		payload["databaseUser"] = "postgres"
	case "mysql":
		endpoint = "mysql.create"
		payload["databaseUser"] = "root"
	case "mariadb":
		endpoint = "mariadb.create"
		payload["databaseUser"] = "root"
	case "mongo":
		endpoint = "mongo.create"
		payload["databaseUser"] = "mongo"
	case "redis":
		endpoint = "redis.create"
		payload["databaseUser"] = "default"
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	resp, err := c.doRequest("POST", endpoint, payload)
	if err != nil {
		return nil, err
	}

	// Try to parse the response as a database object first
	var directResult Database
	if err := json.Unmarshal(resp, &directResult); err == nil && directResult.PostgresID != "" {
		// Direct response with postgresId
		directResult.ID = directResult.PostgresID
		if directResult.Type == "" {
			directResult.Type = dbType
		}
		return &directResult, nil
	}

	// Check if response is just "true" (boolean success indicator)
	if string(resp) == "true" {
		project, err := c.GetProject(projectID)
		if err != nil {
			return nil, fmt.Errorf("database created but failed to fetch project: %w", err)
		}

		for _, env := range project.Environments {
			if env.ID == environmentID {
				var dbs []Database
				switch dbType {
				case "postgres":
					dbs = env.Postgres
				case "mysql":
					dbs = env.Mysql
				case "mariadb":
					dbs = env.Mariadb
				case "mongo":
					dbs = env.Mongo
				case "redis":
					dbs = env.Redis
				}

				for _, db := range dbs {
					if db.Name == name || db.AppName == name {
						id := db.PostgresID
						if db.MysqlID != "" {
							id = db.MysqlID
						}
						if db.MariadbID != "" {
							id = db.MariadbID
						}
						if db.MongoID != "" {
							id = db.MongoID
						}
						if db.RedisID != "" {
							id = db.RedisID
						}

						// If no type-specific ID, try the generic databaseId field
						if id == "" && db.ID != "" {
							id = db.ID
						}

						// Create a result database with the ID explicitly set
						result := Database{
							ID:            id,
							Name:          db.Name,
							AppName:       db.AppName,
							Type:          dbType,
							ProjectID:     db.ProjectID,
							EnvironmentID: db.EnvironmentID,
							Version:       db.Version,
							DockerImage:   db.DockerImage,
							ExternalPort:  db.ExternalPort,
							InternalPort:  db.InternalPort,
							Password:      db.Password,
							PostgresID:    db.PostgresID,
							MysqlID:       db.MysqlID,
							MariadbID:     db.MariadbID,
							MongoID:       db.MongoID,
							RedisID:       db.RedisID,
						}

						if result.ID == "" {
							return nil, fmt.Errorf("database created but ID not found (name: %s, postgresId: %s, databaseId: %s)", name, db.PostgresID, db.ID)
						}

						return &result, nil
					}
				}
			}
		}
		return nil, fmt.Errorf("database created but not found in project environments")
	}

	var wrapper struct {
		Database Database `json:"database"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Database.ID != "" {
		if wrapper.Database.Type == "" {
			wrapper.Database.Type = dbType
		}
		return &wrapper.Database, nil
	}

	var result Database
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	if result.Type == "" {
		result.Type = dbType
	}
	return &result, nil
}

func (c *DokployClient) GetDatabase(dbID string, databaseType string) (*Database, error) {
	var endpoint string
	switch databaseType {
	case "postgres":
		endpoint = fmt.Sprintf("postgres.one?postgresId=%s", dbID)
	case "mysql":
		endpoint = fmt.Sprintf("mysql.one?mysqlId=%s", dbID)
	case "mariadb":
		endpoint = fmt.Sprintf("mariadb.one?mariadbId=%s", dbID)
	case "mongo":
		endpoint = fmt.Sprintf("mongo.one?mongoId=%s", dbID)
	case "redis":
		endpoint = fmt.Sprintf("redis.one?redisId=%s", dbID)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", databaseType)
	}

	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var db Database
	if err := json.Unmarshal(resp, &db); err == nil {
		valid := false
		if db.ID != "" {
			valid = true
		}
		if db.PostgresID != "" {
			valid = true
		}
		if db.MysqlID != "" {
			valid = true
		}
		if db.MariadbID != "" {
			valid = true
		}
		if db.MongoID != "" {
			valid = true
		}
		if db.RedisID != "" {
			valid = true
		}

		if valid {
			if db.ID == "" {
				if db.PostgresID != "" {
					db.ID = db.PostgresID
				}
				if db.MysqlID != "" {
					db.ID = db.MysqlID
				}
				if db.MariadbID != "" {
					db.ID = db.MariadbID
				}
				if db.MongoID != "" {
					db.ID = db.MongoID
				}
				if db.RedisID != "" {
					db.ID = db.RedisID
				}
			}
			db.Type = databaseType
			return &db, nil
		}
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	var dbBytes json.RawMessage
	var ok bool

	switch databaseType {
	case "postgres":
		dbBytes, ok = result["postgres"]
	case "mysql":
		dbBytes, ok = result["mysql"]
	case "mariadb":
		dbBytes, ok = result["mariadb"]
	case "mongo":
		dbBytes, ok = result["mongo"]
	case "redis":
		dbBytes, ok = result["redis"]
	}

	if !ok {
		if val, found := result["database"]; found {
			dbBytes = val
		} else {
			return nil, fmt.Errorf("database key not found in response for type %s", databaseType)
		}
	}

	if err := json.Unmarshal(dbBytes, &db); err != nil {
		return nil, err
	}

	if db.ID == "" {
		if db.PostgresID != "" {
			db.ID = db.PostgresID
		}
		if db.MysqlID != "" {
			db.ID = db.MysqlID
		}
		if db.MariadbID != "" {
			db.ID = db.MariadbID
		}
		if db.MongoID != "" {
			db.ID = db.MongoID
		}
		if db.RedisID != "" {
			db.ID = db.RedisID
		}
	}
	db.Type = databaseType

	return &db, nil
}

func (c *DokployClient) DeleteDatabase(id string) error {
	return fmt.Errorf("delete database requires type update")
}

func (c *DokployClient) DeleteDatabaseWithType(id, dbType string) error {
	var endpoint string
	var idKey string
	switch dbType {
	case "postgres":
		endpoint = "postgres.remove"
		idKey = "postgresId"
	case "mysql":
		endpoint = "mysql.remove"
		idKey = "mysqlId"
	case "mariadb":
		endpoint = "mariadb.remove"
		idKey = "mariadbId"
	case "mongo":
		endpoint = "mongo.remove"
		idKey = "mongoId"
	case "redis":
		endpoint = "redis.remove"
		idKey = "redisId"
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	payload := map[string]string{
		idKey: id,
	}
	_, err := c.doRequest("POST", endpoint, payload)
	return err
}

// --- Domain ---

type Domain struct {
	ID            string `json:"domainId"`
	ApplicationID string `json:"applicationId"`
	ComposeID     string `json:"composeId"`
	ServiceName   string `json:"serviceName"`
	Host          string `json:"host"`
	Path          string `json:"path"`
	Port          int64  `json:"port"`
	HTTPS         bool   `json:"https"`
}

func (c *DokployClient) CreateDomain(domain Domain) (*Domain, error) {
	payload := map[string]interface{}{
		"host":  domain.Host,
		"path":  domain.Path,
		"port":  domain.Port,
		"https": domain.HTTPS,
	}
	if domain.ApplicationID != "" {
		payload["applicationId"] = domain.ApplicationID
	}
	if domain.ComposeID != "" {
		payload["composeId"] = domain.ComposeID
	}
	if domain.ServiceName != "" {
		payload["serviceName"] = domain.ServiceName
	}

	resp, err := c.doRequest("POST", "domain.create", payload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Domain Domain `json:"domain"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Domain.ID != "" {
		return &wrapper.Domain, nil
	}

	var result Domain
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) GetDomainsByApplication(appID string) ([]Domain, error) {
	app, err := c.GetApplication(appID)
	if err != nil {
		return nil, err
	}
	return app.Domains, nil
}

func (c *DokployClient) GetDomainsByCompose(composeID string) ([]Domain, error) {
	comp, err := c.GetCompose(composeID)
	if err != nil {
		return nil, err
	}
	return comp.Domains, nil
}

func (c *DokployClient) GetDomain(id string) (*Domain, error) {
	endpoint := fmt.Sprintf("domain.one?domainId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result Domain
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) DeleteDomain(id string) error {
	payload := map[string]string{
		"domainId": id,
	}
	_, err := c.doRequest("POST", "domain.remove", payload)
	return err
}

func (c *DokployClient) GenerateDomain(appName string) (string, error) {
	payload := map[string]string{
		"appName": appName,
	}
	resp, err := c.doRequest("POST", "domain.generateDomain", payload)
	if err != nil {
		return "", err
	}

	// Try to parse as JSON wrapper
	var wrapper struct {
		Domain string `json:"domain"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Domain != "" {
		return wrapper.Domain, nil
	}

	// Fallback: maybe it returns just the string in quotes or raw?
	// If it is a simple string "foo.bar", Unmarshal might fail or we just return string(resp) trimmed.
	return strings.Trim(string(resp), "\""), nil
}

func (c *DokployClient) UpdateDomain(domain Domain) (*Domain, error) {
	payload := map[string]interface{}{
		"domainId":    domain.ID,
		"host":        domain.Host,
		"path":        domain.Path,
		"port":        domain.Port,
		"https":       domain.HTTPS,
		"serviceName": domain.ServiceName,
	}
	resp, err := c.doRequest("POST", "domain.update", payload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Domain Domain `json:"domain"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Domain.ID != "" {
		return &wrapper.Domain, nil
	}

	var result Domain
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// --- Environment Variable ---

type EnvironmentVariable struct {
	ID            string `json:"id"`
	ApplicationID string `json:"applicationId"`
	Key           string `json:"key"`
	Value         string `json:"value"`
	Scope         string `json:"scope"`
}

func (c *DokployClient) UpdateApplicationEnv(appID string, updateFn func(envMap map[string]string), createEnvFile *bool) error {
	var lastErr error
	for i := 0; i < 5; i++ { // Retry up to 5 times
		app, err := c.GetApplication(appID)
		if err != nil {
			return err
		}

		envMap := ParseEnv(app.Env)
		originalEnvStr := app.Env

		updateFn(envMap) // Modify the map

		newEnvStr := formatEnv(envMap)

		if newEnvStr == originalEnvStr {
			return nil // No changes to be made
		}

		payload := map[string]interface{}{
			"applicationId": appID,
			"env":           newEnvStr,
		}
		if createEnvFile != nil {
			payload["createEnvFile"] = *createEnvFile
		}

		_, err = c.doRequest("POST", "application.saveEnvironment", payload)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(100*(i+1)) * time.Millisecond) // Backoff
			continue
		}

		// Verify write
		verifyApp, err := c.GetApplication(appID)
		if err != nil {
			// If we can't verify, we have to assume it worked or retry
			lastErr = fmt.Errorf("failed to verify environment update: %w", err)
			time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
			continue
		}
		if verifyApp.Env == newEnvStr {
			return nil // Success
		}
		lastErr = fmt.Errorf("environment update conflict, retrying")
	}
	return lastErr
}

func (c *DokployClient) CreateVariable(appID, key, value, scope string, createEnvFile *bool) (*EnvironmentVariable, error) {
	err := c.UpdateApplicationEnv(appID, func(envMap map[string]string) {
		envMap[key] = value
	}, createEnvFile)

	if err != nil {
		return nil, err
	}

	return &EnvironmentVariable{
		ID:            appID + "_" + key,
		ApplicationID: appID,
		Key:           key,
		Value:         value,
		Scope:         scope,
	}, nil
}

func (c *DokployClient) GetVariablesByApplication(appID string) ([]EnvironmentVariable, error) {
	app, err := c.GetApplication(appID)
	if err != nil {
		return nil, err
	}
	envMap := ParseEnv(app.Env)
	var vars []EnvironmentVariable
	for k, v := range envMap {
		vars = append(vars, EnvironmentVariable{
			ID:            appID + "_" + k,
			ApplicationID: appID,
			Key:           k,
			Value:         v,
			Scope:         "runtime",
		})
	}
	return vars, nil
}

func (c *DokployClient) DeleteVariable(id string, createEnvFile *bool) error {
	parts := strings.SplitN(id, "_", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid variable ID format")
	}
	appID, key := parts[0], parts[1]

	return c.UpdateApplicationEnv(appID, func(envMap map[string]string) {
		delete(envMap, key)
	}, createEnvFile)
}

func ParseEnv(env string) map[string]string {
	m := make(map[string]string)
	if env == "" {
		return m
	}
	lines := strings.Split(env, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}

func formatEnv(m map[string]string) string {
	var lines []string
	for k, v := range m {
		lines = append(lines, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(lines, "\n")
}

// --- SSH Key ---

type SSHKey struct {
	ID          string `json:"sshKeyId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PrivateKey  string `json:"privateKey"`
	PublicKey   string `json:"publicKey"`
}

func (c *DokployClient) CreateSSHKey(name, description, privateKey, publicKey string) (*SSHKey, error) {
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

func (c *DokployClient) ListSSHKeys() ([]SSHKey, error) {
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

func (c *DokployClient) findSSHKeyByName(name string) (*SSHKey, error) {
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

func (c *DokployClient) GetSSHKey(id string) (*SSHKey, error) {
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

func (c *DokployClient) DeleteSSHKey(id string) error {
	payload := map[string]string{
		"sshKeyId": id,
	}
	_, err := c.doRequest("POST", "sshKey.remove", payload)
	return err
}

// --- Server ---

type Server struct {
	ID          string `json:"serverId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IPAddress   string `json:"ipAddress"`
	Port        int64  `json:"port"`
	Username    string `json:"username"`
	SSHKeyID    string `json:"sshKeyId"`
}

func (c *DokployClient) CreateServer(server Server) (*Server, error) {
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

func (c *DokployClient) GetServer(id string) (*Server, error) {
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

func (c *DokployClient) ListServers() ([]Server, error) {
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

func (c *DokployClient) UpdateServer(server Server) (*Server, error) {
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

func (c *DokployClient) DeleteServer(id string) error {
	payload := map[string]string{
		"serverId": id,
	}
	_, err := c.doRequest("POST", "server.remove", payload)
	return err
}

// --- Registry ---

type Registry struct {
	ID            string `json:"registryId"`
	Name          string `json:"name"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	ImageRegistry string `json:"imageRegistry"`
	RegistryURL   string `json:"registryUrl,omitempty"`
}

func (c *DokployClient) CreateRegistry(registry Registry) (*Registry, error) {
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

func (c *DokployClient) GetRegistry(id string) (*Registry, error) {
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

func (c *DokployClient) ListRegistries() ([]Registry, error) {
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

func (c *DokployClient) UpdateRegistry(registry Registry) (*Registry, error) {
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

func (c *DokployClient) DeleteRegistry(id string) error {
	payload := map[string]string{
		"registryId": id,
	}
	_, err := c.doRequest("POST", "registry.remove", payload)
	return err
}

// --- Deployment ---

type Deployment struct {
	ID            string `json:"deploymentId"`
	ApplicationID string `json:"applicationId"`
	Status        string `json:"status"`
	Log           string `json:"log"`
	CreatedAt     string `json:"createdAt"`
}

func (c *DokployClient) CreateDeployment(appID string) (*Deployment, error) {
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

func (c *DokployClient) GetDeployment(id string) (*Deployment, error) {
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

func (c *DokployClient) ListDeployments(appID string) ([]Deployment, error) {
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
