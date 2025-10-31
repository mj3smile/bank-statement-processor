package usecase

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mj3smile/bank-statement-processor/internal/model/transaction"
	"github.com/mj3smile/bank-statement-processor/internal/model/upload"
)

func Test_statement_parseTransaction(t *testing.T) {
	type args struct {
		record   []string
		uploadID upload.ID
	}

	tests := []struct {
		name    string
		args    args
		want    *transaction.Transaction
		wantErr bool
	}{
		{
			name: "it should return value as expected when given valid transaction format from CSV",
			args: args{
				record: []string{
					"1674507883",
					"JOHN DOE",
					"DEBIT",
					"250000",
					"SUCCESS",
					"restaurant",
				},
				uploadID: upload.ID(uuid.NewString()),
			},
			want: &transaction.Transaction{
				Timestamp:    1674507883,
				Counterparty: "JOHN DOE",
				Type:         "DEBIT",
				Amount:       250000,
				Status:       "SUCCESS",
				Description:  "restaurant",
			},
		},
		{
			name: "it should return error when given invalid csv format with more than 6 columns",
			args: args{
				record: []string{
					"1674507883",
					"JOHN DOE",
					"DEBIT",
					"250000",
					"SUCCESS",
					"restaurant",
					"abc",
				},
				uploadID: upload.ID(uuid.NewString()),
			},
			wantErr: true,
		},
		{
			name: "it should return error when given invalid timestamp in first column",
			args: args{
				record: []string{
					"SHOULD BE TIMESTAMP",
					"JOHN DOE",
					"DEBIT",
					"250000",
					"SUCCESS",
					"restaurant",
				},
				uploadID: upload.ID(uuid.NewString()),
			},
			wantErr: true,
		},
		{
			name: "it should return error when counterpart in second column is empty",
			args: args{
				record: []string{
					"1674507883",
					"",
					"DEBIT",
					"250000",
					"SUCCESS",
					"restaurant",
				},
				uploadID: upload.ID(uuid.NewString()),
			},
			wantErr: true,
		},
		{
			name: "it should return error when transaction type in third column is not DEBIT or CREDIT",
			args: args{
				record: []string{
					"1674507883",
					"JOHN DOE",
					"SHOULD BE DEBIT OR CREDIT",
					"250000",
					"SUCCESS",
					"restaurant",
				},
				uploadID: upload.ID(uuid.NewString()),
			},
			wantErr: true,
		},
		{
			name: "it should return error when amount in fourth column is not number",
			args: args{
				record: []string{
					"1674507883",
					"JOHN DOE",
					"DEBIT",
					"SHOULD BE NUMBER",
					"SUCCESS",
					"restaurant",
				},
				uploadID: upload.ID(uuid.NewString()),
			},
			wantErr: true,
		},
		{
			name: "it should return error when amount in fourth column is negative number",
			args: args{
				record: []string{
					"1674507883",
					"JOHN DOE",
					"DEBIT",
					"-100000",
					"SUCCESS",
					"restaurant",
				},
				uploadID: upload.ID(uuid.NewString()),
			},
			wantErr: true,
		},
		{
			name: "it should return error when status in fifth column is not SUCCESS or FAILED or PENDING",
			args: args{
				record: []string{
					"1674507883",
					"JOHN DOE",
					"DEBIT",
					"100000",
					"SHOULD BE SUCCESS OR FAILED OR PENDING",
					"restaurant",
				},
				uploadID: upload.ID(uuid.NewString()),
			},
			wantErr: true,
		},
		{
			name: "it should return error when description in sixth column is empty",
			args: args{
				record: []string{
					"1674507883",
					"JOHN DOE",
					"DEBIT",
					"100000",
					"SUCCESS",
					"",
				},
				uploadID: upload.ID(uuid.NewString()),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &statement{}
			got, err := uc.parseTransaction(tt.args.record, tt.args.uploadID)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got == nil && tt.want == nil {
				return
			}

			if got == nil && tt.want != nil {
				t.Errorf("parseTransaction() got = %v, want %v", got, tt.want)
			}

			if got.Timestamp != tt.want.Timestamp {
				t.Errorf("parseTransaction() got = %v, want %v", got, tt.want)
			}

			if got.Counterparty != tt.want.Counterparty {
				t.Errorf("parseTransaction() got = %v, want %v", got, tt.want)
			}

			if got.Type != tt.want.Type {
				t.Errorf("parseTransaction() got = %v, want %v", got, tt.want)
			}

			if got.Amount != tt.want.Amount {
				t.Errorf("parseTransaction() got = %v, want %v", got, tt.want)
			}

			if got.Status != tt.want.Status {
				t.Errorf("parseTransaction() got = %v, want %v", got, tt.want)
			}

			if got.Description != tt.want.Description {
				t.Errorf("parseTransaction() got = %v, want %v", got, tt.want)
			}
		})
	}
}
