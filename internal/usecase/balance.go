package usecase

import (
	"context"
	"fmt"

	"github.com/mj3smile/bank-statement-processor/internal/infra/log"
	"github.com/mj3smile/bank-statement-processor/internal/model/upload"
	"github.com/mj3smile/bank-statement-processor/internal/repository"
)

type Balance interface {
	Get(ctx context.Context, uploadID string) (*GetBalanceResult, error)
}

type balance struct {
	transactionRepo repository.TransactionRepository
	uploadRepo      repository.UploadRepository
}

type GetBalanceResult struct {
	UploadID          string
	Balance           *int64
	UploadTaskStatus  string
	UploadTaskMessage string
}

func NewBalance(transactionRepo repository.TransactionRepository, uploadRepo repository.UploadRepository) Balance {
	return &balance{
		transactionRepo: transactionRepo,
		uploadRepo:      uploadRepo,
	}
}

func (g *balance) Get(ctx context.Context, uploadID string) (*GetBalanceResult, error) {
	task, err := g.uploadRepo.GetByID(ctx, upload.ID(uploadID))
	if err != nil {
		log.Info(ctx, fmt.Sprint("get upload task error: ", err.Error()))
		return nil, err
	}

	response := &GetBalanceResult{
		UploadID:          uploadID,
		UploadTaskStatus:  string(task.Status),
		UploadTaskMessage: task.Message,
	}

	if task.Status != upload.StatusCompleted {
		return response, nil
	}

	b := g.transactionRepo.GetBalanceByUploadID(ctx, upload.ID(uploadID))
	response.Balance = &b
	return response, nil
}
