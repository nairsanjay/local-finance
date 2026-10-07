package investment

import (
	"encoding/json"
	"local-finance/internal/models"
	"math"
	"os"
	"strings"
	"testing"
)

// Fictional report rows; private account data must never become test fixtures.
func indmoneySheets() []models.InvestmentSheet {
	return []models.InvestmentSheet{{Name: "HOLDINGS_BOOK", Rows: [][]string{
		{"Account Details"}, {"Broker Name", "Example US Broker"}, {"Broker Account", "DEMO-US-01"}, {"Holdings as on", "2026-04-01"}, {},
		indmoneyHeaders,
		{"DEMO-A", "01 Mar 2026, 10:30 AM", "0.123456789", "100", "12.345679"},
		{"DEMO-B", "02 Mar 2026, 10:30 AM", "2.5", "20", "50"}, {},
		{"Disclaimer:-"}, {"INDmoney report information"},
	}}}
}

func TestINDmoneyCurrentValuationAndUnavailableCost(t *testing.T) {
	sheets := indmoneySheets()
	snapshot, err := parseINDmoneyHoldings(sheets)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Provider != "INDmoney" || snapshot.Currency != "USD" || snapshot.AccountRef != "DEMO-US-01" || snapshot.AsOf != "2026-04-01" || len(snapshot.Holdings) != 2 || snapshot.CurrentValue == nil || math.Abs(*snapshot.CurrentValue-62.345679) > 1e-9 {
		t.Fatalf("unexpected portfolio: %+v", snapshot)
	}
	h := snapshot.Holdings[0]
	if h.Quantity != 0.123456789 || h.Fields["Holding Since"] != "01 Mar 2026, 10:30 AM" || h.ISIN != "" {
		t.Fatalf("source precision or fields lost: %+v", h)
	}
	if snapshot.InvestedValue != nil || snapshot.UnrealizedReturn != nil || snapshot.ReturnPercent != nil || h.AveragePrice != nil || h.InvestedValue != nil || h.UnrealizedReturn != nil || h.ReturnPercent != nil {
		t.Fatal("missing acquisition costs or returns fabricated")
	}
	if h.ClosingPrice == nil || *h.ClosingPrice != 100 || h.CurrentValue == nil || *h.CurrentValue != 12.345679 {
		t.Fatal("statement valuation lost")
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"invested_value":null`, `"average_price":null`, `"unrealized_return":null`, `"return_percent":null`, `"rows":[`} {
		if !strings.Contains(string(data), expected) {
			t.Fatalf("missing JSON %s", expected)
		}
	}
	if len(snapshot.Warnings) == 0 || len(snapshot.Sheets[0].Rows) != len(sheets[0].Rows) {
		t.Fatal("missing source details")
	}
}

func TestINDmoneyRejectsInvalidReports(t *testing.T) {
	for _, tc := range []struct {
		name   string
		modify func([]models.InvestmentSheet)
	}{
		{"wrong provider", func(s []models.InvestmentSheet) { s[0].Rows[10][0] = "Other provider" }},
		{"missing header", func(s []models.InvestmentSheet) { s[0].Rows[5] = []string{"Other columns"} }},
		{"missing account", func(s []models.InvestmentSheet) { s[0].Rows[2][1] = "" }},
		{"invalid date", func(s []models.InvestmentSheet) { s[0].Rows[3][1] = "2026-99-01" }},
		{"bad quantity", func(s []models.InvestmentSheet) { s[0].Rows[6][2] = "NaN" }},
		{"negative cost", func(s []models.InvestmentSheet) { s[0].Rows[6][4] = "-1" }},
		{"inconsistent value", func(s []models.InvestmentSheet) { s[0].Rows[6][4] = "500" }},
		{"duplicate symbol", func(s []models.InvestmentSheet) { s[0].Rows[7][0] = "demo-a" }},
		{"missing symbol", func(s []models.InvestmentSheet) { s[0].Rows[6][0] = "" }},
		{"truncated holding", func(s []models.InvestmentSheet) { s[0].Rows[6] = []string{"DEMO-A"} }},
		{"inconsistent account", func(s []models.InvestmentSheet) {
			s[0].Rows = append(s[0].Rows, []string{"Broker Account", "DEMO-US-02"})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sheets := indmoneySheets()
			tc.modify(sheets)
			if _, err := parseINDmoneyHoldings(sheets); err == nil {
				t.Fatal("invalid report accepted")
			}
		})
	}
	for _, data := range [][]byte{nil, []byte("not Excel"), {0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1}} {
		if (INDmoneyHoldingsParser{}).CanParse("report.xls", data) {
			t.Fatal("invalid workbook detected")
		}
		if _, err := (INDmoneyHoldingsParser{}).Parse(data); err == nil {
			t.Fatal("invalid workbook accepted")
		}
	}
}

func TestExistingInvestmentJSONValuationCompatibility(t *testing.T) {
	// Stored snapshots predating optional valuations still contain JSON numbers.
	var snapshot models.InvestmentSnapshot
	if err := json.Unmarshal([]byte(`{"current_value":900,"unrealized_return":100,"holdings":[{"closing_price":90,"current_value":270,"unrealized_return":-30}]}`), &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.CurrentValue == nil || *snapshot.CurrentValue != 900 || snapshot.Holdings[0].ClosingPrice == nil || *snapshot.Holdings[0].ClosingPrice != 90 {
		t.Fatal("previous snapshots lost valuations")
	}
}

func TestINDmoneyLegacyWorkbook(t *testing.T) {
	data, err := os.ReadFile("testdata/indmoney-fictional.xls")
	if err != nil {
		t.Fatal(err)
	}
	p := INDmoneyHoldingsParser{}
	if !p.CanParse("renamed.xls", data) || p.CanParse("renamed.xlsx", data) {
		t.Fatal("legacy detection depends on filename or ignores extension")
	}
	snapshot, err := DefaultRegistry.Parse("renamed.xls", data)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.ParserID != p.ID() || len(snapshot.Holdings) != 2 || len(snapshot.Sheets) != 2 || snapshot.Holdings[0].Quantity != 0.123456789 || snapshot.InvestedValue != nil || snapshot.CurrentValue == nil {
		t.Fatalf("bad legacy workbook: %+v", snapshot)
	}
	if len(snapshot.Sheets[0].Rows[4]) != 0 || snapshot.Sheets[1].Rows[0][0] != "Fictional test data only" {
		t.Fatal("blank rows or additional sheets lost")
	}
}

func TestInvestmentRKNumbers(t *testing.T) {
	for _, tc := range []struct {
		encoded uint32
		want    float64
	}{
		{100<<2 | 2, 100}, {250<<2 | 3, 2.5}, {uint32(0xffffff9c) | 2, -25},
		{uint32(math.Float64bits(1.25) >> 32), 1.25},
	} {
		if got := decodeInvestmentRK(tc.encoded); got != tc.want {
			t.Fatalf("RK %x: got %v, want %v", tc.encoded, got, tc.want)
		}
	}
}
