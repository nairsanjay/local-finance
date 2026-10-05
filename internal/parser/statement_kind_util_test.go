package parser

import (
	"bytes"
	"strings"
	"testing"

	"local-finance/internal/parser/extractor"
)

func TestBankParsersRejectInvestmentDocumentTitles(t *testing.T) {
	for _, title := range []string{"Consolidated Account Statement", "CAS\nNSE\nTRANSACTION DETAILS"} {
		rows := []extractor.PositionalRow{line(1, 10, "HDFC Bank")}
		for index, titleLine := range strings.Split(title, "\n") {
			rows = append(rows, line(1, float64(20+index*10), titleLine))
		}
		data := syntheticLayoutPDF(t, rows)
		registry := NewRegistry()
		registry.Register(&HDFCSavingsPDFParser{})
		if parser, confidence, _ := registry.Detect("HDFC_ACCOUNT_STATEMENT.pdf", data); parser != nil || confidence != 0 {
			t.Fatalf("investment document detected as a bank statement: %v, %v", parser, confidence)
		}
		if _, _, err := (&HDFCSavingsPDFParser{}).Parse(bytes.NewReader(data), ParseOptions{}); err == nil {
			t.Fatal("manual HDFC adapter accepted an investment statement")
		}
		unionRows := syntheticFragmentedUnionRows()
		unionRows = append(rows, unionRows...)
		if _, _, err := parseBankHistoryRows(unionRows, unionSavingsProfile); err == nil {
			t.Fatal("manual Union adapter accepted an investment statement")
		}
	}
}

func TestInvestmentMerchantsDoNotExcludeBankStatements(t *testing.T) {
	for _, text := range []string{
		"HDFC Bank Account Statement\n02/04/2026 NSE transfer 100.00",
		"ICICI Bank Account Statement\nUPI CDSL INDMONEY ZERODHA CAS investment 100.00",
		"02/04/2026 Payment for Consolidated Account Statement 100.00",
		"HDFC Bank Account Statement\nCAS\nNSE",
	} {
		if isInvestmentStatement(text) {
			t.Fatalf("ordinary narration excluded the bank document: %q", text)
		}
	}
	rows := syntheticFragmentedUnionRows()
	rows[2].Elements[1].S = "NSE investment transfer"
	if transactions, _, err := parseBankHistoryRows(rows, unionSavingsProfile); err != nil || len(transactions) != 2 {
		t.Fatalf("bank payment to investment merchant not parsed: %v", err)
	}
}
