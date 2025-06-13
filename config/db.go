package config

import "fmt"

const (
	host     = "localhost"
	port     = "5432"
	user     = "poodonkis"
	password = "douglas123"
	database = "as-attendance"
	sslMode  = "disable"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func (cfg DBConfig) String() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s", cfg.User, cfg.Password, cfg.Host,
		cfg.Port, cfg.Database, cfg.SSLMode)
}

func DefaultConfig() DBConfig {
	return DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
		SSLMode:  sslMode,
	}
}
