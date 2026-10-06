package parser

import (
	"slices"
	"strings"
	"testing"

	"local-finance/internal/models"
	"local-finance/internal/parser/extractor"
)

func TestICICISavingsHistoryReconcilesRowsAndDerivesOpeningBalance(t *testing.T) {
	rows := syntheticICICIHistory()
	transactions, meta, err := parseBankHistoryRows(rows, iciciSavingsProfile)
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
	if _, _, err := parseBankHistoryRows(rows, iciciSavingsProfile); err == nil {
		t.Fatal("inconsistent running balances were accepted")
	}

	rows = syntheticICICIHistory()
	rows[3].Elements[0].S = "31/04/2026"
	if _, _, err := parseBankHistoryRows(rows, iciciSavingsProfile); err == nil {
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
	transactions, meta, err := parseBankHistoryRows(rows, unionSavingsProfile)
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

func TestBankHistoryAccountIdentifiers(t *testing.T) {
	for _, profile := range []bankHistoryProfile{iciciSavingsProfile, unionSavingsProfile} {
		t.Run(profile.bankName, func(t *testing.T) {
			for _, account := range []string{"123456781234", "123456785678", "1234XXXX5678"} {
				rows := syntheticICICIHistory()
				rows[0] = line(1, 10, profile.identityTerm)
				rows[1] = line(1, 20, profile.statementTerm)
				rows = slices.Insert(rows, 2, line(1, 25, "Account Number: "+account))
				_, meta, err := parseBankHistoryRows(rows, profile)
				if err != nil {
					t.Fatalf("parse account %s: %v", account, err)
				}
				if meta.AccountNumberMask != "XX"+account[len(account)-4:] {
					t.Fatalf("incorrect account mask: %q", meta.AccountNumberMask)
				}
				if strings.Contains(account, "X") {
					if meta.AccountNumber != "" {
						t.Fatal("masked account was treated as a complete account number")
					}
				} else if meta.AccountNumber != account {
					t.Fatalf("full account number not retained: %q", meta.AccountNumber)
				}
			}
		})
	}
}

func TestBankHistoryWrappedNarrationAndReference(t *testing.T) {
	rows := syntheticICICIHistory()
	rows[2].Elements = slices.Insert(rows[2].Elements, 2, extractor.PositionalElement{X: 230, S: "Reference"})
	rows[3].Elements[1].S = "UPI/DR/123456789012/EXAMPLE"
	rows[3].Elements = slices.Insert(rows[3].Elements, 2, extractor.PositionalElement{X: 230, S: "DEMO"})
	continuation := extractor.PositionalRow{Page: 1, Y: 65, Elements: []extractor.PositionalElement{
		{X: 150, S: "CAFE/DEMO BANK/example@upi/Payment"}, {X: 230, S: "123456"},
	}}
	rows = slices.Insert(rows, 4, continuation, line(1, 64, "Page 1 of 1"))
	transactions, _, err := parseBankHistoryRows(rows, iciciSavingsProfile)
	if err != nil {
		t.Fatal(err)
	}
	if transactions[0].RawNarration != "UPI/DR/123456789012/EXAMPLE CAFE/DEMO BANK/example@upi/Payment" || transactions[0].CleanedPayee != "Example Cafe" {
		t.Fatalf("wrapped narration was not assembled before cleaning: %+v", transactions[0])
	}
	if transactions[0].ReferenceNumber != "DEMO123456" || transactions[0].UPIVPA == nil || *transactions[0].UPIVPA != "example@upi" {
		t.Fatalf("wrapped reference or UPI metadata lost: %+v", transactions[0])
	}
}

func TestBankHistoryAcceptsZeroRunningBalance(t *testing.T) {
	rows := syntheticICICIHistory()
	rows[3] = bankHistoryDataRow("10.04.2026", "Example withdrawal", "1,000.00", "0.00", "0.00 Cr")
	rows[4] = bankHistoryDataRow("11.04.2026", "Example deposit", "0.00", "250.00", "250.00 Cr")
	transactions, meta, err := parseBankHistoryRows(rows, iciciSavingsProfile)
	if err != nil {
		t.Fatal(err)
	}
	if *transactions[0].RunningBalance != 0 || meta.OpeningBalance != 1000 || meta.ClosingBalance != 250 {
		t.Fatalf("zero running balance handled incorrectly: transactions=%+v meta=%+v", transactions, meta)
	}
}

func TestBankHistoryDescendingIncludingSameDay(t *testing.T) {
	rows := syntheticICICIHistory()
	rows = append(rows[:3],
		bankHistoryDataRow("11.04.2026", "Later withdrawal", "50.00", "", "1,100.00 Cr"),
		bankHistoryDataRow("10.04.2026", "Example deposit", "", "250.00", "1,150.00 Cr"),
		bankHistoryDataRow("10.04.2026", "Earlier withdrawal", "100.00", "", "900.00 Cr"),
	)
	transactions, meta, err := parseBankHistoryRows(rows, iciciSavingsProfile)
	if err != nil {
		t.Fatal(err)
	}
	if transactions[0].RawNarration != "Earlier withdrawal" || transactions[1].RawNarration != "Example deposit" || meta.OpeningBalance != 1000 || meta.ClosingBalance != 1100 {
		t.Fatalf("descending transaction order incorrect: transactions=%+v meta=%+v", transactions, meta)
	}
	// An all-one-day export provides no date ordering signal. Its balances must
	// still determine which complete source sequence is chronological.
	rows[3].Elements[0].S = "10.04.2026"
	if _, _, err := parseBankHistoryRows(rows, iciciSavingsProfile); err != nil {
		t.Fatalf("descending same-day-only statement rejected: %v", err)
	}
}

func TestBankHistoryPerPageColumns(t *testing.T) {
	rows := syntheticICICIHistory()
	secondHeader := extractor.PositionalRow{Page: 2, Y: 100, Elements: slices.Clone(rows[2].Elements)}
	secondTransaction := extractor.PositionalRow{Page: 2, Y: 90, Elements: slices.Clone(rows[4].Elements)}
	for i := range secondHeader.Elements {
		secondHeader.Elements[i].X += 30
	}
	for i := range secondTransaction.Elements {
		secondTransaction.Elements[i].X += 30
	}
	rows = append(rows[:4], line(2, 110, "ICICI Bank"), secondHeader, secondTransaction)
	transactions, _, err := parseBankHistoryRows(rows, iciciSavingsProfile)
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 2 || transactions[0].Amount != 100 || transactions[1].Amount != 250 {
		t.Fatalf("page-specific columns handled incorrectly: %+v", transactions)
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
