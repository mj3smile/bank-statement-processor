package http

import (
	"net/http"

	"github.com/mj3smile/bank-statement-processor/internal/usecase"
)

type BalanceHandler struct {
	balanceUseCase usecase.Balance
}

func NewBalanceHandler(balanceUseCase usecase.Balance) *BalanceHandler {
	return &BalanceHandler{
		balanceUseCase: balanceUseCase,
	}
}

func (handler *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	uploadID := r.URL.Query().Get(UploadIDParam)
	if uploadID == "" {
		respondError(w, http.StatusBadRequest, "missing upload_id parameter")
		return
	}

	result, err := handler.balanceUseCase.Get(r.Context(), uploadID)
	if err != nil || result == nil {
		respondError(w, http.StatusNotFound, err.Error())
	}

	balanceInfo := *result
	respondJSON(w, http.StatusOK, GetBalanceResponse{
		UploadID: balanceInfo.UploadID,
		Status:   balanceInfo.UploadTaskStatus,
		Balance:  balanceInfo.Balance,
		Message:  balanceInfo.UploadTaskMessage,
	})
}
