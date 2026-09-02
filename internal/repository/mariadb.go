package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/go-sql-driver/mysql"
)

const createStateTableSQL = `
CREATE TABLE IF NOT EXISTS relic_raid_state (
    state_key VARCHAR(64) NOT NULL PRIMARY KEY,
    payload LONGTEXT NOT NULL CHECK (JSON_VALID(payload)),
    updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
        ON UPDATE CURRENT_TIMESTAMP(6)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci`

type MariaDBConfig struct {
	Host     string
	Name     string
	Password string
	Port     string
	StateKey string
	User     string
}

type MariaDBStateStore struct {
	db       *sql.DB
	stateKey string
}

func NewMariaDBStateStore(cfg MariaDBConfig) (*MariaDBStateStore, error) {
	if cfg.Host == "" || cfg.Name == "" || cfg.User == "" || cfg.StateKey == "" {
		return nil, errors.New("DB_HOST, DB_NAME, DB_USER, and DB_STATE_KEY are required when DB_ENABLED is true")
	}

	driverConfig := mysql.NewConfig()
	driverConfig.User = cfg.User
	driverConfig.Passwd = cfg.Password
	driverConfig.Net = "tcp"
	driverConfig.Addr = net.JoinHostPort(cfg.Host, cfg.Port)
	driverConfig.DBName = cfg.Name
	driverConfig.ParseTime = true
	driverConfig.Collation = "utf8mb4_unicode_ci"
	driverConfig.Params = map[string]string{"charset": "utf8mb4"}

	db, err := sql.Open("mysql", driverConfig.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open MariaDB: %w", err)
	}
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(8)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping MariaDB: %w", err)
	}
	if _, err := db.ExecContext(ctx, createStateTableSQL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize MariaDB state table: %w", err)
	}

	return &MariaDBStateStore{db: db, stateKey: cfg.StateKey}, nil
}

func (s *MariaDBStateStore) Close() error {
	return s.db.Close()
}

func (s *MariaDBStateStore) Load(ctx context.Context) ([]byte, error) {
	var payload []byte
	err := s.db.QueryRowContext(
		ctx,
		"SELECT payload FROM relic_raid_state WHERE state_key = ?",
		s.stateKey,
	).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPersistentStateNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load MariaDB state: %w", err)
	}
	return payload, nil
}

func (s *MariaDBStateStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *MariaDBStateStore) Save(ctx context.Context, payload []byte) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO relic_raid_state (state_key, payload)
         VALUES (?, ?)
         ON DUPLICATE KEY UPDATE payload = VALUES(payload)`,
		s.stateKey,
		payload,
	)
	if err != nil {
		return fmt.Errorf("save MariaDB state: %w", err)
	}
	return nil
}
