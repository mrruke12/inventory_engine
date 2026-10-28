//go:build integration

package infrastructure

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mrruke12/inventory_engine/test/integration"
	container "github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/goleak"
)

var (
	pgContainer *container.PostgresContainer
	pool        *pgxpool.Pool
)

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) (code int) {
	if err := setup(); err != nil {
		fmt.Fprintf(os.Stderr, "setup: %v\n", err)
		return 1
	}

	ignore := goleak.IgnoreCurrent()

	code = m.Run()

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "teardown panic: %v\n%s", r, debug.Stack())
				if code == 0 {
					code = 1
				}
			}
		}()
		if err := teardown(); err != nil {
			fmt.Fprintf(os.Stderr, "teardown: %v\n", err)
			code = 1
		}
	}()

	if err := goleak.Find(ignore); err != nil {
		fmt.Fprintf(os.Stderr, "goleak: %v\n", err)
		code = 1
	}

	return code
}

func setup() error {
	var err error

	pgContainer, pool, err = integration.SetupPostgres()

	return err
}

func teardown() error {
	pool.Close()

	return pgContainer.Terminate(context.Background())
}
