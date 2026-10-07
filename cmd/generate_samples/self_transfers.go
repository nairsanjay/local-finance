package main

import (
	"encoding/csv"
	"os"
)

func generateHDFCSelfTransferCSV(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	writer := csv.NewWriter(file)
	err = writer.WriteAll([][]string{
		{"Date", "Narration", "Chq/Ref", "Value Dt", "Withdrawal", "Deposit", "Closing Balance"},
		{"01/04/2026", "NEFT CR-ABC123-EMPLOYER-SALARY", "ABC123", "01/04/2026", "", "1000", "1000"},
		{"02/04/2026", "UPI-ALEX-alex@okhdfcbank-HDFC-123456789012-SELF TRANSFER", "DEMO123", "02/04/2026", "500", "", "500"},
		{"03/04/2026", "UPI/CR/123456789013/ALEX/HDFC/alex@okhdfcbank/SELF TRANSFER", "DEMO124", "03/04/2026", "", "500", "1000"},
		{"04/04/2026", "UPI-SHOP-shop@okhdfcbank-HDFC-123456789014-PAYMENT", "123456789014", "04/04/2026", "100", "", "900"},
	})
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}
