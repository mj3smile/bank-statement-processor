package repository

import (
	"context"

	"github.com/mj3smile/bank-statement-processor/internal/model/transaction"
	"github.com/mj3smile/bank-statement-processor/internal/model/upload"
)

type UploadRepository interface {
	Save(ctx context.Context, uploadTask *upload.Task) error
	Update(ctx context.Context, updateValue *upload.Task) error
	GetByID(ctx context.Context, uploadID upload.ID) (*upload.Task, error)
}

type TransactionRepository interface {
	Save(ctx context.Context, t *transaction.Transaction) error
	GetBalanceByUploadID(ctx context.Context, uploadID upload.ID) int64
	GetIssuesWithFilters(ctx context.Context, filters *transaction.IssuesFilters) ([]*transaction.Transaction, int, error)
	//PrepareDataForFilters(ctx context.Context, uploadID upload.ID)
	//CalculateBalance(uploadID upload.ID) int64
}
