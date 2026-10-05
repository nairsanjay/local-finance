// Command generate_bank_history_samples creates privacy-safe PDF fixtures for
// the ICICI and Union Bank savings parsers. All names, references, and amounts
// in these documents are fictional.
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/jung-kurt/gofpdf"
)

func main() {
	dir := filepath.Join("samples", "savings")
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal(err)
	}
	for _, sample := range []struct {
		name   string
		bank   string
		title  string
		period string
		rows   [][]string
	}{
		{
			name: "ICICI_Savings_Synthetic.pdf", bank: "ICICI BANK", title: "Statement of Transactions in Saving Account",
			period: "Statement Period: 10/04/2026 to 11/04/2026",
			rows: [][]string{
				{"Transaction Date", "Transaction Remarks", "Withdrawal", "Deposit", "Balance"},
				{"10.04.2026", "UPI/DEMO CAFE/demo-cafe@upi", "100.00", "", "900.00 Cr"},
				{"11.04.2026", "NEFT CR/EXAMPLE EMPLOYER", "", "250.00", "1,150.00 Cr"},
			},
		},
		{
			name: "Union_Bank_Savings_Synthetic.pdf", bank: "UNION BANK OF INDIA (UBIN)", title: "DETAILS OF STATEMENT",
			period: "Period From 10-04-2026 to 11-04-2026",
			rows: [][]string{
				{"Date", "Particulars", "Chq Num", "Withdrawal", "Deposit", "Balance"},
				{"10-04-2026", "UPI/DR/123456789012/DEMO STORE", "", "100.00", "", "900.00 Cr"},
				{"11-04-2026", "NEFT CR/EXAMPLE PAYER", "", "", "250.00", "1,150.00 Cr"},
			},
		},
	} {
		if err := writeSample(filepath.Join(dir, sample.name), sample.bank, sample.title, sample.period, sample.rows); err != nil {
			log.Fatal(err)
		}
		log.Printf("Generated %s", filepath.Join(dir, sample.name))
	}
}

func writeSample(path, bank, title, period string, rows [][]string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 12, 12)
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(190, 8, bank, "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(190, 6, title, "", 1, "L", false, 0, "")
	pdf.CellFormat(190, 6, period, "", 1, "L", false, 0, "")
	pdf.CellFormat(190, 6, "Opening Balance: 1,000.00 Cr", "", 1, "L", false, 0, "")
	pdf.Ln(3)

	// Fixed positions resemble the separated transaction columns used in bank
	// downloads, while leaving enough horizontal space for the PDF text extractor.
	columns := []float64{12, 42, 103, 145, 174}
	if len(rows[0]) == 6 {
		columns = []float64{12, 36, 98, 130, 158, 181}
	}
	for rowIndex, row := range rows {
		if rowIndex == 0 {
			pdf.SetFont("Helvetica", "B", 8)
		} else {
			pdf.SetFont("Helvetica", "", 8)
		}
		y := pdf.GetY()
		for index, value := range row {
			if value != "" {
				pdf.Text(columns[index], y+4, value)
			}
		}
		pdf.SetY(y + 8)
	}
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "", 8)
	pdf.CellFormat(190, 5, "Total Debits: 100.00", "", 1, "L", false, 0, "")
	pdf.CellFormat(190, 5, "Total Credits: 250.00", "", 1, "L", false, 0, "")
	pdf.CellFormat(190, 5, "Closing Balance: 1,150.00 Cr", "", 1, "L", false, 0, "")
	return pdf.OutputFileAndClose(path)
}
