package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mj3smile/bank-statement-processor/internal/event"
	"github.com/mj3smile/bank-statement-processor/internal/event/consumer"
	handler "github.com/mj3smile/bank-statement-processor/internal/handler/http"
	"github.com/mj3smile/bank-statement-processor/internal/infra/log"
	"github.com/mj3smile/bank-statement-processor/internal/infra/server"
	repository "github.com/mj3smile/bank-statement-processor/internal/repository/memory"
	"github.com/mj3smile/bank-statement-processor/internal/usecase"
)

func main() {
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	eventBus := event.NewBus(appCtx)
	uploadRepo := repository.NewUploadRepository()
	transactionRepo := repository.NewTransactionRepository()
	defer eventBus.Close()

	statementUseCase := usecase.NewStatement(appCtx, transactionRepo, uploadRepo, eventBus)
	balanceUseCase := usecase.NewBalance(transactionRepo, uploadRepo)
	issuesUseCase := usecase.NewIssues(transactionRepo, uploadRepo)

	statementHandler := handler.NewStatementHandler(statementUseCase)
	balanceHandler := handler.NewBalanceHandler(balanceUseCase)
	issuesHandler := handler.NewIssuesHandler(issuesUseCase)
	healthHandler := handler.NewHealthHandler()

	reconciliationConsumer := consumer.NewReconciliationConsumer(eventBus, 3)
	go reconciliationConsumer.Start(appCtx)

	router := server.NewRouter(statementHandler, balanceHandler, issuesHandler, healthHandler)
	addr := ":8080"
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		log.Info(appCtx, fmt.Sprint("starting http server at", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(appCtx, fmt.Sprintf("server failed: %v", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info(appCtx, "shutting down server...")
	appCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal(appCtx, fmt.Sprintf("server forced to shutdown: %v", err))
	}

	log.Info(shutdownCtx, "server exited")
}
