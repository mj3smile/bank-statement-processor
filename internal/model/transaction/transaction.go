package transaction

import "github.com/mj3smile/bank-statement-processor/internal/model/upload"

type (
	ID     string
	Type   string
	Status string
)

const (
	TypeCredit Type = "CREDIT"
	TypeDebit  Type = "DEBIT"

	StatusSuccess Status = "SUCCESS"
	StatusFailed  Status = "FAILED"
	StatusPending Status = "PENDING"
)

type Transaction struct {
	ID           ID
	UploadID     upload.ID
	Timestamp    int64
	Counterparty string
	Type         Type
	Amount       int64
	Status       Status
	Description  string
}

type IssuesFilters struct {
	UploadID  upload.ID
	Status    *Status
	MinAmount *int64
	MaxAmount *int64
	FromDate  *int64
	ToDate    *int64
	Page      int
	PageSize  int
}
