package db

import (
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
	"local-finance/internal/models"
)

func TestTransferCategoryMigrationRepairsExistingRowsOnce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.db")
	database, err := NewDB(path)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate a database from the preceding application version.
	if err := goose.DownTo(database.conn, "migrations", 14); err != nil {
		t.Fatal(err)
	}
	account, err := database.GetOrCreateAccount("Test Bank", models.AccountTypeSavings, "", "XX1001", "", "", "", "", "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	category := "cat_transfers"
	for _, tx := range []*models.Transaction{
		{TxHash: "debit", TxType: models.TxTypeDebit, Amount: 500, CategoryID: &category, IsManualCategory: true, Notes: "keep note", Tags: "keep tag"},
		{TxHash: "credit", TxType: models.TxTypeCredit, Amount: 500, CategoryID: &category},
		{TxHash: "shop", TxType: models.TxTypeDebit, Amount: 100},
		{TxHash: "income", TxType: models.TxTypeCredit, Amount: 1000},
	} {
		tx.AccountID = account.ID
		tx.TxDate = "2026-04-01"
		if _, err := database.UpsertTransaction(tx); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := database.Exec("UPDATE categories SET name='Renamed transfers' WHERE id=?", category); err != nil {
		t.Fatal(err)
	}
	database.Close()
	database, err = NewDB(path)
	if err != nil {
		t.Fatal(err)
	}
	overview, err := database.GetAnalyticsOverview()
	if err != nil {
		t.Fatal(err)
	}
	if overview.TotalExpense != 100 || overview.TotalIncome != 1000 {
		t.Fatalf("migration did not exclude transfers: %+v", overview)
	}
	rows, count, err := database.ListTransactions(TransactionFilter{Limit: 10})
	if err != nil || count != 4 {
		t.Fatalf("ledger lost rows: %d %v", count, err)
	}
	for _, row := range rows {
		want := row.TxHash == "debit" || row.TxHash == "credit"
		if row.IsTransfer != want {
			t.Fatalf("unexpected flag: %+v", row)
		}
		if row.TxHash == "debit" && (row.CategoryID == nil || *row.CategoryID != category || !row.IsManualCategory || row.Notes != "keep note" || row.Tags != "keep tag") {
			t.Fatal("migration changed user data")
		}
	}
	// A completed migration is not reapplied on subsequent startups.
	if _, err := database.UpsertTransaction(&models.Transaction{AccountID: account.ID, TxHash: "later", TxDate: "2026-04-02", TxType: models.TxTypeDebit, Amount: 50, CategoryID: &category}); err != nil {
		t.Fatal(err)
	}
	database.Close()
	database, err = NewDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	overview, err = database.GetAnalyticsOverview()
	if err != nil || overview.TotalExpense != 150 {
		t.Fatalf("migration reran: %+v %v", overview, err)
	}
}
