package usecase

import (
	"context"
	"errors"

	"github.com/mj3smile/bank-statement-processor/internal/model/transaction"
	"github.com/mj3smile/bank-statement-processor/internal/model/upload"
	"github.com/mj3smile/bank-statement-processor/internal/repository"
)

type Issues interface {
	GetIssues(ctx context.Context, filters *transaction.IssuesFilters) (*IssuesResult, error)
}

type issues struct {
	transactionRepo repository.TransactionRepository
	uploadRepo      repository.UploadRepository
}

type IssuesResult struct {
	Transactions []*transaction.Transaction
	TotalCount   int
}

func NewIssues(transactionRepo repository.TransactionRepository, uploadRepo repository.UploadRepository) Issues {
	return &issues{
		transactionRepo: transactionRepo,
		uploadRepo:      uploadRepo,
	}
}

func (i *issues) GetIssues(ctx context.Context, filters *transaction.IssuesFilters) (*IssuesResult, error) {
	_, err := i.uploadRepo.GetByID(ctx, upload.ID(filters.UploadID))
	if err != nil {
		return nil, errors.New("upload not found")
	}

	transactions, totalCount, err := i.transactionRepo.GetIssuesWithFilters(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &IssuesResult{
		Transactions: transactions,
		TotalCount:   totalCount,
	}, nil
}
