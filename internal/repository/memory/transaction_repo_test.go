package memory

import (
	"context"
	"reflect"
	"testing"

	"github.com/mj3smile/bank-statement-processor/internal/model/transaction"
	"github.com/mj3smile/bank-statement-processor/internal/model/upload"
)

func Test_transactionRepository_GetBalanceByUploadID(t *testing.T) {
	type args struct {
		ctx      context.Context
		uploadID upload.ID
	}

	tests := []struct {
		name         string
		args         args
		want         int64
		transactions []*transaction.Transaction
	}{
		{
			name: "it should return balance as expected when given transactions with status SUCCESS",
			args: args{context.Background(), upload.ID("ABCDEFG")},
			want: 75,
			transactions: []*transaction.Transaction{
				{
					ID:       transaction.ID("1234"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusSuccess,
					Amount:   100,
					Type:     transaction.TypeCredit,
				},
				{
					ID:       transaction.ID("5678"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusSuccess,
					Amount:   25,
					Type:     transaction.TypeDebit,
				},
			},
		},
		{
			name: "it should return balance as expected when given transactions with status SUCCESS, FAILED, and PENDING",
			args: args{context.Background(), upload.ID("ABCDEFG")},
			want: 86,
			transactions: []*transaction.Transaction{
				{
					ID:       transaction.ID("1234"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusSuccess,
					Amount:   125,
					Type:     transaction.TypeCredit,
				},
				{
					ID:       transaction.ID("5678"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusFailed,
					Amount:   25,
					Type:     transaction.TypeDebit,
				},
				{
					ID:       transaction.ID("9102"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusSuccess,
					Amount:   50,
					Type:     transaction.TypeDebit,
				},
				{
					ID:       transaction.ID("2123"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusPending,
					Amount:   25,
					Type:     transaction.TypeDebit,
				},
				{
					ID:       transaction.ID("1123"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusSuccess,
					Amount:   11,
					Type:     transaction.TypeCredit,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewTransactionRepository()
			if len(tt.transactions) > 0 {
				for _, t := range tt.transactions {
					_ = tr.Save(tt.args.ctx, t)
				}
			}

			if got := tr.GetBalanceByUploadID(tt.args.ctx, tt.args.uploadID); got != tt.want {
				t.Errorf("GetBalanceByUploadID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_transactionRepository_GetIssuesWithFilters(t *testing.T) {
	type args struct {
		ctx     context.Context
		filters *transaction.IssuesFilters
	}

	tests := []struct {
		name           string
		args           args
		wantList       []*transaction.Transaction
		wantTotalCount int
		wantErr        bool
		transactions   []*transaction.Transaction
	}{
		{
			name: "it should return empty list when given transactions with status SUCCESS only",
			args: args{
				ctx: context.Background(),
				filters: &transaction.IssuesFilters{
					UploadID: "ABCDEFG",
					Page:     1,
					PageSize: 20,
				},
			},
			transactions: []*transaction.Transaction{
				{
					ID:       transaction.ID("1234"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusSuccess,
					Amount:   100,
					Type:     transaction.TypeCredit,
				},
				{
					ID:       transaction.ID("5678"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusSuccess,
					Amount:   25,
					Type:     transaction.TypeDebit,
				},
			},
			wantList: []*transaction.Transaction{},
		},
		{
			name: "it should return list as expected when given transactions with status SUCCESS, FAILED, and PENDING",
			args: args{
				ctx: context.Background(),
				filters: &transaction.IssuesFilters{
					UploadID: "ABCDEFG",
					Page:     1,
					PageSize: 20,
				},
			},
			transactions: []*transaction.Transaction{
				{
					ID:       transaction.ID("1234"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusSuccess,
					Amount:   125,
					Type:     transaction.TypeCredit,
				},
				{
					ID:       transaction.ID("5678"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusFailed,
					Amount:   25,
					Type:     transaction.TypeDebit,
				},
				{
					ID:       transaction.ID("9102"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusSuccess,
					Amount:   50,
					Type:     transaction.TypeDebit,
				},
				{
					ID:       transaction.ID("2123"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusPending,
					Amount:   25,
					Type:     transaction.TypeDebit,
				},
				{
					ID:       transaction.ID("1123"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusSuccess,
					Amount:   11,
					Type:     transaction.TypeCredit,
				},
			},
			wantList: []*transaction.Transaction{
				{
					ID:       transaction.ID("5678"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusFailed,
					Amount:   25,
					Type:     transaction.TypeDebit,
				},
				{
					ID:       transaction.ID("2123"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusPending,
					Amount:   25,
					Type:     transaction.TypeDebit,
				},
			},
			wantTotalCount: 2,
		},
		{
			name: "it should return list as expected when given filters as defined",
			args: args{
				ctx: context.Background(),
				filters: &transaction.IssuesFilters{
					UploadID:  "ABCDEFG",
					Page:      2,
					PageSize:  1,
					MinAmount: int64Ptr(20),
					MaxAmount: int64Ptr(50),
					Status:    transactionStatusPtr(transaction.StatusFailed),
				},
			},
			transactions: []*transaction.Transaction{
				{
					ID:       transaction.ID("1234"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusFailed,
					Amount:   125,
					Type:     transaction.TypeCredit,
				},
				{
					ID:       transaction.ID("5678"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusFailed,
					Amount:   25,
					Type:     transaction.TypeDebit,
				},
				{
					ID:       transaction.ID("1234"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusFailed,
					Amount:   19,
					Type:     transaction.TypeCredit,
				},
				{
					ID:       transaction.ID("9102"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusPending,
					Amount:   50,
					Type:     transaction.TypeDebit,
				},
				{
					ID:       transaction.ID("2123"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusFailed,
					Amount:   25,
					Type:     transaction.TypeDebit,
				},
			},
			wantList: []*transaction.Transaction{
				{
					ID:       transaction.ID("2123"),
					UploadID: upload.ID("ABCDEFG"),
					Status:   transaction.StatusFailed,
					Amount:   25,
					Type:     transaction.TypeDebit,
				},
			},
			wantTotalCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewTransactionRepository()
			if tt.transactions != nil {
				for _, t := range tt.transactions {
					_ = tr.Save(tt.args.ctx, t)
				}
			}

			gotList, gotTotalCount, err := tr.GetIssuesWithFilters(tt.args.ctx, tt.args.filters)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIssuesWithFilters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotList, tt.wantList) {
				t.Errorf("GetIssuesWithFilters() got = %v, want %v", gotList, tt.wantList)
			}
			if gotTotalCount != tt.wantTotalCount {
				t.Errorf("GetIssuesWithFilters() got1 = %v, want %v", gotTotalCount, tt.wantTotalCount)
			}
		})
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}

func transactionStatusPtr(status transaction.Status) *transaction.Status {
	return &status
}
