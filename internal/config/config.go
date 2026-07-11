package config

import "os"

type Config struct {
	AppName string
	Host    string
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

	host := os.Getenv("APP_HOST")

	return Config{
		AppName: name,
		Host:    host,
		Port:    port,
	}
}

func (c Config) Addr() string {
	if c.Host == "" {
		return ":" + c.Port
	}
	return c.Host + ":" + c.Port
}
