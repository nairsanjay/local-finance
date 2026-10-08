package parser

import (
	"bytes"
	"encoding/json"
	"slices"
	"testing"

	"github.com/jung-kurt/gofpdf"
	"local-finance/internal/models"
	"local-finance/internal/parser/extractor"
)

func TestUnionAdjacentBaselinesAndSplitYear(t *testing.T) {
	rows := syntheticFragmentedUnionRows()
	data := syntheticBaselinePDF(t, rows)
	extracted, err := extractor.ExtractPDFPositionalRows(bytes.NewReader(data), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(extracted) < 8 {
		t.Fatalf("fixture must retain separate physical baselines, got %d", len(extracted))
	}
	transactions, meta, err := (&UnionSavingsPDFParser{}).Parse(bytes.NewReader(data), ParseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 2 || transactions[0].Date != "2026-04-02" || transactions[0].TxType != models.TxTypeDebit || transactions[0].Amount != 100 || transactions[1].Date != "2026-04-03" || transactions[1].TxType != models.TxTypeCredit || transactions[1].Amount != 250 {
		t.Fatalf("fragmented records handled incorrectly: %+v", transactions)
	}
	if meta.OpeningBalance != 1000 || meta.ClosingBalance != 1150 {
		t.Fatalf("fragmented balances do not reconcile: %+v", meta)
	}
}

func TestBankBaselineCoalescingRejectsConflictsAndCrossPageDates(t *testing.T) {
	rows := syntheticFragmentedUnionRows()
	rows = slices.Insert(rows, 4, extractor.PositionalRow{Page: 1, Y: 504, Elements: []extractor.PositionalElement{{X: 370, S: "999.00"}}})
	if _, _, err := parseBankHistoryRows(rows, unionSavingsProfile); err == nil {
		t.Fatal("conflicting financial fragments were merged")
	}
	rows = syntheticFragmentedUnionRows()
	rows[4].Page = 2
	if _, _, err := parseBankHistoryRows(rows, unionSavingsProfile); err == nil {
		t.Fatal("a date was inferred across page boundaries")
	}
	rows = syntheticFragmentedUnionRows()
	rows[4].Elements[0].S = "reference"
	if _, _, err := parseBankHistoryRows(rows, unionSavingsProfile); err == nil {
		t.Fatal("a date was inferred without an exact four-digit year fragment")
	}
}

func TestUnionMergedSerialDateAndNeighboringNarration(t *testing.T) {
	base := syntheticFragmentedUnionRows()
	rows := append(slices.Clone(base[:2]),
		extractor.PositionalRow{Page: 1, Y: 512, Elements: []extractor.PositionalElement{{X: 105, S: "Example debit"}}},
		extractor.PositionalRow{Page: 1, Y: 506.5, Elements: []extractor.PositionalElement{{X: 35.5, S: "1 02-04-2026"}, {X: 370, S: "100.00"}, {X: 517, S: "900.00 CR"}}},
		extractor.PositionalRow{Page: 1, Y: 490, Elements: []extractor.PositionalElement{{X: 105, S: "Example credit"}}},
		extractor.PositionalRow{Page: 1, Y: 484.5, Elements: []extractor.PositionalElement{{X: 35.5, S: "2 03-04-2026"}, {X: 450, S: "250.00"}, {X: 517, S: "1,150.00 CR"}}},
	)
	data := syntheticBaselinePDF(t, rows)
	transactions, meta, err := (&UnionSavingsPDFParser{}).Parse(bytes.NewReader(data), ParseOptions{})
	if err != nil {
		extracted, _ := extractor.ExtractPDFPositionalRows(bytes.NewReader(data), "")
		t.Fatalf("%v; synthetic rows=%+v", err, extracted)
	}
	if len(transactions) != 2 || transactions[0].Date != "2026-04-02" || transactions[0].RawNarration != "Example debit" || transactions[1].Date != "2026-04-03" || transactions[1].RawNarration != "Example credit" || meta.OpeningBalance != 1000 || meta.ClosingBalance != 1150 {
		t.Fatalf("merged serial/date record handled incorrectly: transactions=%+v meta=%+v", transactions, meta)
	}
	columns, _ := discoverBankHistoryColumns(base[1], unionSavingsProfile)
	for _, element := range []extractor.PositionalElement{
		{X: 150, S: "1 02-04-2026", EndX: 205},  // particulars, not SI/date
		{X: 35.5, S: "1 31-04-2026", EndX: 95},  // invalid date
		{X: 35.5, S: "1 02-04-2026", EndX: 180}, // crosses into particulars
		{X: 35.5, S: "1 02-04-2026"},            // unmeasured extent
	} {
		if date := bankHistoryMergedSerialDate(element, columns); date != "" {
			t.Fatalf("ambiguous element supplied a transaction date: %+v", element)
		}
	}
}

func TestUnionNarrationStraddlesOneFinancialAnchor(t *testing.T) {
	base := syntheticFragmentedUnionRows()
	rows := slices.Clone(base[:2])
	for index, entry := range []struct {
		amount, balance string
		x               float64
	}{
		{"100.00", "900.00 CR", 370}, {"250.00", "1,150.00 CR", 450},
	} {
		anchor := 506.8 - float64(index)*22
		date := "1 02-04-2026"
		if index == 1 {
			date = "2 03-04-2026"
		}
		rows = append(rows,
			extractor.PositionalRow{Page: 1, Y: anchor + 5.5, Elements: []extractor.PositionalElement{{X: 105, EndX: 227.3, S: "Example payment"}}},
			extractor.PositionalRow{Page: 1, Y: anchor, Elements: []extractor.PositionalElement{{X: 35.5, EndX: 91.9, S: date}, {X: entry.x, S: entry.amount}, {X: 517, S: entry.balance}}},
			extractor.PositionalRow{Page: 1, Y: anchor - 5.5, Elements: []extractor.PositionalElement{{X: 105, EndX: 200.6, S: "continued narration"}}},
		)
	}
	transactions, _, err := parseBankHistoryRows(rows, unionSavingsProfile)
	if err != nil || len(transactions) != 2 || transactions[0].RawNarration != "Example payment continued narration" || transactions[1].RawNarration != "Example payment continued narration" {
		t.Fatalf("symmetric narration was lost: %+v, %v", transactions, err)
	}
	for _, competing := range []extractor.PositionalRow{
		{Page: 1, Y: 501.3, Elements: []extractor.PositionalElement{{X: 35.5, EndX: 91.9, S: "9 04-04-2026"}, {X: 370, S: "50.00"}, {X: 517, S: "850.00 CR"}}},
		{Page: 1, Y: 501.3, Elements: []extractor.PositionalElement{{X: 370, S: "999.00"}}},
	} {
		conflict := slices.Clone(rows)
		conflict[4] = competing
		if _, _, err := parseBankHistoryRows(conflict, unionSavingsProfile); err == nil {
			t.Fatal("competing date or financial anchor was accepted")
		}
	}
}

func TestBankNarrationMeasuredOverlapAtCenteredHeadingBoundary(t *testing.T) {
	for _, adjacent := range []bool{false, true} {
		rows := syntheticFragmentedUnionRows()[:2]
		rows[1].Elements = slices.Clone(rows[1].Elements)
		rows[1].Elements[1].X = 63.6
		rows[1].Elements[2].X = 150.1
		for index, entry := range []struct {
			date, narration, amount, balance string
			x                                float64
		}{
			{"02-04-2026", "Example debit with long description past blank reference column", "100.00", "900.00 CR", 370},
			{"03-04-2026", "Example credit with long description past blank reference column", "250.00", "1,150.00 CR", 450},
		} {
			y := 512 - float64(index)*22
			elements := []extractor.PositionalElement{{X: 63.6, S: entry.date}, {X: entry.x, S: entry.amount}, {X: 517, S: entry.balance}}
			if adjacent {
				rows = append(rows, extractor.PositionalRow{Page: 1, Y: y, Elements: []extractor.PositionalElement{{X: 105, S: entry.narration}}})
				y -= 5.5
			} else {
				elements = append(elements, extractor.PositionalElement{X: 105, S: entry.narration})
			}
			rows = append(rows, extractor.PositionalRow{Page: 1, Y: y, Elements: elements})
		}
		data := syntheticBaselinePDFWithFontSize(t, rows, 4)
		transactions, _, err := (&UnionSavingsPDFParser{}).Parse(bytes.NewReader(data), ParseOptions{})
		if err != nil || len(transactions) != 2 || transactions[0].RawNarration != "Example debit with long description past blank reference column" || transactions[1].RawNarration != "Example credit with long description past blank reference column" {
			extracted, _ := extractor.ExtractPDFPositionalRows(bytes.NewReader(data), "")
			t.Fatalf("adjacent=%v: narration at x105 was lost: %+v, %v; synthetic rows=%+v", adjacent, transactions, err, extracted)
		}
	}
	for _, element := range []extractor.PositionalElement{
		{X: 105, S: "Example", EndX: 106},
		{X: 105, S: "Example"},
		{X: 100, S: "Example", EndX: 140},
		{X: 105, S: "02-04-", EndX: 130},
		{X: 105, S: "2026", EndX: 130},
		{X: 105, S: "100.00", EndX: 140},
		{X: 105, S: "100.00 CR", EndX: 140},
	} {
		if bankHistoryDescriptionElement(element, 106.85, 206.55) {
			t.Fatalf("ambiguous overlap accepted: %+v", element)
		}
	}
	// Extent beyond the inferred end is allowed identically for normal and
	// small-overlap description starts; blank reference columns may be crossed.
	for _, x := range []float64{105, 150.1} {
		if !bankHistoryDescriptionElement(extractor.PositionalElement{X: x, S: "Long description", EndX: 280}, 106.85, 206.55) {
			t.Fatalf("long description starting at %v was rejected", x)
		}
	}
}

func syntheticFragmentedUnionRows() []extractor.PositionalRow {
	return []extractor.PositionalRow{
		line(1, 580, "Union Bank of India UBIN DETAILS OF STATEMENT"),
		{Page: 1, Y: 550, Elements: []extractor.PositionalElement{{X: 29, S: "SI"}, {X: 64, S: "Date"}, {X: 150, S: "Particulars"}, {X: 263, S: "Chq Num"}, {X: 338, S: "Withdrawal"}, {X: 426, S: "Deposit"}, {X: 513, S: "Balance"}}},
		{Page: 1, Y: 512, Elements: []extractor.PositionalElement{{X: 105, S: "02-04-"}, {X: 150, S: "Example debit"}}},
		{Page: 1, Y: 507, Elements: []extractor.PositionalElement{{X: 35, S: "1"}, {X: 370, S: "100.00"}, {X: 517, S: "900.00 CR"}}},
		{Page: 1, Y: 501, Elements: []extractor.PositionalElement{{X: 105, S: "2026"}}},
		{Page: 1, Y: 490, Elements: []extractor.PositionalElement{{X: 105, S: "03-04-"}, {X: 150, S: "Example credit"}}},
		{Page: 1, Y: 485, Elements: []extractor.PositionalElement{{X: 35, S: "2"}, {X: 450, S: "250.00"}, {X: 517, S: "1,150.00 CR"}}},
		{Page: 1, Y: 479, Elements: []extractor.PositionalElement{{X: 105, S: "2026"}}},
	}
}

func syntheticBaselinePDF(t *testing.T, rows []extractor.PositionalRow) []byte {
	t.Helper()
	return syntheticBaselinePDFWithFontSize(t, rows, 8)
}

func syntheticBaselinePDFWithFontSize(t *testing.T, rows []extractor.PositionalRow, fontSize float64, pageSetup ...func(*gofpdf.Fpdf)) []byte {
	t.Helper()
	pdf := gofpdf.NewCustom(&gofpdf.InitType{UnitStr: "pt", Size: gofpdf.SizeType{Wd: 800, Ht: 650}})
	// Explicit widths exercise measured text extents. The PDF reader does not
	// supply widths for gofpdf's implicit core-font dictionaries.
	widths := make([]int, 256)
	for index := range widths {
		widths[index] = 600
	}
	font, err := json.Marshal(map[string]any{
		"Tp": "Type1", "Name": "Courier", "Cw": widths, "Enc": "cp1252", "Up": -100, "Ut": 50,
		"Desc": map[string]any{"Ascent": 629, "Descent": -157, "CapHeight": 562, "Flags": 33, "StemV": 51, "MissingWidth": 600, "FontBBox": map[string]int{"Xmin": -23, "Ymin": -250, "Xmax": 715, "Ymax": 805}},
	})
	if err != nil {
		t.Fatal(err)
	}
	pdf.AddFontFromBytes("fixture", "", font, nil)
	page := 0
	for _, row := range rows {
		if row.Page != page {
			pdf.AddPage()
			pdf.SetFont("fixture", "", fontSize)
			for _, setup := range pageSetup {
				setup(pdf)
			}
			page = row.Page
		}
		for _, element := range row.Elements {
			pdf.Text(element.X, 650-row.Y, element.S)
		}
	}
	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}
