package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"local-finance/internal/db"
)

func testDatabase(t *testing.T) *db.DB {
	t.Helper()
	database, err := db.NewDB(filepath.Join(t.TempDir(), "ledger.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

func syntheticSamplePDF(t *testing.T, filename string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "samples", "savings", filename))
	if err != nil {
		t.Fatalf("read synthetic PDF fixture: %v", err)
	}
	return data
}

// Read checked-in fictional statements, including distinct full account numbers
// and large metadata that exercises detection beyond the former 64KB limit.
func bankHistorySamplePDF(t *testing.T, bank, account string) []byte {
	t.Helper()
	prefix := "ICICI"
	if bank == "Union Bank of India" {
		prefix = "Union_Bank"
	}
	data, err := os.ReadFile(filepath.Join("testdata", "bank_history", prefix+"_"+account+".pdf"))
	if err != nil {
		t.Fatalf("read bank history PDF fixture: %v", err)
	}
	return data
}
