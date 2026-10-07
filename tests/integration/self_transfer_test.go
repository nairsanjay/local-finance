package integration_test

import (
	"bytes"
	"testing"

	"local-finance/internal/db"
	"local-finance/internal/service"
)

func TestSelfTransfersExcludedFromIncomeAndExpense(t *testing.T) {
	database := testDatabase(t)
	svc := service.NewTransactionService(database)
	statement := selfTransferStatement(t)
	for attempt := 0; attempt < 2; attempt++ {
		result, err := svc.ImportStatement("hdfc.csv", bytes.NewReader(statement), "", "hdfc_savings_csv_v1", "")
		if err != nil {
			t.Fatal(err)
		}
		if attempt == 0 && result.InsertedCount != 4 || attempt == 1 && (result.InsertedCount != 0 || result.DuplicateCount != 4) {
			t.Fatalf("unexpected import counts: %+v", result)
		}
		overview, err := database.GetAnalyticsOverview()
		if err != nil {
			t.Fatal(err)
		}
		if overview.TotalIncome != 1000 || overview.TotalExpense != 100 || overview.NetSavings != 900 {
			t.Fatalf("self transfers affected totals: %+v", overview)
		}
		if len(overview.MonthlyTrends) != 1 || overview.MonthlyTrends[0].Income != 1000 || overview.MonthlyTrends[0].Expense != 100 {
			t.Fatalf("self transfers affected monthly cash flow: %+v", overview.MonthlyTrends)
		}
		if len(overview.TopPayees) != 1 || overview.TopPayees[0].TotalSpent != 100 {
			t.Fatalf("self transfers affected payee spending: %+v", overview.TopPayees)
		}
		category := "cat_transfers"
		_, count, err := database.ListTransactions(db.TransactionFilter{CategoryID: category, Limit: 10})
		if err != nil || count != 2 {
			t.Fatalf("want both self-transfer rows retained; count=%d, err=%v", count, err)
		}
	}
}
