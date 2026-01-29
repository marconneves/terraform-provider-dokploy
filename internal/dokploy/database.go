package dokploy

import (
	"encoding/json"
	"fmt"
)

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

func (c *Client) CreateDatabase(projectID, environmentID, name, dbType, password, dockerImage string) (*Database, error) {
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

func (c *Client) GetDatabase(dbID string, databaseType string) (*Database, error) {
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

func (c *Client) DeleteDatabase(id string) error {
	return fmt.Errorf("delete database requires type update")
}

func (c *Client) DeleteDatabaseWithType(id, dbType string) error {
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
