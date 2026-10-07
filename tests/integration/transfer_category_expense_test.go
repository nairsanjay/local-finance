package integration_test

import (
	"strings"
	"testing"
	"time"

	"local-finance/internal/db"
	"local-finance/internal/models"
	"local-finance/internal/service"
)

func TestTransferCategoryExcludedFromSpending(t *testing.T) {
	database := testDatabase(t)
	account, err := database.GetOrCreateAccount("Test Bank", models.AccountTypeSavings, "", "XX1001", "", "", "", "", "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	category := "cat_transfers"
	for _, tx := range []*models.Transaction{
		{TxHash: "transfer", TxType: models.TxTypeDebit, Amount: 500, CategoryID: &category, CleanedPayee: "Person", Notes: "keep note", Tags: "keep tag"},
		{TxHash: "shop", TxType: models.TxTypeDebit, Amount: 100, CleanedPayee: "Shop"},
		{TxHash: "credit", TxType: models.TxTypeCredit, Amount: 1000, CleanedPayee: "Employer"},
	} {
		tx.AccountID = account.ID
		tx.TxDate = "2026-04-01"
		if _, err := database.UpsertTransaction(tx); err != nil {
			t.Fatal(err)
		}
	}
	// Use the stable category ID, even if the user renames its display label.
	if _, err := database.Exec("UPDATE categories SET name='Renamed transfers' WHERE id=?", category); err != nil {
		t.Fatal(err)
	}
	assertTotals := func(want float64) {
		t.Helper()
		overview, err := database.GetAnalyticsOverview()
		if err != nil {
			t.Fatal(err)
		}
		if overview.TotalExpense != want || overview.TotalIncome != 1000 {
			t.Fatalf("overview: %+v", overview)
		}
		if len(overview.MonthlyTrends) != 1 || overview.MonthlyTrends[0].Expense != want {
			t.Fatalf("trends: %+v", overview.MonthlyTrends)
		}
		wrapped, err := database.GetWrappedStory("2026")
		if err != nil {
			t.Fatal(err)
		}
		if wrapped.TotalExpense != want {
			t.Fatalf("wrapped expense: %v", wrapped.TotalExpense)
		}
		review, err := database.GetMonthlyReview("2026-04", time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatal(err)
		}
		if review.Current.Amount != want {
			t.Fatalf("review expense: %v", review.Current.Amount)
		}
		merchants, err := database.ListMerchants("", "", "total_spend")
		if err != nil {
			t.Fatal(err)
		}
		if merchants.TotalSpend != want {
			t.Fatalf("merchant spend: %v", merchants.TotalSpend)
		}
	}
	assertTotals(100)
	profile, err := database.GetMerchantProfile("Person")
	if err != nil {
		t.Fatal(err)
	}
	if profile.TotalSpend != 0 || len(profile.MonthlySpendHistory) != 0 {
		t.Fatalf("transfer in merchant profile: %+v", profile)
	}
	rows, count, err := database.ListTransactions(db.TransactionFilter{Limit: 10})
	if err != nil || count != 3 {
		t.Fatalf("ledger rows lost: %d %v", count, err)
	}
	var transferID string
	for _, row := range rows {
		if row.TxHash == "transfer" {
			transferID = row.ID
			if row.Notes != "keep note" || row.Tags != "keep tag" || row.IsTransfer {
				t.Fatal("ledger mutated")
			}
		}
	}
	// Reclassification immediately restores the expense, without rewriting flags.
	otherCategory := "cat_others"
	if _, err := database.UpdateTransaction(transferID, models.UpdateTransactionRequest{CategoryID: &otherCategory}); err != nil {
		t.Fatal(err)
	}
	assertTotals(600)
}

func TestTransferCategoryExcludedAfterResetAndImport(t *testing.T) {
	database := testDatabase(t)
	const statement = `Date,Narration,Chq/Ref,Value Dt,Withdrawal,Deposit,Closing Balance
01/04/2026,NEFT CR-ABC123-EMPLOYER-SALARY,ABC123,01/04/2026,,1000,1000
02/04/2026,UPI-ALEX-alex@okhdfcbank-HDFC-123456789012-SELF TRANSFER,DEMO123,02/04/2026,500,,500
03/04/2026,UPI-SHOP-shop@okhdfcbank-HDFC-123456789014-PAYMENT,123456789014,03/04/2026,100,,400
`
	for _, reset := range []bool{false, true} {
		if reset {
			if err := database.ResetDatabase(); err != nil {
				t.Fatal(err)
			}
		}
		svc := service.NewTransactionService(database)
		for attempt := 0; attempt < 2; attempt++ {
			result, err := svc.ImportStatement("hdfc.csv", strings.NewReader(statement), "", "hdfc_savings_csv_v1", "")
			if err != nil {
				t.Fatal(err)
			}
			if attempt == 0 && result.InsertedCount != 3 || attempt == 1 && (result.InsertedCount != 0 || result.DuplicateCount != 3) {
				t.Fatalf("unexpected import counts: %+v", result)
			}
			overview, err := database.GetAnalyticsOverview()
			if err != nil || overview.TotalExpense != 100 || overview.TotalIncome != 1000 {
				t.Fatalf("reset=%v attempt=%d: totals %+v error %v", reset, attempt, overview, err)
			}
			flow, err := database.GetCashFlowIntelligence("2026-04")
			if err != nil || flow.Summary.TotalOutflow != 100 || flow.Sankey.TotalOutflow != 100 {
				t.Fatalf("cash flow includes category transfer: %+v %v", flow, err)
			}
			rows, count, err := database.ListTransactions(db.TransactionFilter{Limit: 10})
			if err != nil || count != 3 {
				t.Fatalf("ledger lost rows: %d %v", count, err)
			}
			for _, row := range rows {
				if row.RawNarration == "UPI-ALEX-alex@okhdfcbank-HDFC-123456789012-SELF TRANSFER" && (row.CategoryID == nil || *row.CategoryID != "cat_transfers" || row.IsTransfer) {
					t.Fatalf("test must cover category-only transfer: %+v", row)
				}
			}
		}
	}
}
