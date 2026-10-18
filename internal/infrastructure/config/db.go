package config

import (
	"fmt"

	"github.com/mrruke12/inventory_engine/pkg/env"
)

type DBConfig struct {
	Host     string
	Port     int
	DBName   string
	User     string
	Password string
	SSLMode  string
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
	values, err := env.RequireKeys([]string{
		"DB_HOST",
		"DB_NAME",
		"DB_USER",
		"DB_PASSWORD",
		"DB_SSL_MODE",
	})

	if err != nil {
		return nil, err
	}

	port, err := env.RequireInt("DB_PORT")

	if err != nil {
		return nil, err
	}

	return &DBConfig{
		Host:     values[0],
		Port:     port,
		DBName:   values[1],
		User:     values[2],
		Password: values[3],
		SSLMode:  values[4],
	}, nil
}
