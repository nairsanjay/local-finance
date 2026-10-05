package parser

import (
	"testing"

	"local-finance/internal/models"
	"local-finance/internal/parser/extractor"
)

func TestICICISavingsHistoryReconcilesRowsAndDerivesOpeningBalance(t *testing.T) {
	rows := syntheticICICIHistory()
	transactions, meta, err := parseBankHistoryRows(rows, "sample.pdf", iciciSavingsProfile)
	if err != nil {
		t.Fatalf("parse ICICI account history: %v", err)
	}
	if len(transactions) != 2 {
		t.Fatalf("got %d rows; want 2", len(transactions))
	}
	if transactions[0].Date != "2026-04-10" || transactions[0].TxType != models.TxTypeDebit || transactions[0].Amount != 100 || *transactions[0].RunningBalance != 900 {
		t.Fatalf("unexpected first normalized row: %+v", transactions[0])
	}
	if transactions[1].TxType != models.TxTypeCredit || transactions[1].Amount != 250 || *transactions[1].RunningBalance != 1150 {
		t.Fatalf("unexpected second normalized row: %+v", transactions[1])
	}
	if meta.OpeningBalance != 1000 || meta.ClosingBalance != 1150 || meta.TotalDebits != 100 || meta.TotalCredits != 250 {
		t.Fatalf("unexpected reconciled balances/totals: %+v", meta)
	}
}

func TestICICISavingsHistoryRejectsBrokenBalanceAndMalformedRows(t *testing.T) {
	rows := syntheticICICIHistory()
	rows[4].Elements[len(rows[4].Elements)-1].S = "1,151.00 Cr"
	if _, _, err := parseBankHistoryRows(rows, "sample.pdf", iciciSavingsProfile); err == nil {
		t.Fatal("inconsistent running balances were accepted")
	}

	rows = syntheticICICIHistory()
	rows[3].Elements[0].S = "31/04/2026"
	if _, _, err := parseBankHistoryRows(rows, "sample.pdf", iciciSavingsProfile); err == nil {
		t.Fatal("invalid dated transaction was silently skipped")
	}
}

func TestUnionSavingsPDFUsesItsSpecificSchemaAndControls(t *testing.T) {
	rows := []extractor.PositionalRow{
		line(1, 10, "UBIN"),
		line(1, 20, "DETAILS OF STATEMENT"),
		line(1, 30, "PERIOD FROM 10-04-2026 TO 11-04-2026"),
		line(1, 40, "Opening Balance 1,000.00 Cr"),
		{Page: 1, Y: 50, Elements: []extractor.PositionalElement{{X: 0, S: "Date"}, {X: 150, S: "Particulars"}, {X: 230, S: "Chq Num"}, {X: 300, S: "Withdrawal"}, {X: 400, S: "Deposit"}, {X: 500, S: "Balance"}}},
		bankHistoryDataRow("10-04-2026", "UPI/DR/123456789012/FOOD/example@upi", "100.00", "", "900.00 Cr"),
		bankHistoryDataRow("11-04-2026", "NEFT CR/EXAMPLE", "", "250.00", "1,150.00 Cr"),
		line(1, 90, "Total Debits 100.00"),
		line(1, 100, "Total Credits 250.00"),
		line(1, 110, "Closing Balance 1,150.00 Cr"),
	}
	transactions, meta, err := parseBankHistoryRows(rows, "sample.pdf", unionSavingsProfile)
	if err != nil {
		t.Fatalf("parse Union account history: %v", err)
	}
	if len(transactions) != 2 || transactions[0].TxType != models.TxTypeDebit || transactions[1].TxType != models.TxTypeCredit {
		t.Fatalf("wrong Union row directions: %+v", transactions)
	}
	if meta.OpeningBalance != 1000 || meta.ClosingBalance != 1150 || meta.TotalDebits != 100 || meta.TotalCredits != 250 {
		t.Fatalf("printed Union controls did not reconcile: %+v", meta)
	}
}

func TestBankPDFAdaptersAreDetectedByInstitutionAndDocumentType(t *testing.T) {
	tests := []struct {
		parser StatementParser
		name   string
		body   string
		bank   string
	}{
		{&ICICISavingsPDFParser{}, "statement.pdf", "%PDF ICICI Bank Statement of Transactions in Saving Account", "ICICI Bank"},
		{&UnionSavingsPDFParser{}, "statement.pdf", "%PDF Union Bank of India UBIN DETAILS OF STATEMENT WITHDRAWAL DEPOSIT", "Union Bank of India"},
	}
	for _, test := range tests {
		confidence, bank, accountType := test.parser.CanParse(test.name, []byte(test.body))
		if confidence < 0.9 || bank != test.bank || accountType != models.AccountTypeSavings {
			t.Errorf("%s detection: confidence=%.2f bank=%q type=%q", test.parser.ID(), confidence, bank, accountType)
		}
	}
	if parser, _, _ := DefaultRegistry.Detect("statement.pdf", []byte("%PDF ICICI Bank Statement of Transactions in Saving Account")); parser == nil || parser.ID() != "icici_savings_pdf_v1" {
		t.Fatalf("registry did not select the ICICI savings parser: %v", parser)
	}
}

func syntheticICICIHistory() []extractor.PositionalRow {
	return []extractor.PositionalRow{
		line(1, 10, "ICICI Bank"),
		line(1, 20, "Statement of Transactions in Saving Account"),
		{Page: 1, Y: 30, Elements: []extractor.PositionalElement{{X: 0, S: "Transaction Date"}, {X: 150, S: "Transaction Remarks"}, {X: 300, S: "Withdrawal"}, {X: 400, S: "Deposit"}, {X: 500, S: "Balance"}}},
		bankHistoryDataRow("10.04.2026", "Example withdrawal", "100.00", "", "900.00 Cr"),
		bankHistoryDataRow("11.04.2026", "Example deposit", "", "250.00", "1,150.00 Cr"),
	}
}

func bankHistoryDataRow(date, narration, debit, credit, balance string) extractor.PositionalRow {
	elements := []extractor.PositionalElement{{X: 0, S: date}, {X: 150, S: narration}}
	if debit != "" {
		elements = append(elements, extractor.PositionalElement{X: 300, S: debit})
	}
	if credit != "" {
		elements = append(elements, extractor.PositionalElement{X: 400, S: credit})
	}
	elements = append(elements, extractor.PositionalElement{X: 500, S: balance})
	return extractor.PositionalRow{Page: 1, Y: 70, Elements: elements}
}

func line(page int, y float64, text string) extractor.PositionalRow {
	return extractor.PositionalRow{Page: page, Y: y, Elements: []extractor.PositionalElement{{X: 0, S: text}}}
}
