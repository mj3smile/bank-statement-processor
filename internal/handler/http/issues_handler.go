package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/mj3smile/bank-statement-processor/internal/model/transaction"
	"github.com/mj3smile/bank-statement-processor/internal/model/upload"
	"github.com/mj3smile/bank-statement-processor/internal/usecase"
)

type IssuesHandler struct {
	getIssuesUseCase usecase.Issues
}

func NewIssuesHandler(getIssuesUseCase usecase.Issues) *IssuesHandler {
	return &IssuesHandler{
		getIssuesUseCase: getIssuesUseCase,
	}
}

func (handler *IssuesHandler) GetIssues(w http.ResponseWriter, r *http.Request) {
	uploadID := r.URL.Query().Get("upload_id")
	if uploadID == "" {
		respondError(w, http.StatusBadRequest, "upload_id is required")
		return
	}

	filters, err := handler.parseFilters(r, uploadID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := handler.getIssuesUseCase.GetIssues(r.Context(), filters)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	transactions := make([]TransactionDTO, 0)
	for _, t := range result.Transactions {
		transactions = append(transactions, TransactionDTO{
			ID:           string(t.ID),
			Timestamp:    t.Timestamp,
			Counterparty: t.Counterparty,
			Type:         string(t.Type),
			Amount:       t.Amount,
			Status:       string(t.Status),
			Description:  t.Description,
		})
	}

	response := GetIssuesResponse{
		UploadID:     uploadID,
		Transactions: transactions,
		Pagination: PaginationMeta{
			Page:       filters.Page,
			PageSize:   filters.PageSize,
			TotalItems: result.TotalCount,
			TotalPages: (result.TotalCount + filters.PageSize - 1) / filters.PageSize,
		},
	}

	respondJSON(w, http.StatusOK, response)
}

func (handler *IssuesHandler) parseFilters(r *http.Request, uploadID string) (*transaction.IssuesFilters, error) {
	filters := &transaction.IssuesFilters{
		UploadID: upload.ID(uploadID),
		Page:     1,
		PageSize: 20, // default
	}

	query := r.URL.Query()
	if pageStr := query.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return nil, errors.New("invalid page number")
		}
		filters.Page = page
	}

	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 100 {
			return nil, errors.New("invalid page_size (must be between 1 and 100)")
		}
		filters.PageSize = pageSize
	}

	if statusStr := query.Get("status"); statusStr != "" {
		status := transaction.Status(strings.ToUpper(statusStr))
		if status != transaction.StatusFailed && status != transaction.StatusPending {
			return nil, errors.New("status must be FAILED or PENDING")
		}
		filters.Status = &status
	}

	if minAmountStr := query.Get("min_amount"); minAmountStr != "" {
		minAmount, err := strconv.ParseInt(minAmountStr, 10, 64)
		if err != nil || minAmount < 0 {
			return nil, errors.New("invalid min_amount")
		}
		filters.MinAmount = &minAmount
	}

	if maxAmountStr := query.Get("max_amount"); maxAmountStr != "" {
		maxAmount, err := strconv.ParseInt(maxAmountStr, 10, 64)
		if err != nil || maxAmount < 0 {
			return nil, errors.New("invalid max_amount")
		}
		filters.MaxAmount = &maxAmount
	}

	if filters.MinAmount != nil && filters.MaxAmount != nil && *filters.MinAmount > *filters.MaxAmount {
		return nil, errors.New("min_amount must be less than max_amount")
	}

	if fromDateStr := query.Get("from_date"); fromDateStr != "" {
		fromDate, err := strconv.ParseInt(fromDateStr, 10, 64)
		if err != nil || fromDate < 0 {
			return nil, errors.New("invalid from_date")
		}
		filters.FromDate = &fromDate
	}

	if toDateStr := query.Get("to_date"); toDateStr != "" {
		toDate, err := strconv.ParseInt(toDateStr, 10, 64)
		if err != nil || toDate < 0 {
			return nil, errors.New("invalid to_date")
		}
		filters.ToDate = &toDate
	}

	if filters.FromDate != nil && filters.ToDate != nil && *filters.FromDate > *filters.ToDate {
		return nil, errors.New("from_date must be before to_date")
	}

	return filters, nil
}
