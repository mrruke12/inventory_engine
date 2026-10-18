package config

import "github.com/joho/godotenv"

func Load() error {
	return godotenv.Load()
}
