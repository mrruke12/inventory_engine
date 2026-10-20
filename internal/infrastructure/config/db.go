package config

import (
	"fmt"

	"github.com/mrruke12/inventory_engine/pkg/env"
)

type DBConfig struct {
	Host     string `env:"DB_HOST"`
	Port     int    `env:"DB_PORT"`
	DBName   string `env:"DB_NAME"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	SSLMode  string `env:"DB_SSL_MODE"`
}

func (c *DBConfig) ConnString() string {
	return fmt.Sprintf(
		`postgresql://%s:%s@%s:%d/%s?sslmode=%s`,
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
		c.SSLMode,
	)
}

func GetDBConfig() (*DBConfig, error) {
	cfg := &DBConfig{}

	err := env.UnmarshalStruct(cfg)

	if err != nil {
		return nil, err
	}

	return cfg, nil
}
