package main

import (
	"context"
	"fmt"
	"sync"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/application/orderusecase"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
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

	err = postgres.Migrate(conf.ConnString())

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

	uow := postgres.NewUnitOfWork(pool, postgres.NewRepositoryProvider)
	orderRepo := postgres.NewOrderRepository(pool)

	s := orderusecase.NewService(uow, orderRepo)

	var wg sync.WaitGroup

	for i := range 10 {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			for range 1000 {
				_, err := s.CreateOrder(context.Background(), []order.OrderItem{
					{ProductID: uuid.MustParse("01a15049-79f6-733f-9e4c-a620316a3dcc"), Quantity: 1},
				})

				// err = s.CleanOrders(context.Background())

				if err != nil {
					fmt.Println(i, err)
				}
			}
		}(i)
	}

	wg.Wait()

	fmt.Println("done")

	pool.Close()
}
