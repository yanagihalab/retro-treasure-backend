package config

import "os"

type Config struct {
	AppName string
	Port    string
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

	return Config{
		AppName: name,
		Port:    port,
	}
}
