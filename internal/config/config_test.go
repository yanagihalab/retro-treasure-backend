package config

import "testing"

func TestLoadDatabaseConfig(t *testing.T) {
	t.Setenv("DB_ENABLED", "true")
	t.Setenv("DB_HOST", "127.0.0.1")
	t.Setenv("DB_NAME", "relic_raid")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_USER", "relic_raid_app")

	cfg := Load()
	if !cfg.Database.Enabled {
		t.Fatal("database should be enabled")
	}
	if cfg.Database.Port != "3306" {
		t.Fatalf("unexpected default database port: %s", cfg.Database.Port)
	}
	if cfg.Database.StateKey != "primary" {
		t.Fatalf("unexpected default state key: %s", cfg.Database.StateKey)
	}
	if cfg.Database.Password != "secret" {
		t.Fatal("database password was not loaded")
	}
}

func TestLoadRespectsAppHost(t *testing.T) {
	t.Setenv("APP_HOST", "127.0.0.1")
	t.Setenv("APP_PORT", "8090")

	cfg := Load()
	if got, want := cfg.Addr(), "127.0.0.1:8090"; got != want {
		t.Fatalf("Addr() = %q, want %q", got, want)
	}
}
