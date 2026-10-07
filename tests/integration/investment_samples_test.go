package integration_test

import (
	"bytes"
	"os"
	"testing"

	"local-finance/internal/service"
)

func TestCheckedInZerodhaSampleImport(t *testing.T) {
	data, err := os.ReadFile("../../samples/investments/zerodha-fictional.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	svc := service.NewInvestmentService(testDatabase(t))
	snapshot, duplicate, err := svc.Import("zerodha-fictional.xlsx", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if duplicate || snapshot.Provider != "Zerodha" || snapshot.AccountRef != "DEMO0001" || snapshot.AsOf != "2026-04-01" || snapshot.Currency != "INR" || len(snapshot.Holdings) != 2 || len(snapshot.Sheets) != 2 || snapshot.InvestedValue == nil || *snapshot.InvestedValue != 800 || snapshot.CurrentValue == nil || *snapshot.CurrentValue != 900 || snapshot.UnrealizedReturn == nil || *snapshot.UnrealizedReturn != 100 || snapshot.ReturnPercent == nil || *snapshot.ReturnPercent != 12.5 {
		t.Fatalf("sample differs from documented portfolio: %+v", snapshot)
	}
	repeated, duplicate, err := svc.Import("renamed.xlsx", bytes.NewReader(data))
	if err != nil || !duplicate || repeated.ID != snapshot.ID {
		t.Fatalf("sample reimport duplicated snapshot: %+v %v %v", repeated, duplicate, err)
	}
}
