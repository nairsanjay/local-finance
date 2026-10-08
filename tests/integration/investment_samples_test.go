package integration_test

import (
	"bytes"
	"math"
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
	if snapshot.Holdings[0].AssetClass != "Equity" || snapshot.Holdings[1].AssetClass != "Mutual Fund" {
		t.Fatalf("sample asset classes incorrect: %+v", snapshot.Holdings)
	}
	repeated, duplicate, err := svc.Import("renamed.xlsx", bytes.NewReader(data))
	if err != nil || !duplicate || repeated.ID != snapshot.ID {
		t.Fatalf("sample reimport duplicated snapshot: %+v %v %v", repeated, duplicate, err)
	}
}

func TestCheckedInINDmoneySampleImport(t *testing.T) {
	data, err := os.ReadFile("../../samples/investments/indmoney-fictional.xls")
	if err != nil {
		t.Fatal(err)
	}
	svc := service.NewInvestmentService(testDatabase(t))
	preview, err := svc.Preview("generic.xls", bytes.NewReader(data))
	if err != nil || preview.ID != "" {
		t.Fatalf("sample preview failed: %+v %v", preview, err)
	}
	snapshot, duplicate, err := svc.Import("indmoney-fictional.xls", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if duplicate || snapshot.Provider != "INDmoney" || snapshot.AccountRef != "DEMO-US-01" || snapshot.AsOf != "2026-04-01" || snapshot.Currency != "USD" || len(snapshot.Holdings) != 2 || len(snapshot.Sheets) != 2 || snapshot.CurrentValue == nil || math.Abs(*snapshot.CurrentValue-62.345679) > 1e-9 {
		t.Fatalf("sample differs from documented portfolio: %+v", snapshot)
	}
	if snapshot.InvestedValue != nil || snapshot.UnrealizedReturn != nil || snapshot.ReturnPercent != nil {
		t.Fatal("sample fabricated unavailable acquisition costs or returns")
	}
	first := snapshot.Holdings[0]
	if first.Quantity != 0.123456789 || first.Fields["Holding Since"] != "01 Mar 2026, 10:30 AM" || first.CurrentValue == nil || math.Abs(*first.CurrentValue-12.345679) > 1e-9 {
		t.Fatalf("sample lost fractional shares, valuation, or provider fields: %+v", first)
	}
	for _, holding := range snapshot.Holdings {
		if holding.AveragePrice != nil || holding.InvestedValue != nil || holding.UnrealizedReturn != nil || holding.ReturnPercent != nil {
			t.Fatal("holding fabricated unavailable costs or returns")
		}
	}
	repeated, duplicate, err := svc.Import("renamed.xls", bytes.NewReader(data))
	if err != nil || !duplicate || repeated.ID != snapshot.ID {
		t.Fatalf("sample reimport duplicated snapshot: %+v %v %v", repeated, duplicate, err)
	}
}
