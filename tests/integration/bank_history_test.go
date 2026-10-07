package integration_test

import (
	"bytes"
	"testing"

	"local-finance/internal/service"
)

func TestBankHistoryIdentifiesDistinctAccountsAndIdempotentImports(t *testing.T) {
	for _, bank := range []string{"ICICI Bank", "Union Bank of India"} {
		t.Run(bank, func(t *testing.T) {
			database := testDatabase(t)
			svc := service.NewTransactionService(database)
			var previousAccountID string
			for _, number := range []string{"123456781234", "123456785678", "987654321234"} {
				data := syntheticBankHistoryPDF(t, bank, number)
				// Large metadata makes this a valid PDF beyond the old 64KB
				// detection slice while retaining compressed content streams.
				if len(data) <= 65536 {
					t.Fatal("synthetic PDF must exceed the former detection limit")
				}
				preview, err := svc.PreviewStatement("statement.pdf", bytes.NewReader(data), "", "", "")
				if err != nil {
					t.Fatal(err)
				}
				if preview.Error != "" || preview.AccountNumberMask != "XX"+number[len(number)-4:] || preview.ExistingTxsCount != 0 || preview.NewTxsCount != 2 {
					t.Fatalf("preview lost account identity: %+v", preview)
				}
				result, err := svc.ImportStatement("statement.pdf", bytes.NewReader(data), "", "", "")
				if err != nil {
					t.Fatal(err)
				}
				if result.InsertedCount != 2 || result.AccountID == previousAccountID {
					t.Fatalf("distinct accounts were merged: %+v", result)
				}
				previousAccountID = result.AccountID
				repeat, err := svc.ImportStatement("statement.pdf", bytes.NewReader(data), "", "", "")
				if err != nil {
					t.Fatal(err)
				}
				if repeat.AccountID != result.AccountID || repeat.InsertedCount != 0 || repeat.DuplicateCount != 2 {
					t.Fatalf("reimport changed account or duplicated transactions: %+v", repeat)
				}
			}
			accounts, err := database.ListAccounts()
			if err != nil {
				t.Fatal(err)
			}
			if len(accounts) != 3 {
				t.Fatalf("got %d accounts; want 3", len(accounts))
			}
		})
	}
}
