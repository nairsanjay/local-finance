package integration_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jung-kurt/gofpdf"
	"local-finance/internal/db"
)

func testDatabase(t *testing.T) *db.DB {
	t.Helper()
	database, err := db.NewDB(filepath.Join(t.TempDir(), "ledger.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

func syntheticSamplePDF(t *testing.T, filename string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "samples", "savings", filename))
	if err != nil {
		t.Fatalf("read synthetic PDF fixture: %v", err)
	}
	return data
}

// Construct only fictional account headers and transactions in memory; no
// user statements or generated files are needed for these regressions.
func syntheticBankHistoryPDF(t *testing.T, bank, account string) []byte {
	t.Helper()
	pdf := gofpdf.NewCustom(&gofpdf.InitType{UnitStr: "pt", Size: gofpdf.SizeType{Wd: 800, Ht: 600}})
	pdf.SetSubject(strings.Repeat("Synthetic test metadata. ", 4000), false)
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 9)
	title := "Statement of Transactions in Saving Account"
	identity := bank
	if bank == "Union Bank of India" {
		title = "DETAILS OF STATEMENT"
		identity += " UBIN"
	}
	pdf.Text(40, 40, identity)
	pdf.Text(40, 60, title)
	pdf.Text(40, 80, "Account Number: "+account)
	columns := []float64{40, 180, 430, 550, 680}
	for i, heading := range []string{"Date", "Transaction Remarks", "Withdrawal", "Deposit", "Balance"} {
		pdf.Text(columns[i], 110, heading)
	}
	for index, row := range [][]string{
		{"10/04/2026", "Example withdrawal", "100.00", "", "900.00 Cr"},
		{"11/04/2026", "Example deposit", "", "250.00", "1,150.00 Cr"},
	} {
		for i, value := range row {
			if value != "" {
				pdf.Text(columns[i], float64(140+index*25), value)
			}
		}
	}
	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}
