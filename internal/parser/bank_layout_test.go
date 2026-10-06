package parser

import (
	"bytes"
	"testing"

	"github.com/jung-kurt/gofpdf"
	"local-finance/internal/models"
	"local-finance/internal/parser/extractor"
)

func TestICICIDepositsBeforeWithdrawalsAndModeColumn(t *testing.T) {
	rows := []extractor.PositionalRow{
		line(1, 10, "ICICI Bank Account Statement"),
		{Page: 1, Elements: []extractor.PositionalElement{{X: 20, S: "DATE"}, {X: 90, S: "MODE"}, {X: 170, S: "PARTICULARS"}, {X: 380, S: "DEPOSITS"}, {X: 480, S: "WITHDRAWALS"}, {X: 580, S: "BALANCE"}}},
		{Page: 1, Elements: []extractor.PositionalElement{{X: 20, S: "2-Apr-2026"}, {X: 90, S: "UPI"}, {X: 170, S: "Example withdrawal"}, {X: 480, S: "100.00"}, {X: 580, S: "900.00"}}},
		{Page: 2, Elements: []extractor.PositionalElement{{X: 20, S: "DATE"}, {X: 90, S: "MODE"}, {X: 170, S: "PARTICULARS"}, {X: 380, S: "DEPOSITS"}, {X: 480, S: "WITHDRAWALS"}, {X: 580, S: "BALANCE"}}},
		{Page: 2, Elements: []extractor.PositionalElement{{X: 20, S: "3-Apr-2026"}, {X: 90, S: "NEFT"}, {X: 170, S: "Example deposit"}, {X: 380, S: "250.00"}, {X: 580, S: "1,150.00"}}},
	}
	transactions, meta, err := (&ICICISavingsPDFParser{}).Parse(bytes.NewReader(syntheticLayoutPDF(t, rows)), ParseOptions{Filename: "statement.pdf"})
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 2 || transactions[0].Date != "2026-04-02" || transactions[0].TxType != models.TxTypeDebit || transactions[0].Amount != 100 || transactions[1].TxType != models.TxTypeCredit || transactions[1].Amount != 250 {
		t.Fatalf("reversed financial columns handled incorrectly: %+v", transactions)
	}
	if meta.OpeningBalance != 1000 || meta.ClosingBalance != 1150 {
		t.Fatalf("balance reconciliation incorrect: %+v", meta)
	}
}

func TestUnionSerialColumnAndCenteredHeadings(t *testing.T) {
	rows := []extractor.PositionalRow{
		line(1, 10, "Union Bank of India UBIN DETAILS OF STATEMENT"),
		{Page: 1, Elements: []extractor.PositionalElement{{X: 10, S: "SI"}, {X: 70, S: "Date"}, {X: 170, S: "Particulars"}, {X: 300, S: "Chq Num"}, {X: 400, S: "Withdrawal"}, {X: 500, S: "Deposit"}, {X: 600, S: "Balance"}}},
		{Page: 1, Elements: []extractor.PositionalElement{{X: 10, S: "1"}, {X: 45, S: "2/04/2026"}, {X: 150, S: "Example withdrawal"}, {X: 375, S: "100.00"}, {X: 575, S: "900.00"}}},
		{Page: 2, Elements: []extractor.PositionalElement{{X: 10, S: "SI"}, {X: 70, S: "Date"}, {X: 170, S: "Particulars"}, {X: 300, S: "Chq Num"}, {X: 400, S: "Withdrawal"}, {X: 500, S: "Deposit"}, {X: 600, S: "Balance"}}},
		{Page: 2, Elements: []extractor.PositionalElement{{X: 10, S: "2"}, {X: 45, S: "3/04/2026"}, {X: 150, S: "Example deposit"}, {X: 475, S: "250.00"}, {X: 575, S: "1,150.00"}}},
	}
	transactions, _, err := (&UnionSavingsPDFParser{}).Parse(bytes.NewReader(syntheticLayoutPDF(t, rows)), ParseOptions{Filename: "statement.pdf"})
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 2 || transactions[0].Date != "2026-04-02" || transactions[0].Amount != 100 || transactions[1].Amount != 250 {
		t.Fatalf("serial/date/alignment handled incorrectly: %+v", transactions)
	}
}

func TestICICIAccountSummaryIsNotAnImportableStatement(t *testing.T) {
	data := syntheticLayoutPDF(t, []extractor.PositionalRow{line(1, 10, "ICICI Bank Account Summary"), line(1, 20, "Available Balance 1,000.00")})
	if _, _, err := (&ICICISavingsPDFParser{}).Parse(bytes.NewReader(data), ParseOptions{}); err == nil {
		t.Fatal("an account summary without transactions was accepted")
	}
}

func syntheticLayoutPDF(t *testing.T, rows []extractor.PositionalRow) []byte {
	t.Helper()
	pdf := gofpdf.NewCustom(&gofpdf.InitType{UnitStr: "pt", Size: gofpdf.SizeType{Wd: 800, Ht: 600}})
	page, y := 0, 0.0
	for _, row := range rows {
		if row.Page != page {
			pdf.AddPage()
			pdf.SetFont("Helvetica", "", 8)
			page, y = row.Page, 40
		}
		for _, element := range row.Elements {
			pdf.Text(element.X+15, y, element.S)
		}
		y += 25
	}
	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}
