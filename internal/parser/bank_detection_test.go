package parser

import (
	"bytes"
	"reflect"
	"testing"
)

func TestBankTableDetectionAllowsCreditCardPaymentNarration(t *testing.T) {
	for _, fixture := range []struct {
		adapter StatementParser
		text    string
	}{
		{&ICICISavingsPDFParser{}, "%PDF ICICI Bank Statement of Transactions in Saving Account DATE TRANSACTION REMARKS WITHDRAWAL DEPOSIT BALANCE\n02.04.2026 CREDIT CARD PAYMENT"},
		{&UnionSavingsPDFParser{}, "%PDF Union Bank of India UBIN DETAILS OF STATEMENT DATE PARTICULARS WITHDRAWAL DEPOSIT BALANCE\n02.04.2026 CREDIT CARD PAYMENT"},
		{&HDFCSavingsPDFParser{}, "%PDF HDFC Bank Account Statement DATE NARRATION WITHDRAWAL DEPOSIT CLOSING BALANCE\n02.04.2026 CREDIT CARD PAYMENT"},
	} {
		confidence, _, _ := fixture.adapter.CanParse("statement.pdf", []byte(fixture.text))
		if confidence < 0.9 {
			t.Fatalf("%s rejected a bank table containing card-payment narration: %v", fixture.adapter.ID(), confidence)
		}
		selected, _, _ := DefaultRegistry.Detect("statement.pdf", []byte(fixture.text))
		if selected == nil || selected.ID() != fixture.adapter.ID() {
			t.Fatalf("bank table detected as another format: want %s, got %v", fixture.adapter.ID(), selected)
		}
		for _, input := range []struct{ filename, text string }{
			{"statement_CC.pdf", fixture.text},
			{"statement.pdf", fixture.text + "\nTOTAL AMOUNT DUE 100.00"},
			{"statement.pdf", "%PDF CREDIT CARD STATEMENT TOTAL AMOUNT DUE 100.00"},
		} {
			if confidence, _, _ := fixture.adapter.CanParse(input.filename, []byte(input.text)); confidence != 0 {
				t.Fatalf("%s accepted explicit card-format evidence", fixture.adapter.ID())
			}
		}
	}
}

func TestICICIBankCardPaymentAutoAndManualMatch(t *testing.T) {
	rows := syntheticICICIHistoryHeaderRows()
	rows[4].Elements[2].S = "CREDIT CARD PAYMENT"
	data := syntheticBaselinePDFWithFontSize(t, rows, 4)
	automatic, _, _ := DefaultRegistry.Detect("statement.pdf", data)
	if automatic == nil || automatic.ID() != "icici_savings_pdf_v1" {
		t.Fatalf("ICICI savings PDF with card-payment narration detected incorrectly: %v", automatic)
	}
	opts := ParseOptions{Filename: "statement.pdf"}
	autoRows, autoMeta, autoErr := automatic.Parse(bytes.NewReader(data), opts)
	manualRows, manualMeta, manualErr := (&ICICISavingsPDFParser{}).Parse(bytes.NewReader(data), opts)
	if autoErr != nil || manualErr != nil || len(autoRows) != 2 || !reflect.DeepEqual(autoRows, manualRows) || !reflect.DeepEqual(autoMeta, manualMeta) {
		t.Fatalf("automatic/manual bank parsing differs: auto=%v manual=%v", autoErr, manualErr)
	}
}
