package integration_test

import (
	"database/sql"
	"local-finance/internal/db"
	"path/filepath"
	"testing"
)

func TestINDmoneyPreviewValuationMigration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preview.db")
	database, err := db.NewDB(path)
	if err != nil {
		t.Fatal(err)
	}
	database.Close()
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate a version-14 preview with entirely fictional holdings.
	legacy := `{"id":"preview-demo","parser_id":"indmoney_us_holdings_xls_v1","provider":"INDmoney","account_ref":"DEMO-US-01","as_of":"2026-04-01","currency":"USD","filename":"fictional.xls","imported_at":"2026-04-02T00:00:00Z","invested_value":20,"current_value":null,"unrealized_return":null,"return_percent":null,"holdings":[{"symbol":"DEMO-A","quantity":2,"average_price":10,"closing_price":null,"invested_value":20,"current_value":null,"unrealized_return":null,"return_percent":null,"fields":{"Total Value ($)":"20"}}],"sheets":[],"warnings":[]}`
	if _, err := conn.Exec(`INSERT INTO investment_snapshots VALUES('preview-demo','INDmoney','DEMO-US-01','2026-04-01','demo-hash','2026-04-02T00:00:00Z',?)`, legacy); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`DELETE FROM goose_db_version WHERE version_id=15`); err != nil {
		t.Fatal(err)
	}
	conn.Close()
	database, err = db.NewDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	list, err := database.ListInvestmentSnapshots()
	if err != nil || len(list) != 1 {
		t.Fatal("migration lost snapshot", err)
	}
	s := list[0]
	if s.ID != "preview-demo" || s.Filename != "fictional.xls" || s.ImportedAt != "2026-04-02T00:00:00Z" || s.InvestedValue != nil || s.CurrentValue == nil || *s.CurrentValue != 20 || s.UnrealizedReturn != nil {
		t.Fatalf("incorrect migrated snapshot: %+v", s)
	}
	h := s.Holdings[0]
	if h.InvestedValue != nil || h.AveragePrice != nil || h.CurrentValue == nil || *h.CurrentValue != 20 || h.ClosingPrice == nil || *h.ClosingPrice != 10 || h.Fields["Total Value ($)"] != "20" {
		t.Fatalf("holding/source data changed: %+v", h)
	}
}
