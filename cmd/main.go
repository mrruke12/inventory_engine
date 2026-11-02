package main

import (
	"context"
	"fmt"

	"github.com/mrruke12/inventory_engine/internal/infrastructure/config"
	"github.com/mrruke12/inventory_engine/internal/infrastructure/postgres"
)

func main() {
	err := config.Load()

	if err != nil {
		panic(err)
	}

	conf, err := config.GetDBConfig()

	if err != nil {
		panic(err)
	}

	err = postgres.Migrate("file://internal/infrastructure/migrations", conf.ConnString())

	if err != nil {
		panic(err)
	}

	pool, err := postgres.Connect(context.Background(), conf.ConnString())

	if err != nil {
		panic(err)
	}

	err = pool.Ping(context.Background())

	if err != nil {
		panic(err)
	}

	defer pool.Close()

	fmt.Println("All's set up")
}
