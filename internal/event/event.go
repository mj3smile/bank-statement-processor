package event

import (
	"time"

	"github.com/mj3smile/bank-statement-processor/internal/model/transaction"
	"github.com/mj3smile/bank-statement-processor/internal/model/upload"
)

type Event interface {
	EventType() string
	EventTime() time.Time
}

type FailedTransactionEvent struct {
	TransactionID transaction.ID
	UploadID      upload.ID
	Timestamp     int64
	Counterparty  string
	Amount        int64
	Description   string
	PublishedAt   time.Time
}

func (e FailedTransactionEvent) EventType() string {
	return "failed_transaction"
}

func (e FailedTransactionEvent) EventTime() time.Time {
	return e.PublishedAt
}

func NewFailedTransactionEvent(transactionID transaction.ID, uploadID upload.ID, counterparty, description string, timestamp, amount int64) FailedTransactionEvent {
	return FailedTransactionEvent{
		TransactionID: transactionID,
		UploadID:      uploadID,
		Timestamp:     timestamp,
		Counterparty:  counterparty,
		Amount:        amount,
		Description:   description,
		PublishedAt:   time.Now(),
	}
}
