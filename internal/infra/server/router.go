package server

import (
	"net/http"

	handler "github.com/mj3smile/bank-statement-processor/internal/handler/http"
)

func NewRouter(
	statementHandler *handler.StatementHandler,
	balanceHandler *handler.BalanceHandler,
	issuesHandler *handler.IssuesHandler,
	healthHandler *handler.HealthHandler,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler.GetHealth)
	mux.HandleFunc("POST /statements", statementHandler.UploadStatement)
	mux.HandleFunc("GET /balance", balanceHandler.GetBalance)
	mux.HandleFunc("GET /transactions/issues", issuesHandler.GetIssues)

	return handler.Logger(mux)
}
