package usecase

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mj3smile/bank-statement-processor/internal/event"
	"github.com/mj3smile/bank-statement-processor/internal/infra/log"
	"github.com/mj3smile/bank-statement-processor/internal/model/transaction"
	"github.com/mj3smile/bank-statement-processor/internal/model/upload"
	"github.com/mj3smile/bank-statement-processor/internal/repository"
)

type Statement interface {
	Upload(ctx context.Context, file multipart.File, filename string) (upload.ID, error)
}

type statement struct {
	appCtx          context.Context
	transactionRepo repository.TransactionRepository
	uploadRepo      repository.UploadRepository
	eventBus        event.Bus
}

func NewStatement(appCtx context.Context, transactionRepo repository.TransactionRepository, uploadRepo repository.UploadRepository, eventBus event.Bus) Statement {
	return &statement{
		appCtx:          appCtx,
		transactionRepo: transactionRepo,
		uploadRepo:      uploadRepo,
		eventBus:        eventBus,
	}
}

func (uc *statement) Upload(ctx context.Context, file multipart.File, filename string) (upload.ID, error) {
	uploadID := uuid.NewString()
	task := &upload.Task{
		ID:        upload.ID(uploadID),
		Status:    upload.StatusProcessing,
		Message:   upload.MessageProcessing,
		Filename:  filename,
		StartedAt: time.Now(),
	}

	err := uc.uploadRepo.Save(ctx, task)
	if err != nil {
		log.Info(ctx, fmt.Sprint("save upload task error:", err.Error()))
		return "", err
	}

	go uc.processStatement(uc.appCtx, task.ID, file)
	return task.ID, nil
}

func (uc *statement) processStatement(ctx context.Context, uploadID upload.ID, file multipart.File) {
	defer file.Close()
	csvReader := csv.NewReader(file)

	if _, err := csvReader.Read(); err != nil {
		uc.markUploadAsFailed(ctx, uploadID, "failed to read CSV header: "+err.Error())
		return
	}

	lineNumber := 1
	for {
		select {
		case <-ctx.Done():
			uc.markUploadAsFailed(ctx, uploadID, "processing cancelled")
			return
		default:
		}

		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			uc.markUploadAsFailed(ctx, uploadID, fmt.Sprintf("error at line %d: %v", lineNumber, err))
			return
		}

		lineNumber++
		//log.Info(ctx, fmt.Sprint(lineNumber, " record:", record, "len:", len(record)))

		t, err := uc.parseTransaction(record, uploadID)
		if err != nil {
			uc.markUploadAsFailed(ctx, uploadID, fmt.Sprintf("invalid data at line %d: %v", lineNumber, err))
			return
		}

		if err := uc.transactionRepo.Save(ctx, t); err != nil {
			uc.markUploadAsFailed(ctx, uploadID, fmt.Sprintf("failed to save transaction at line %d: %v", lineNumber, err))
			return
		}

		if t.Status == transaction.StatusFailed {
			failedEvent := event.NewFailedTransactionEvent(t.ID, uploadID, t.Counterparty, t.Description, t.Timestamp, t.Amount)
			uc.eventBus.Publish(failedEvent)
		}
	}

	//uc.transactionRepo.PrepareDataForFilters(ctx, uploadID)
	uc.markUploadAsCompleted(ctx, uploadID)
}

func (uc *statement) parseTransaction(record []string, uploadID upload.ID) (*transaction.Transaction, error) {
	if len(record) != 6 {
		return nil, fmt.Errorf("invalid CSV format: expected 6 columns, got %d", len(record))
	}

	timestampStr := strings.TrimSpace(record[0])
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp '%s': %w", timestampStr, err)
	}

	counterparty := strings.TrimSpace(record[1])
	if counterparty == "" {
		return nil, errors.New("counterparty cannot be empty")
	}

	transactionType := transaction.Type(strings.ToUpper(strings.TrimSpace(record[2])))
	if transactionType != transaction.TypeCredit && transactionType != transaction.TypeDebit {
		return nil, fmt.Errorf("invalid transaction type '%s': must be CREDIT or DEBIT", transactionType)
	}

	amountStr := strings.TrimSpace(record[3])
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid amount '%s': %w", amountStr, err)
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive, got %d", amount)
	}

	status := transaction.Status(strings.ToUpper(strings.TrimSpace(record[4])))
	if status != transaction.StatusSuccess && status != transaction.StatusFailed && status != transaction.StatusPending {
		return nil, fmt.Errorf("invalid status '%s': must be SUCCESS, FAILED, or PENDING", status)
	}

	description := strings.TrimSpace(record[5])
	if description == "" {
		return nil, errors.New("description cannot be empty")
	}

	return &transaction.Transaction{
		ID:           transaction.ID(uuid.NewString()),
		UploadID:     uploadID,
		Timestamp:    timestamp,
		Counterparty: counterparty,
		Type:         transactionType,
		Amount:       amount,
		Status:       status,
		Description:  description,
	}, nil
}

func (uc *statement) markUploadAsFailed(ctx context.Context, uploadID upload.ID, reason string) {
	info := &upload.Task{
		ID:          uploadID,
		Status:      upload.StatusFailed,
		Message:     reason,
		CompletedAt: time.Now(),
	}

	err := uc.uploadRepo.Update(ctx, info)
	if err != nil {
		log.Info(ctx, fmt.Sprint("failed to mark upload as failed:", err.Error()))

	}
}

func (uc *statement) markUploadAsCompleted(ctx context.Context, uploadID upload.ID) {
	info := &upload.Task{
		ID:          uploadID,
		Status:      upload.StatusCompleted,
		Message:     "",
		CompletedAt: time.Now(),
	}

	err := uc.uploadRepo.Update(ctx, info)
	if err != nil {
		log.Info(ctx, fmt.Sprint("failed to mark upload as failed:", err.Error()))
	}
}
