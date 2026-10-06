package integration_test

import (
	"bytes"
	"testing"

	"local-finance/internal/models"
	"local-finance/internal/parser"
	"local-finance/internal/service"
)

func TestSyntheticBankHistoryPDFSamplesEndToEnd(t *testing.T) {
	tests := []struct {
		file, parserID, bank string
		firstDate            string
	}{
		{"ICICI_Savings_Synthetic.pdf", "icici_savings_pdf_v1", "ICICI Bank", "2026-04-10"},
		{"Union_Bank_Savings_Synthetic.pdf", "union_savings_pdf_v1", "Union Bank of India", "2026-04-10"},
	}

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			data := syntheticSamplePDF(t, tc.file)

			// These fixtures use compressed content streams. A generic filename
			// must be detected from extracted PDF text rather than the bank name.
			selected, confidence, detected := parser.DefaultRegistry.Detect("statement.pdf", data)
			if selected == nil {
				t.Fatal("registry did not detect the synthetic statement")
			}
			if selected.ID() != tc.parserID || confidence < 0.9 || detected.BankName != tc.bank || detected.AccountType != models.AccountTypeSavings {
				t.Fatalf("unexpected detection: parser=%v confidence=%.2f meta=%+v", selected.ID(), confidence, detected)
			}

			transactions, meta, err := selected.Parse(bytes.NewReader(data), parser.ParseOptions{Filename: tc.file})
			if err != nil {
				t.Fatalf("parse synthetic statement: %v", err)
			}
			if len(transactions) != 2 {
				t.Fatalf("got %d transactions; want 2", len(transactions))
			}
			if transactions[0].Date != tc.firstDate || transactions[0].TxType != models.TxTypeDebit || transactions[0].Amount != 100 || transactions[0].RunningBalance == nil || *transactions[0].RunningBalance != 900 {
				t.Errorf("unexpected first debit row: %+v", transactions[0])
			}
			if transactions[1].TxType != models.TxTypeCredit || transactions[1].Amount != 250 || transactions[1].RunningBalance == nil || *transactions[1].RunningBalance != 1150 {
				t.Errorf("unexpected second credit row: %+v", transactions[1])
			}
			if meta.BankName != tc.bank || meta.OpeningBalance != 1000 || meta.ClosingBalance != 1150 || meta.TotalDebits != 100 || meta.TotalCredits != 250 {
				t.Errorf("unexpected statement controls: %+v", meta)
			}

			svc := service.NewTransactionService(testDatabase(t))
			preview, err := svc.PreviewStatement("statement.pdf", bytes.NewReader(data), "", "", "")
			if err != nil {
				t.Fatal(err)
			}
			if preview.Error != "" || preview.TotalTransactions != 2 {
				t.Fatalf("generic filename preview failed: %+v", preview)
			}
			result, err := svc.ImportStatement("statement.pdf", bytes.NewReader(data), "", "", "")
			if err != nil {
				t.Fatal(err)
			}
			if result.TotalParsed != 2 || result.InsertedCount != 2 {
				t.Fatalf("generic filename import failed: %+v", result)
			}
		})
	}
}
