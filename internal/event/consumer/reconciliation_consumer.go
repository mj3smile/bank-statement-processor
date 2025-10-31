package consumer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mj3smile/bank-statement-processor/internal/event"
	"github.com/mj3smile/bank-statement-processor/internal/infra/log"
	"github.com/mj3smile/bank-statement-processor/internal/model/transaction"
)

type ReconciliationConsumer struct {
	eventBus       event.Bus
	workerCount    int
	maxRetries     int
	baseBackoff    time.Duration
	processedCache sync.Map
	wg             sync.WaitGroup
}

func NewReconciliationConsumer(eventBus event.Bus, workerCount int) *ReconciliationConsumer {
	return &ReconciliationConsumer{
		eventBus:    eventBus,
		workerCount: workerCount,
		maxRetries:  3,
		baseBackoff: time.Second,
	}
}

func (c *ReconciliationConsumer) Start(ctx context.Context) {
	log.Info(ctx, fmt.Sprintf("starting reconciliation consumer with %d workers", c.workerCount))

	eventChan := c.eventBus.Subscribe()
	for i := 0; i < c.workerCount; i++ {
		c.wg.Add(1)
		go c.worker(ctx, i, eventChan)
	}

	<-ctx.Done()
	log.Info(ctx, "reconciliation consumer received shutdown signal")

	c.wg.Wait()
	log.Info(ctx, "all reconciliation workers stopped")
}

func (c *ReconciliationConsumer) worker(ctx context.Context, id int, eventChan <-chan event.Event) {
	defer c.wg.Done()

	log.Info(ctx, fmt.Sprintf("reconciliation worker %d started", id))
	for {
		select {
		case <-ctx.Done():
			log.Info(ctx, fmt.Sprintf("reconciliation worker %d stopping (context cancelled)", id))
			return

		case evt, ok := <-eventChan:
			if !ok {
				log.Info(ctx, fmt.Sprintf("reconciliation worker %d stopping (channel closed)", id))
				return
			}

			if failedTransactionEvent, ok := evt.(event.FailedTransactionEvent); ok {
				if err := c.processEvent(ctx, id, failedTransactionEvent); err != nil {
					log.Info(ctx, fmt.Sprintf("worker %d: failed to process event: %v", id, err))
				}
			} else {
				log.Info(ctx, fmt.Sprintf("worker %d: received unexpected event type: %s", id, evt.EventType()))
			}
		}
	}
}

func (c *ReconciliationConsumer) processEvent(ctx context.Context, workerID int, evt event.FailedTransactionEvent) error {
	if c.isAlreadyProcessed(evt.TransactionID) {
		log.Info(ctx, fmt.Sprintf("worker %d: transaction %s already processed, skipping", workerID, evt.TransactionID))
		return nil
	}

	log.Info(ctx, fmt.Sprintf("worker %d: reconciling failed transaction %s (upload: %s, counterparty: %s, amount: %d)",
		workerID, evt.TransactionID, evt.UploadID, evt.Counterparty, evt.Amount))

	var lastErr error
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("reconciliation cancelled for transaction %s", evt.TransactionID)
		default:
		}

		if attempt > 0 {
			backoffDuration := c.calculateBackoff(attempt)
			log.Info(ctx, fmt.Sprintf("worker %d: retry attempt %d/%d for transaction %s (backoff: %v)",
				workerID, attempt+1, c.maxRetries, evt.TransactionID, backoffDuration))

			select {
			case <-time.After(backoffDuration):
			case <-ctx.Done():
				return fmt.Errorf("reconciliation cancelled during backoff for transaction %s", evt.TransactionID)
			}
		}

		if err := c.reconcile(ctx, evt); err != nil {
			lastErr = err
			log.Info(ctx, fmt.Sprintf("worker %d: reconciliation attempt %d failed for transaction %s: %v",
				workerID, attempt+1, evt.TransactionID, err))

			if attempt == c.maxRetries-1 {
				log.Info(ctx, fmt.Sprintf("worker %d: max retries (%d) reached for transaction %s, giving up",
					workerID, c.maxRetries, evt.TransactionID))
				return fmt.Errorf("failed to reconcile transaction %s after %d attempts: %w",
					evt.TransactionID, c.maxRetries, lastErr)
			}
			continue
		}

		c.markAsProcessed(evt.TransactionID)
		log.Info(ctx, fmt.Sprintf("worker %d: successfully reconciled transaction %s after %d attempt(s)",
			workerID, evt.TransactionID, attempt+1))
		return nil
	}

	return lastErr
}

func (c *ReconciliationConsumer) reconcile(ctx context.Context, evt event.FailedTransactionEvent) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("reconciliation work cancelled")
	case <-time.After(100 * time.Millisecond):
		// simulate random failure
		if time.Now().UnixNano()%5 == 0 {
			return fmt.Errorf("reconciliation failed (simulated transient error)")
		}
	}

	log.Info(ctx, fmt.Sprintf("reconciliation completed: tx=%s, counterparty=%s, amount=%d, description=%s",
		evt.TransactionID, evt.Counterparty, evt.Amount, evt.Description))

	return nil
}

func (c *ReconciliationConsumer) calculateBackoff(attempt int) time.Duration {
	backOffTime := int(c.baseBackoff)
	for _ = range attempt {
		backOffTime *= 2
	}
	return time.Duration(backOffTime)
}

func (c *ReconciliationConsumer) isAlreadyProcessed(transactionID transaction.ID) bool {
	_, exists := c.processedCache.Load(transactionID)
	return exists
}

func (c *ReconciliationConsumer) markAsProcessed(transactionID transaction.ID) {
	c.processedCache.Store(transactionID, true)
}

func (c *ReconciliationConsumer) GetProcessedCount() int {
	count := 0
	c.processedCache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
