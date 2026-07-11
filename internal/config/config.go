package config

import "os"

type Config struct {
	AppName         string
	BasePath        string
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
		DataDir:         dataDir,
		Host:            host,
		PersistencePath: persistencePath,
		Port:            port,
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
