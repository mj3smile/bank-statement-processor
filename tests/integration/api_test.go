package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/mj3smile/bank-statement-processor/internal/event"
	"github.com/mj3smile/bank-statement-processor/internal/event/consumer"
	handler "github.com/mj3smile/bank-statement-processor/internal/handler/http"
	"github.com/mj3smile/bank-statement-processor/internal/infra/log"
	"github.com/mj3smile/bank-statement-processor/internal/infra/server"
	"github.com/mj3smile/bank-statement-processor/internal/model/upload"
	repository "github.com/mj3smile/bank-statement-processor/internal/repository/memory"
	"github.com/mj3smile/bank-statement-processor/internal/usecase"
)

func TestFullWorkflow_UploadProcessQuery(t *testing.T) {
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

	csvContent := `timestamp,counterparty,type,amount,status,description
1674507883,JOHN DOE,DEBIT,250000,SUCCESS,restaurant
1674508123,ACME CORP,CREDIT,1500000,SUCCESS,salary
1674508456,JANE SMITH,DEBIT,75000,FAILED,payment failed
1674508789,BOB BROWN,DEBIT,100000,PENDING,processing
1674509012,ALICE GREEN,CREDIT,500000,SUCCESS,consulting`

	// 1. upload CSV
	var uploadID string
	t.Run("Upload CSV", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "test.csv")
		if err != nil {
			t.Errorf("error while creating csv file: %v", err)
			return
		}

		_, err = io.WriteString(part, csvContent)
		if err != nil {
			t.Errorf("error while writing csv content to file: %v", err)
			return
		}
		writer.Close()

		req := httptest.NewRequest("POST", "/statements", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if !reflect.DeepEqual(http.StatusAccepted, w.Code) {
			t.Errorf("http response: status code: got = %v, want %v", w.Code, http.StatusAccepted)
		}

		var response map[string]string
		json.NewDecoder(w.Body).Decode(&response)
		uploadID = response["upload_id"]
		if uploadID == "" {
			t.Errorf("upload_id is empty")
		}
	})

	// wait for async processing
	time.Sleep(500 * time.Millisecond)

	// 2. get balance
	t.Run("Get Balance", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/balance?upload_id="+uploadID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !reflect.DeepEqual(http.StatusOK, w.Code) {
			t.Errorf("http response status code: got = %v, want %v", w.Code, http.StatusOK)
		}

		var response handler.GetBalanceResponse
		json.NewDecoder(w.Body).Decode(&response)

		if !reflect.DeepEqual(upload.StatusCompleted, upload.Status(response.Status)) {
			t.Errorf("http response: field status: got = %v, want %v", response.Status, upload.StatusCompleted)
		}

		if response.Balance == nil {
			t.Error("http response: field balance is nil")
		}

		// 1500000 (credit) + 500000 (credit) - 250000 (debit) = 1750000
		if !reflect.DeepEqual(int64(1750000), *response.Balance) {
			t.Errorf("http response: field balance: got = %v, want %v", *response.Balance, int64(1750000))
		}
	})

	// 3. get issues
	t.Run("Get Issues", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/transactions/issues?upload_id="+uploadID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !reflect.DeepEqual(http.StatusOK, w.Code) {
			t.Errorf("http response status code: got = %v, want %v", w.Code, http.StatusOK)
		}

		var response handler.GetIssuesResponse
		json.NewDecoder(w.Body).Decode(&response)

		// 1 FAILED + 1 PENDING
		if !reflect.DeepEqual(2, response.Pagination.TotalItems) {
			t.Errorf("http response: field total_items: got = %v, want %v", response.Pagination.TotalItems, 2)
		}

		if len(response.Transactions) != 2 {
			t.Errorf("http response: field transactions: got = %v, want %v", len(response.Transactions), 2)
		}
	})

	// 4. reconcile failed transactions
	t.Run("Verify Reconciliation", func(t *testing.T) {
		time.Sleep(5 * time.Second) // wait for reconciliation

		processedCount := reconciliationConsumer.GetProcessedCount()
		if processedCount == 0 {
			t.Errorf("reconciliation count: got = %v, want >%v", processedCount, 0)
		}

		log.Info(appCtx, fmt.Sprint("processed count:", processedCount))
	})
}
