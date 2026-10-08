package parser

import (
	"bytes"
	"slices"
	"testing"

	"github.com/jung-kurt/gofpdf"
	"local-finance/internal/models"
	"local-finance/internal/parser/extractor"
)

func syntheticICICIHistoryHeaderRows() []extractor.PositionalRow {
	return []extractor.PositionalRow{
		line(1, 640, "ICICI Bank Statement of transactions in saving account"),
		{Page: 1, Y: 625, Elements: []extractor.PositionalElement{{X: 60.2, S: "Transaction"}, {X: 399, S: "Withdrawal"}, {X: 473.8, S: "Deposit"}, {X: 532.1, S: "Balance"}}},
		{Page: 1, Y: 620, Elements: []extractor.PositionalElement{{X: 23.5, S: "S"}, {X: 31.2, S: "No."}, {X: 122.4, S: "Cheque"}, {X: 154.1, S: "Number"}, {X: 247.1, S: "Transaction"}, {X: 297, S: "Remarks"}}},
		{Page: 1, Y: 615, Elements: []extractor.PositionalElement{{X: 74.4, S: "Date"}, {X: 395.6, S: "Amount (INR)"}, {X: 461.6, S: "Amount (INR)"}, {X: 537.6, S: "(INR)"}}},
		{Page: 1, Y: 595, Elements: []extractor.PositionalElement{{X: 30.1, S: "1"}, {X: 61.4, S: "02.04.2026"}, {X: 200, S: "Example debit"}, {X: 399, S: "100.00"}, {X: 532.1, S: "900.00"}}},
		{Page: 1, Y: 545, Elements: []extractor.PositionalElement{{X: 30.1, S: "2"}, {X: 61.4, S: "03.04.2026"}, {X: 200, S: "Example credit"}, {X: 473.8, S: "250.00"}, {X: 532.1, S: "1,150.00"}}},
	}
}

func TestICICITransactionHistoryThreeBaselineHeader(t *testing.T) {
	rows := syntheticICICIHistoryHeaderRows()
	data := syntheticBaselinePDFWithFontSize(t, rows, 4)
	transactions, meta, err := (&ICICISavingsPDFParser{}).Parse(bytes.NewReader(data), ParseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 2 || transactions[0].Date != "2026-04-02" || transactions[0].Amount != 100 || transactions[0].RawNarration != "Example debit" || transactions[1].Amount != 250 || meta.OpeningBalance != 1000 || meta.ClosingBalance != 1150 {
		t.Fatalf("split headings handled incorrectly: transactions=%+v meta=%+v", transactions, meta)
	}
	merged := coalesceICICIHistoryHeaders(rows)
	columns, complete := discoverBankHistoryColumns(merged[1], iciciSavingsProfile)
	if !complete || columns.date != 60.2 || columns.description != 247.1 || columns.reference != 122.4 || columns.serial != 23.5 {
		t.Fatalf("heading words assigned to wrong columns: %+v", columns)
	}
}

func TestICICIHeaderMergeRejectsDataAmbiguityAndCrossPage(t *testing.T) {
	for _, change := range []func([]extractor.PositionalRow){
		func(rows []extractor.PositionalRow) { rows[2].Elements[3].S = "100.00" },
		func(rows []extractor.PositionalRow) { rows[3].Page = 2 },
		func(rows []extractor.PositionalRow) { rows[3].Y = 614 },
		func(rows []extractor.PositionalRow) {
			rows[3].Elements = append(rows[3].Elements, extractor.PositionalElement{X: 75, S: "Date"})
		},
	} {
		rows := syntheticICICIHistoryHeaderRows()
		for index := range rows {
			rows[index].Elements = slices.Clone(rows[index].Elements)
		}
		change(rows)
		if _, _, err := parseBankHistoryRows(coalesceICICIHistoryHeaders(rows), iciciSavingsProfile); err == nil {
			t.Fatal("ambiguous/data-filled/distant/cross-page header accepted")
		}
	}
}

func TestICICIHeaderWithoutReportedFontWidths(t *testing.T) {
	rows := syntheticICICIHistoryHeaderRows()
	rows[2].Elements = []extractor.PositionalElement{{X: 23.5, S: "S NO."}, {X: 122.4, S: "CHEQUE NUMBER"}, {X: 247.1, S: "TRANSACTION REMARKS"}}
	for row := range rows {
		for element := range rows[row].Elements {
			rows[row].Elements[element].EndX = rows[row].Elements[element].X
		}
	}
	transactions, _, err := parseBankHistoryRows(coalesceICICIHistoryHeaders(rows), iciciSavingsProfile)
	if err != nil || len(transactions) != 2 {
		t.Fatalf("known heading coordinates require no fabricated font width: %+v, %v", transactions, err)
	}
}

func TestICICIHistoryRightAlignedSmallAmounts(t *testing.T) {
	rows := syntheticICICIHistoryHeaderRows()
	// Synthetic shaded financial cells center their labels while tiny amounts
	// sit at the right edge, beyond the old left-heading midpoint.
	rows[0].Elements[0].S += " for the period October 8, 2025 - October 7, 2026"
	rows[4].Elements[1].S = "11.10.2025"
	rows[4].Elements[3] = extractor.PositionalElement{X: 438, S: "1.00"}
	rows[4].Elements[4] = extractor.PositionalElement{X: 539, S: "999.00"}
	rows[5].Elements[1].S = "01.10.2026"
	rows[5].Elements[3] = extractor.PositionalElement{X: 502, S: "1.00"}
	rows[5].Elements[4] = extractor.PositionalElement{X: 539, S: "1000.00"}
	data := syntheticICICIHistoryCellPDF(t, rows)
	tx, meta, err := (&ICICISavingsPDFParser{}).Parse(bytes.NewReader(data), ParseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(tx) != 2 || tx[0].TxType != models.TxTypeDebit || tx[1].TxType != models.TxTypeCredit || meta.TotalDebits != 1 || meta.TotalCredits != 1 || meta.OpeningBalance != 1000 || meta.ClosingBalance != 1000 || meta.StartDate != "2025-10-08" || meta.EndDate != "2026-10-07" {
		t.Fatalf("small amounts or annual period misread: %+v %+v", tx, meta)
	}
	extracted, err := extractor.ExtractPDFPositionalRows(bytes.NewReader(data), "")
	if err != nil {
		t.Fatal(err)
	}
	if extracted[1].Elements[1].X != 399 {
		t.Fatal("shared PDF extractor changed the heading coordinates")
	}
	original := slices.Clone(extracted[1].Elements)
	if err := alignICICIHistoryFinancialHeaders(extracted, data); err != nil {
		t.Fatal(err)
	}
	if original[1].X != 399 || extracted[1].Elements[1].X != 423 {
		t.Fatal("normalization mutated source rows")
	}
	rows[5].Elements[4].S = "1001.00"
	if _, _, err := (&ICICISavingsPDFParser{}).Parse(bytes.NewReader(syntheticICICIHistoryCellPDF(t, rows)), ParseOptions{}); err == nil {
		t.Fatal("broken running balance accepted")
	}
}

func syntheticICICIHistoryCellPDF(t *testing.T, rows []extractor.PositionalRow) []byte {
	t.Helper()
	return syntheticBaselinePDFWithFontSize(t, rows, 4, func(document *gofpdf.Fpdf) {
		for _, center := range []float64{423, 489, 548} {
			document.Rect(center-30, 650-625-8, 60, 12, "D")
		}
	})
}
