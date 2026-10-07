package integration_test

import (
	"path/filepath"
	"testing"
	"time"

	"local-finance/internal/db"
	"local-finance/internal/models"
)

func TestDetectedTransfersUseTransferCategoryAndPreserveManualOverrides(t *testing.T) {
	path := filepath.Join(t.TempDir(), "transfers.db")
	database, err := db.NewDB(path)
	if err != nil {
		t.Fatal(err)
	}
	account, err := database.GetOrCreateAccount("Test Bank", models.AccountTypeSavings, "", "XX1001", "", "", "", "", "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	other := "cat_others"
	for _, manual := range []bool{false, true} {
		hash := "automatic"
		if manual {
			hash = "manual"
		}
		tx := &models.Transaction{AccountID: account.ID, TxHash: hash, TxDate: "2026-04-01", TxType: models.TxTypeDebit, Amount: 500, IsTransfer: true, IsManualCategory: manual, CategoryID: &other, Notes: "keep note"}
		if _, err := database.UpsertTransaction(tx); err != nil {
			t.Fatal(err)
		}
		want := "cat_transfers"
		if manual {
			want = other
		}
		if tx.CategoryID == nil || *tx.CategoryID != want {
			t.Fatalf("category: %+v", tx)
		}
	}
	// Simulate transfers flagged by an older version before reopening the app.
	if _, err := database.Exec("UPDATE transactions SET category_id='cat_others' WHERE tx_hash='automatic'"); err != nil {
		t.Fatal(err)
	}
	database.Close()
	database, err = db.NewDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	rows, _, err := database.ListTransactions(db.TransactionFilter{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		want := "cat_transfers"
		if row.TxHash == "manual" {
			want = other
		}
		if row.CategoryID == nil || *row.CategoryID != want || row.Notes != "keep note" {
			t.Fatalf("existing transfer classification lost: %+v", row)
		}
	}
}

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
