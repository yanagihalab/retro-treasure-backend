package config

import (
	"os"
	"strings"
)

type DatabaseConfig struct {
	Enabled  bool
	Host     string
	Name     string
	Password string
	Port     string
	StateKey string
	User     string
}

type Config struct {
	AppName         string
	BasePath        string
	Database        DatabaseConfig
	DataDir         string
	Host            string
	PersistencePath string
	Port            string
}

func Load() Config {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	name := os.Getenv("APP_NAME")
	if name == "" {
		name = "retro-treasure-api"
	}

	host := os.Getenv("APP_HOST")
	basePath := normalizeBasePath(os.Getenv("APP_BASE_PATH"))
	dataDir := os.Getenv("DATA_DIR")
	persistencePath := os.Getenv("APP_STATE_FILE")
	if persistencePath == "" && dataDir != "" {
		persistencePath = dataDir + "/state.json"
	}

	return Config{
		AppName:         name,
		BasePath:        basePath,
		Database:        loadDatabaseConfig(),
		DataDir:         dataDir,
		Host:            host,
		PersistencePath: persistencePath,
		Port:            port,
	}
}

func loadDatabaseConfig() DatabaseConfig {
	databasePort := os.Getenv("DB_PORT")
	if databasePort == "" {
		databasePort = "3306"
	}

	stateKey := os.Getenv("DB_STATE_KEY")
	if stateKey == "" {
		stateKey = "primary"
	}

	return DatabaseConfig{
		Enabled:  envEnabled(os.Getenv("DB_ENABLED")),
		Host:     os.Getenv("DB_HOST"),
		Name:     os.Getenv("DB_NAME"),
		Password: os.Getenv("DB_PASSWORD"),
		Port:     databasePort,
		StateKey: stateKey,
		User:     os.Getenv("DB_USER"),
	}
}

func envEnabled(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func (c Config) Addr() string {
	if c.Host == "" {
		return ":" + c.Port
	}
	return c.Host + ":" + c.Port
}

func normalizeBasePath(path string) string {
	if path == "" || path == "/" {
		return ""
	}
	if path[0] != '/' {
		path = "/" + path
	}
	for len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return path
}
