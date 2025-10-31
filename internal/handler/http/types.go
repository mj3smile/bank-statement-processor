package http

type UploadStatementResponse struct {
	UploadID string `json:"upload_id"`
	Message  string `json:"message,omitempty"`
}

type GetBalanceResponse struct {
	UploadID string `json:"upload_id"`
	Status   string `json:"status"`
	Balance  *int64 `json:"balance,omitempty"`
	Message  string `json:"message,omitempty"`
}

type GetIssuesResponse struct {
	UploadID     string           `json:"upload_id"`
	Transactions []TransactionDTO `json:"transactions"`
	Pagination   PaginationMeta   `json:"pagination"`
}

type TransactionDTO struct {
	ID           string `json:"id"`
	Timestamp    int64  `json:"timestamp"`
	Counterparty string `json:"counterparty"`
	Type         string `json:"type"`
	Amount       int64  `json:"amount"`
	Status       string `json:"status"`
	Description  string `json:"description"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type GetHealthResponse struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

const (
	UploadIDParam = "upload_id"
)
