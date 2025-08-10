package gmcfg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DatabaseConfig represents a single database configuration
type DatabaseConfig struct {
	Type string `json:"type"` // "sqlite" or "postgres"
	Role string `json:"role"` // "sync", "oltp", etc.
	Path string `json:"path"` // File path for SQLite, connection string for Postgres
}

// DatabasesConfig represents the overall database configuration
type DatabasesConfig struct {
	Default   string                    `json:"default"`
	Databases map[string]DatabaseConfig `json:"databases"`
}

const DatabasesConfigFile = "databases.json"

// LoadDatabasesConfig loads the database configuration from file
func (s *FileStore) LoadDatabasesConfig() (config DatabasesConfig, err error) {
	err = s.Load(DatabasesConfigFile, &config)
	if err != nil {
		// If file doesn't exist, return default config
		if os.IsNotExist(err) {
			config = getDefaultDatabasesConfig()
			err = nil
		}
		goto end
	}

	// Expand paths for SQLite databases
	err = s.expandDatabasePaths(&config)

end:
	return config, err
}

// SaveDatabasesConfig saves the database configuration to file
func (s *FileStore) SaveDatabasesConfig(config DatabasesConfig) (err error) {
	err = s.Save(DatabasesConfigFile, config)
	return err
}

// GetDatabaseConfig returns configuration for a named database
func (s *FileStore) GetDatabaseConfig(name string) (config DatabaseConfig, err error) {
	var dbsConfig DatabasesConfig
	var exists bool

	if name == "" {
		err = fmt.Errorf("database name is required")
		goto end
	}

	dbsConfig, err = s.LoadDatabasesConfig()
	if err != nil {
		goto end
	}

	// If name is "default", use the default database
	if name == "default" {
		if dbsConfig.Default == "" {
			err = fmt.Errorf("no default database configured")
			goto end
		}
		name = dbsConfig.Default
	}

	config, exists = dbsConfig.Databases[name]
	if !exists {
		err = fmt.Errorf("database '%s' not found in configuration", name)
		goto end
	}

end:
	return config, err
}

// expandDatabasePaths expands ~ and relative paths in database configurations
func (s *FileStore) expandDatabasePaths(config *DatabasesConfig) (err error) {
	var homeDir string

	homeDir, err = os.UserHomeDir()
	if err != nil {
		goto end
	}

	for name, dbConfig := range config.Databases {
		if dbConfig.Type == "sqlite" {
			dbConfig.Path = s.expandPath(dbConfig.Path, homeDir)
			config.Databases[name] = dbConfig
		}
	}

end:
	return err
}

// expandPath expands ~ to home directory and resolves relative paths
func (s *FileStore) expandPath(path, homeDir string) string {
	var result string

	if strings.HasPrefix(path, "~/") {
		result = filepath.Join(homeDir, path[2:])
		goto end
	}

	if !filepath.IsAbs(path) {
		result = filepath.Join(homeDir, path)
		goto end
	}

	result = path

end:
	return result
}

// getDefaultDatabasesConfig returns the default database configuration
func getDefaultDatabasesConfig() DatabasesConfig {
	return DatabasesConfig{
		Default: "default-archive",
		Databases: map[string]DatabaseConfig{
			"default-archive": {
				Type: "sqlite",
				Role: "sync",
				Path: "~/gmail-archive/default.db",
			},
		},
	}
}
