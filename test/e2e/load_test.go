//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/application/orderusecase"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
	"github.com/mrruke12/inventory_engine/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	backgroundCleanerDelay time.Duration = time.Second
	testDuration           time.Duration = time.Second * 20
	initialProductStock                  = 500
	loadProductCount                     = 10
	cleanOrderBatchSize                  = 1_000
)

// Tests concurrent queries over orders, checks for
// mismatches in final products stock to detect anomalies
func TestLoad(t *testing.T) {
	maxConns := pool.Config().MaxConns
	require.GreaterOrEqual(t, int(maxConns), 5, "max conns must be >= 5")

	test.CleanupPostgres(t, pool)
	ctx, cancel := context.WithTimeout(context.Background(), testDuration+2*time.Minute)
	defer cancel()

	_, err := pool.Exec(ctx, `alter table orders alter column expires_at set default now() + interval '10 seconds'`)
	require.NoError(t, err)

	productIDs := make([]uuid.UUID, loadProductCount)
	for i := range productIDs {
		productIDs[i] = uuid.NewV7()
		_, err := pool.Exec(ctx,
			`insert into products (id, name, stock) values ($1, $2, $3)`,
			productIDs[i], fmt.Sprintf("load-product-%02d", i), initialProductStock,
		)
		require.NoError(t, err, "insert initial product")
	}

	stats := newLoadStats()
	panicCh := make(chan error, 5)
	actionsStop := make(chan struct{})
	cleanerDone := make(chan struct{})
	actionsDone := make(chan struct{})

	var stopActionsOnce sync.Once
	stopActions := func() {
		stopActionsOnce.Do(func() { close(actionsStop) })
	}

	workCtx, cancelWork := context.WithCancel(ctx)
	defer cancelWork()

	cleanerCtx, cancelCleaner := context.WithCancel(ctx)
	defer cancelCleaner()

	shouldStop := func() bool {
		select {
		case <-workCtx.Done():
			return true
		case <-actionsStop:
			return true
		default:
			return false
		}
	}

	waitOrStop := func(d time.Duration) {
		timer := time.NewTimer(d)
		defer timer.Stop()

		select {
		case <-workCtx.Done():
		case <-actionsStop:
		case <-timer.C:
		}
	}

	var actions sync.WaitGroup
	var cleanerWorker sync.WaitGroup

	runWorker := func(name string, wg *sync.WaitGroup, fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if recovered := recover(); recovered != nil {
					panicCh <- fmt.Errorf("%s worker panic: %v\n%s", name, recovered, debug.Stack())
					stopActions()
					cancelWork()
					cancelCleaner()
					cancel()
				}
				if name == "cleaner" {
					close(cleanerDone)
				}
			}()
			fn()
		}()
	}

	var ordersMu sync.RWMutex
	orders := make([]uuid.UUID, 0, 1024)

	addOrder := func(id uuid.UUID) {
		ordersMu.Lock()
		orders = append(orders, id)
		ordersMu.Unlock()
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var rngMu sync.Mutex

	randomOrder := func() (uuid.UUID, bool) {
		ordersMu.RLock()
		defer ordersMu.RUnlock()

		if len(orders) == 0 {
			return uuid.Nil(), false
		}

		rngMu.Lock()
		index := rng.Intn(len(orders))
		rngMu.Unlock()

		return orders[index], true
	}

	runWorker("cleaner", &cleanerWorker, func() {
		ticker := time.NewTicker(backgroundCleanerDelay)
		defer ticker.Stop()

		done := actionsDone
		draining := false

		for {
			select {
			case <-cleanerCtx.Done():
				return

			case <-done:
				draining = true
				done = nil

			case <-ticker.C:
			}

			err := service.CleanOrders(cleanerCtx, cleanOrderBatchSize)
			stats.recordResult("clean", err)

			if !draining {
				continue
			}

			pending, err := countPendingOrders(cleanerCtx)
			if err != nil {
				if cleanerCtx.Err() != nil {
					return
				}
				stats.recordError("clean", err)
				continue
			}

			slog.Info("pending remaining", "count", pending)

			if pending == 0 {
				slog.Info("cleaner finished")
				return
			}
		}
	})

	runWorker("create", &actions, func() {
		var sequence uint64

		for !shouldStop() {
			productID := productIDs[sequence%uint64(len(productIDs))]

			id, err := service.CreateOrder(workCtx, orderusecase.CreateOrderCommand{
				IdempotencyKey: uuid.NewV7().String(),
				OrderItems:     []order.OrderItem{{ProductID: productID, Amount: 1}},
			})

			stats.recordResult("create", err)

			if err == nil {
				stats.created.Add(1)
				addOrder(id)
			}

			sequence++
		}
	})

	runWorker("release", &actions, func() {
		for !shouldStop() {
			id, ok := randomOrder()
			if !ok {
				waitOrStop(time.Millisecond)
				continue
			}

			_, err := service.ReleaseOrder(workCtx, id)
			stats.recordResult("release", err)
		}
	})

	runWorker("complete", &actions, func() {
		for !shouldStop() {
			id, ok := randomOrder()
			if !ok {
				waitOrStop(time.Millisecond)
				continue
			}

			_, err := service.CompleteOrder(workCtx, id)
			stats.recordResult("complete", err)
		}
	})

	runWorker("cancel", &actions, func() {
		for !shouldStop() {
			id, ok := randomOrder()
			if !ok {
				waitOrStop(time.Millisecond)
				continue
			}

			_, err := service.CancelOrder(workCtx, id)
			stats.recordResult("cancel", err)
		}
	})

	timer := time.NewTimer(testDuration)
	defer timer.Stop()

	failOnWorkerPanic := func() {
		select {
		case workerPanic := <-panicCh:
			t.Fatal(workerPanic)
		default:
		}
	}

	select {
	case <-timer.C:
		stopActions()

	case workerPanic := <-panicCh:
		stopActions()
		cancelWork()
		cancelCleaner()
		actions.Wait()
		<-cleanerDone
		cleanerWorker.Wait()
		t.Fatal(workerPanic)

	case <-ctx.Done():
		stopActions()
		cancelWork()
		cancelCleaner()
		actions.Wait()
		<-cleanerDone
		cleanerWorker.Wait()
		failOnWorkerPanic()
		t.Fatalf("load test timed out: %v", ctx.Err())
	}

	actions.Wait()
	close(actionsDone)

	select {
	case <-cleanerDone:
		cleanerWorker.Wait()

	case workerPanic := <-panicCh:
		cancelCleaner()
		<-cleanerDone
		cleanerWorker.Wait()
		t.Fatal(workerPanic)

	case <-ctx.Done():
		cancelCleaner()
		<-cleanerDone
		cleanerWorker.Wait()
		failOnWorkerPanic()
		t.Fatalf("timed out draining pending orders: %v", ctx.Err())
	}

	failOnWorkerPanic()

	statusCounts, err := countOrdersByStatus(ctx)
	require.NoError(t, err, "count orders by status")

	completedByProductID, err := countCompletedByProductID(ctx)
	require.NoError(t, err)

	productStocks, err := readProductStocks(ctx)
	require.NoError(t, err, "read product stocks")

	t.Logf("created orders: %d", stats.created.Load())
	t.Logf("operation statistics: %s", stats.summary())
	t.Logf("orders by status: %v", statusCounts)
	t.Logf("cleaner errors: %v", stats.errorsFor("clean"))

	for _, productID := range productIDs {
		stock, ok := productStocks[productID]
		require.True(t, ok, "product %s is missing from final stock results", productID)

		completed, ok := completedByProductID[productID]
		require.True(t, ok, "product %s is missing from final completed by product ID results", productID)

		assert.Equalf(t, initialProductStock, stock+completed,
			"product %s stock changed after load test", productID)
	}

	require.Zero(t, statusCounts[string(order.StatusPending)],
		"pending orders remain after cleaner drain")
}

type operationStats struct {
	attempts  uint64
	succeeded uint64
	errors    map[string]uint64
}

type loadStats struct {
	mu         sync.Mutex
	created    atomic.Uint64
	operations map[string]*operationStats
}

func newLoadStats() *loadStats {
	stats := &loadStats{operations: make(map[string]*operationStats)}

	for _, name := range []string{"create", "release", "complete", "cancel", "clean"} {
		stats.operations[name] = &operationStats{errors: make(map[string]uint64)}
	}

	return stats
}

func (s *loadStats) recordResult(operation string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stat := s.operations[operation]
	stat.attempts++

	if err == nil {
		stat.succeeded++
		return
	}

	stat.errors[err.Error()]++
}

func (s *loadStats) recordError(operation string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stat := s.operations[operation]
	stat.attempts++
	stat.errors[err.Error()]++
}

func (s *loadStats) summary() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := ""

	for _, name := range []string{"create", "release", "complete", "cancel", "clean"} {
		stat := s.operations[name]

		result += fmt.Sprintf(
			"%s={attempts:%d succeeded:%d errors:%d details:%v} ",
			name,
			stat.attempts,
			stat.succeeded,
			errorCount(stat.errors),
			stat.errors,
		)
	}

	return result
}

func (s *loadStats) errorsFor(operation string) map[string]uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(map[string]uint64, len(s.operations[operation].errors))

	for message, count := range s.operations[operation].errors {
		result[message] = count
	}

	return result
}

func errorCount(errors map[string]uint64) uint64 {
	var count uint64

	for _, n := range errors {
		count += n
	}

	return count
}

func countPendingOrders(ctx context.Context) (int64, error) {
	var pending int64

	err := pool.QueryRow(ctx,
		`select count(*) from orders where status = 'pending'`,
	).Scan(&pending)

	return pending, err
}

func countCompletedByProductID(ctx context.Context) (map[uuid.UUID]int, error) {
	counts := make(map[uuid.UUID]int)

	rows, err := pool.Query(ctx, `
		select p.id as product_id,
			coalesce(sum(oi.amount) filter (where o.status = 'completed'), 0) as reserved_amount
		from products p
		left join order_items oi on oi.product_id = p.id
		left join orders o on o.id = oi.order_id
		group by p.id;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var productID uuid.UUID
		var count int

		if err := rows.Scan(&productID, &count); err != nil {
			return nil, err
		}

		counts[productID] = count
	}

	return counts, rows.Err()
}

func countOrdersByStatus(ctx context.Context) (map[string]int64, error) {
	counts := map[string]int64{
		string(order.StatusPending):   0,
		string(order.StatusReleased):  0,
		string(order.StatusExpired):   0,
		string(order.StatusCanceled):  0,
		string(order.StatusCompleted): 0,
	}

	rows, err := pool.Query(ctx,
		`select status::text, count(*) from orders group by status`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64

		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}

		counts[status] = count
	}

	return counts, rows.Err()
}

func readProductStocks(ctx context.Context) (map[uuid.UUID]int, error) {
	rows, err := pool.Query(ctx, `select id, stock from products`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stocks := make(map[uuid.UUID]int, loadProductCount)

	for rows.Next() {
		var id uuid.UUID
		var stock int

		if err := rows.Scan(&id, &stock); err != nil {
			return nil, err
		}

		stocks[id] = stock
	}

	return stocks, rows.Err()
}
