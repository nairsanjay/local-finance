package investment

import (
	"local-finance/internal/models"
	"testing"
)

type demoParser struct{ id string }

func (demoParser) Info() ParserInfo {
	return ParserInfo{Provider: "Example Broker", Name: "Portfolio statement", Extensions: []string{".csv"}}
}

func (p demoParser) ID() string                               { return p.id }
func (demoParser) CanParse(filename string, data []byte) bool { return filename == "demo.csv" }
func (demoParser) Parse(data []byte) (*models.InvestmentSnapshot, error) {
	return &models.InvestmentSnapshot{Provider: "Example Broker", AccountRef: "DEMO", AsOf: "2026-04-01", Currency: "INR", Holdings: []models.InvestmentHolding{}}, nil
}
func TestRegistryAcceptsIndependentProvider(t *testing.T) {
	r := NewRegistry()
	r.Register(demoParser{id: "example_csv_v1"})
	formats := r.List()
	if len(formats) != 1 || formats[0].Provider != "Example Broker" || formats[0].Extensions[0] != ".csv" || formats[0].ID != "example_csv_v1" {
		t.Fatalf("provider capabilities not exposed: %+v", formats)
	}
	snapshot, err := r.Parse("demo.csv", nil)
	if err != nil || snapshot.Provider != "Example Broker" || snapshot.ParserID != "example_csv_v1" {
		t.Fatal("registry depends on Zerodha", err)
	}
	r.Register(demoParser{id: "another_csv_v1"})
	if _, err := r.Parse("demo.csv", nil); err == nil {
		t.Fatal("ambiguous statement accepted")
	}
}
