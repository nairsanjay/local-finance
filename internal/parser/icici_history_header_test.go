package parser

import (
	"bytes"
	"slices"
	"testing"

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
