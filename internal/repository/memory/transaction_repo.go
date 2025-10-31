package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/mj3smile/bank-statement-processor/internal/model/transaction"
	"github.com/mj3smile/bank-statement-processor/internal/model/upload"
	"github.com/mj3smile/bank-statement-processor/internal/repository"
)

//type problematicTransactions struct {
//	byStatusFailed    []*transaction.Transaction
//	byStatusPending   []*transaction.Transaction
//	sortedByTimestamp []*transaction.Transaction
//	sortedByAmount    []*transaction.Transaction
//}

type transactionRepository struct {
	mu                sync.RWMutex
	transactions      map[transaction.ID]*transaction.Transaction
	uploadIdToIssues  map[upload.ID][]*transaction.Transaction
	uploadIdToBalance map[upload.ID]int64
}

func NewTransactionRepository() repository.TransactionRepository {
	return &transactionRepository{
		transactions:      make(map[transaction.ID]*transaction.Transaction),
		uploadIdToIssues:  make(map[upload.ID][]*transaction.Transaction),
		uploadIdToBalance: make(map[upload.ID]int64),
	}
}

func (tr *transactionRepository) Save(ctx context.Context, t *transaction.Transaction) error {
	if t == nil {
		return errors.New("transaction cannot be nil")
	}

	if t.ID == "" {
		return errors.New("transaction ID cannot be empty")
	}

	if t.UploadID == "" {
		return errors.New("upload ID cannot be empty")
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()

	if _, exists := tr.transactions[t.ID]; exists {
		return errors.New("transaction already exists")
	}

	tr.transactions[t.ID] = t
	if t.Status == transaction.StatusSuccess {
		balance := tr.uploadIdToBalance[t.UploadID]
		if t.Type == transaction.TypeCredit {
			balance += t.Amount
		} else if t.Type == transaction.TypeDebit {
			balance -= t.Amount
		}
		tr.uploadIdToBalance[t.UploadID] = balance

	} else {
		tr.uploadIdToIssues[t.UploadID] = append(tr.uploadIdToIssues[t.UploadID], t)
	}

	return nil
}

func (tr *transactionRepository) GetBalanceByUploadID(ctx context.Context, uploadID upload.ID) int64 {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	balance := tr.uploadIdToBalance[uploadID]
	return balance
}

func (tr *transactionRepository) GetIssuesWithFilters(ctx context.Context, filters *transaction.IssuesFilters) ([]*transaction.Transaction, int, error) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	allTransactions, exists := tr.uploadIdToIssues[filters.UploadID]
	if !exists {
		return []*transaction.Transaction{}, 0, nil
	}

	filtered := make([]*transaction.Transaction, 0)
	for _, t := range allTransactions {
		if filters.Status != nil && t.Status != *filters.Status {
			continue
		}

		if filters.MinAmount != nil && t.Amount < *filters.MinAmount {
			continue
		}
		if filters.MaxAmount != nil && t.Amount > *filters.MaxAmount {
			continue
		}

		if filters.FromDate != nil && t.Timestamp < *filters.FromDate {
			continue
		}
		if filters.ToDate != nil && t.Timestamp > *filters.ToDate {
			continue
		}

		filtered = append(filtered, t)
	}

	totalCount := len(filtered)
	offset := (filters.Page - 1) * filters.PageSize
	if offset >= totalCount {
		return []*transaction.Transaction{}, totalCount, nil
	}

	end := offset + filters.PageSize
	if end > totalCount {
		end = totalCount
	}

	return filtered[offset:end], totalCount, nil
}

//func (tr *transactionRepository) PrepareDataForFilters(ctx context.Context, uploadID upload.ID) {
//	tr.mu.Lock()
//	defer tr.mu.Unlock()
//
//	c, ok := tr.uploadIdToIssues[uploadID]
//	if !ok {
//		return
//	}
//
//	sort.Slice(c.sortedByAmount, func(i, j int) bool { return c.sortedByAmount[i].Amount < c.sortedByAmount[j].Amount })
//	sort.Slice(c.sortedByTimestamp, func(i, j int) bool { return c.sortedByTimestamp[i].Timestamp < c.sortedByTimestamp[j].Timestamp })
//	for _, t := range c.sortedByTimestamp {
//		if t.Status == transaction.StatusFailed {
//			c.byStatusFailed = append(c.byStatusFailed, t)
//		} else if t.Status == transaction.StatusPending {
//			c.byStatusPending = append(c.byStatusPending, t)
//		}
//	}
//
//	tr.uploadIdToIssues[uploadID] = c
//}

//func (tr *transactionRepository) CalculateBalance(uploadID upload.ID) int64 {
//	tr.mu.RLock()
//	defer tr.mu.RUnlock()
//
//	txIDs, exists := tr.uploadIdToTransactionsID[uploadID]
//	if !exists {
//		return 0
//	}
//
//	var balance int64
//	for _, txID := range txIDs {
//		if tx, exists := tr.transactions[txID]; exists {
//			if tx.Status == transaction.StatusCompleted {
//				if tx.Type == transaction.TypeCredit {
//					balance += tx.Amount
//				} else if tx.Type == transaction.TypeDebit {
//					balance -= tx.Amount
//				}
//			}
//		}
//	}
//
//	return balance
//}
