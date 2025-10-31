package http

import (
	"encoding/json"
	"net/http"

	"github.com/mj3smile/bank-statement-processor/internal/usecase"
)

type StatementHandler struct {
	statementUseCase usecase.Statement
}

func NewStatementHandler(uploadStatementUseCase usecase.Statement) *StatementHandler {
	return &StatementHandler{
		statementUseCase: uploadStatementUseCase,
	}
}

func (handler *StatementHandler) UploadStatement(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 100 << 20 // 100 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusRequestEntityTooLarge, "file too large or invalid form data")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "missing or invalid file parameter")
		return
	}
	defer file.Close()

	if !isCSVFile(header.Filename) {
		file.Close()
		respondError(w, http.StatusBadRequest, "file must be a CSV")
		return
	}

	uploadID, err := handler.statementUseCase.Upload(r.Context(), file, header.Filename)
	if err != nil {
		file.Close()
		respondError(w, http.StatusInternalServerError, "failed to process upload: "+err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, UploadStatementResponse{
		UploadID: string(uploadID),
		Message:  "CSV upload accepted and processing started",
	})
}

func isCSVFile(filename string) bool {
	return len(filename) > 4 && filename[len(filename)-4:] == ".csv"
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}
