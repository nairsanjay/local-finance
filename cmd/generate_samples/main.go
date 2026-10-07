package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jung-kurt/gofpdf"
)

func main() {
	selfTransfersOnly := flag.Bool("self-transfers-only", false, "Generate only the synthetic self-transfer CSV fixture")
	flag.Parse()

	savingsDir := filepath.Join("samples", "savings")
	ccDir := filepath.Join("samples", "credit_cards")

	_ = os.MkdirAll(savingsDir, 0755)
	_ = os.MkdirAll(ccDir, 0755)

	if err := generateHDFCSelfTransferCSV(filepath.Join(savingsDir, "HDFC_Self_Transfer_Statement.csv")); err != nil {
		log.Fatal(err)
	}
	if *selfTransfersOnly {
		return
	}

	log.Println("🚀 Generating masked sample statements...")

	// 1. HDFC Savings PDF
	if err := generateHDFCSavingsPDF(filepath.Join(savingsDir, "HDFC_Savings_Account_Statement.pdf"), false); err != nil {
		log.Fatalf("Failed to generate HDFC Savings PDF: %v", err)
	}
	log.Println("✅ Generated samples/savings/HDFC_Savings_Account_Statement.pdf")

	// 2. HDFC Current PDF
	if err := generateHDFCSavingsPDF(filepath.Join(savingsDir, "HDFC_Current_Account_Statement.pdf"), true); err != nil {
		log.Fatalf("Failed to generate HDFC Current PDF: %v", err)
	}
	log.Println("✅ Generated samples/savings/HDFC_Current_Account_Statement.pdf")

	// 3. HDFC Savings CSV
	if err := generateHDFCSavingsCSV(filepath.Join(savingsDir, "HDFC_Savings_Account_Statement.csv")); err != nil {
		log.Fatalf("Failed to generate HDFC Savings CSV: %v", err)
	}
	log.Println("✅ Generated samples/savings/HDFC_Savings_Account_Statement.csv")

	// 4. HDFC Regalia CC PDF
	if err := generateHDFCRegaliaPDF(filepath.Join(ccDir, "HDFC_Regalia_Credit_Card.pdf")); err != nil {
		log.Fatalf("Failed to generate HDFC Regalia PDF: %v", err)
	}
	log.Println("✅ Generated samples/credit_cards/HDFC_Regalia_Credit_Card.pdf")

	// 5. HDFC Swiggy CC PDF
	if err := generateHDFCSwiggyPDF(filepath.Join(ccDir, "HDFC_Swiggy_Credit_Card.pdf")); err != nil {
		log.Fatalf("Failed to generate HDFC Swiggy PDF: %v", err)
	}
	log.Println("✅ Generated samples/credit_cards/HDFC_Swiggy_Credit_Card.pdf")

	// 6. HDFC RuPay CC PDF
	if err := generateHDFCRuPayPDF(filepath.Join(ccDir, "HDFC_RuPay_Credit_Card.pdf")); err != nil {
		log.Fatalf("Failed to generate HDFC RuPay PDF: %v", err)
	}
	log.Println("✅ Generated samples/credit_cards/HDFC_RuPay_Credit_Card.pdf")

	// 7. ICICI Amazon Pay CC PDF
	if err := generateICICIAmazonPDF(filepath.Join(ccDir, "ICICI_Amazon_Pay_Credit_Card.pdf")); err != nil {
		log.Fatalf("Failed to generate ICICI Amazon Pay PDF: %v", err)
	}
	log.Println("✅ Generated samples/credit_cards/ICICI_Amazon_Pay_Credit_Card.pdf")

	// 8. Axis Flipkart CC PDF
	if err := generateAxisFlipkartPDF(filepath.Join(ccDir, "Axis_Flipkart_Credit_Card.pdf")); err != nil {
		log.Fatalf("Failed to generate Axis Flipkart PDF: %v", err)
	}
	log.Println("✅ Generated samples/credit_cards/Axis_Flipkart_Credit_Card.pdf")

	// 9. HDFC Credit Card CSV
	if err := generateHDFCCC_CSV(filepath.Join(ccDir, "HDFC_Credit_Card_Statement.csv")); err != nil {
		log.Fatalf("Failed to generate HDFC CC CSV: %v", err)
	}
	log.Println("✅ Generated samples/credit_cards/HDFC_Credit_Card_Statement.csv")

	// 10. ICICI Savings Synthetic PDF
	if err := generateICICISavingsSyntheticPDF(filepath.Join(savingsDir, "ICICI_Savings_Synthetic.pdf")); err != nil {
		log.Fatalf("Failed to generate ICICI Savings Synthetic PDF: %v", err)
	}
	log.Println("✅ Generated samples/savings/ICICI_Savings_Synthetic.pdf")

	// 11. Union Bank Savings Synthetic PDF
	if err := generateUnionBankSavingsSyntheticPDF(filepath.Join(savingsDir, "Union_Bank_Savings_Synthetic.pdf")); err != nil {
		log.Fatalf("Failed to generate Union Bank Savings Synthetic PDF: %v", err)
	}
	log.Println("✅ Generated samples/savings/Union_Bank_Savings_Synthetic.pdf")

	log.Println("🎉 All masked sample statements created successfully!")
}

// -----------------------------------------------------------------------------
// 1. HDFC Savings / Current Account PDF
// -----------------------------------------------------------------------------
func generateHDFCSavingsPDF(filename string, isCurrent bool) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Bank Header
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(190, 7, "HDFC BANK LIMITED", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 8)
	accTypeStr := "SAVINGS ACCOUNT"
	accNoStr := "50100098760099"
	custName := "MR. RAHUL SHARMA"
	if isCurrent {
		accTypeStr = "CURRENT ACCOUNT"
		accNoStr = "50200012347496"
		custName = "TECHLABS INNOVATIONS PVT LTD"
	}

	pdf.CellFormat(95, 4, custName, "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Account Branch : INDIRANAGAR BANGALORE", "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "FLAT 402 PALM GROVE APTS 100 FT ROAD", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "RTGS/NEFT IFSC : HDFC0000123", "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "INDIRANAGAR BENGALURU 560038 KARNATAKA", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, fmt.Sprintf("Cust ID : 98765432 | Account Type : %s", accTypeStr), "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "Nomination : Registered", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, fmt.Sprintf("Account No : %s", accNoStr), "", 1, "R", false, 0, "")

	pdf.CellFormat(190, 4, "From : 01/08/2026 To : 30/08/2026", "", 1, "R", false, 0, "")
	pdf.Ln(4)

	// Table Header
	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(18, 6, "Date", "1", 0, "C", true, 0, "")
	pdf.CellFormat(72, 6, "Narration", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 6, "Chq./Ref.No.", "1", 0, "C", true, 0, "")
	pdf.CellFormat(18, 6, "Value Dt", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 6, "Withdrawal Amt.", "1", 0, "R", true, 0, "")
	pdf.CellFormat(18, 6, "Deposit Amt.", "1", 0, "R", true, 0, "")
	pdf.CellFormat(19, 6, "Closing Balance", "1", 1, "R", true, 0, "")

	// Sample Transactions
	txs := []struct {
		date, narr, ref, vdate, debit, credit, bal string
	}{
		{"01/08/26", "OPENING BALANCE", "", "01/08/26", "", "", "87,965.05"},
		{"02/08/26", "UPI-SWIGGY-swiggy@icici-YESB0000001-423561234567-PAYMENT", "423561234567", "02/08/26", "540.00", "", "87,425.05"},
		{"03/08/26", "UPI-BLINKIT-blinkit@axisbank-UTIB0000456-423598765432-GROCERIES", "423598765432", "03/08/26", "820.00", "", "86,605.05"},
		{"05/08/26", "SALARY CREDIT-ACME TECH LABS PVT LTD-AUG 2026", "SALARYAUG26", "05/08/26", "", "1,25,000.00", "2,11,605.05"},
		{"06/08/26", "NEFT DR-HDFCN000123456-CRED CLUB-CRED CC PAYMENT", "HDFCN000123456", "06/08/26", "44,110.56", "", "1,67,494.49"},
		{"08/08/26", "UPI-AMAZON-amazon@apl-HDFC0000123-423567890123-ORDER", "423567890123", "08/08/26", "1,899.00", "", "1,65,595.49"},
		{"10/08/26", "UPI-UBER INDIA-uber.india@hdfcbank-423612345678-RIDE", "423612345678", "10/08/26", "450.00", "", "1,65,145.49"},
		{"12/08/26", "UPI-STARBUCKS-starbucks@hdfcbank-423712345678-COFFEE", "423712345678", "12/08/26", "380.00", "", "1,64,765.49"},
		{"15/08/26", "ACH D-NETFLIX-MUMBAI-NETFLIX SUBSCRIPTION", "ACH891234", "15/08/26", "649.00", "", "1,64,116.49"},
		{"18/08/26", "UPI-SHELL PETROL-shell@icici-423812345678-FUEL", "423812345678", "18/08/26", "3,200.00", "", "1,60,916.49"},
		{"20/08/26", "UPI-ZOMATO-zomato@hdfcbank-423912345678-DINING", "423912345678", "20/08/26", "1,120.00", "", "1,59,796.49"},
		{"22/08/26", "UPI-APOLLO PHARMACY-apollo@axis-424012345678-MEDS", "424012345678", "22/08/26", "460.00", "", "1,59,336.49"},
		{"25/08/26", "UPI-DMART RETAIL-dmart@icici-424112345678-STORE", "424112345678", "25/08/26", "4,250.00", "", "1,55,086.49"},
		{"28/08/26", "UPI-SPOTIFY INDIA-spotify@hdfc-424212345678-MUSIC", "424212345678", "28/08/26", "119.00", "", "1,54,967.49"},
		{"30/08/26", "UPI-UPI LITE TOPUP-upilite@hdfc-424312345678-WALLET", "424312345678", "30/08/26", "1,000.00", "", "1,53,967.49"},
	}

	pdf.SetFont("Helvetica", "", 7)
	for _, tx := range txs {
		pdf.CellFormat(18, 5, tx.date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(72, 5, tx.narr, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 5, tx.ref, "1", 0, "C", false, 0, "")
		pdf.CellFormat(18, 5, tx.vdate, "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 5, tx.debit, "1", 0, "R", false, 0, "")
		pdf.CellFormat(18, 5, tx.credit, "1", 0, "R", false, 0, "")
		pdf.CellFormat(19, 5, tx.bal, "1", 1, "R", false, 0, "")
	}

	return pdf.OutputFileAndClose(filename)
}

// -----------------------------------------------------------------------------
// 2. HDFC Savings CSV
// -----------------------------------------------------------------------------
func generateHDFCSavingsCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	rows := [][]string{
		{"HDFC BANK - ACCOUNT STATEMENT"},
		{"Account No : 50100098760099", "Cust ID : 98765432", "Branch : INDIRANAGAR BANGALORE"},
		{"Account Type : SAVINGS ACCOUNT", "From : 01/08/2026", "To : 30/08/2026"},
		{},
		{"Date", "Narration", "Chq./Ref.No.", "Value Dt", "Withdrawal Amt.", "Deposit Amt.", "Closing Balance"},
		{"01/08/26", "OPENING BALANCE", "", "01/08/26", "", "", "87,965.05"},
		{"02/08/26", "UPI-SWIGGY-swiggy@icici-YESB0000001-423561234567-PAYMENT", "423561234567", "02/08/26", "540.00", "", "87,425.05"},
		{"03/08/26", "UPI-BLINKIT-blinkit@axisbank-UTIB0000456-423598765432-GROCERIES", "423598765432", "03/08/26", "820.00", "", "86,605.05"},
		{"05/08/26", "SALARY CREDIT-ACME TECH LABS PVT LTD-AUG 2026", "SALARYAUG26", "05/08/26", "", "1,25,000.00", "2,11,605.05"},
		{"06/08/26", "NEFT DR-HDFCN000123456-CRED CLUB-CRED CC PAYMENT", "HDFCN000123456", "06/08/26", "44,110.56", "", "1,67,494.49"},
		{"08/08/26", "UPI-AMAZON-amazon@apl-HDFC0000123-423567890123-ORDER", "423567890123", "08/08/26", "1,899.00", "", "1,65,595.49"},
		{"10/08/26", "UPI-UBER INDIA-uber.india@hdfcbank-423612345678-RIDE", "423612345678", "10/08/26", "450.00", "", "1,65,145.49"},
		{"18/08/26", "UPI-SHELL PETROL-shell@icici-423812345678-FUEL", "423812345678", "18/08/26", "3,200.00", "", "1,60,916.49"},
		{"25/08/26", "UPI-DMART RETAIL-dmart@icici-424112345678-STORE", "424112345678", "25/08/26", "4,250.00", "", "1,56,666.49"},
	}

	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// 3. HDFC Regalia CC PDF
// -----------------------------------------------------------------------------
func generateHDFCRegaliaPDF(filename string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(190, 7, "HDFC BANK CREDIT CARD STATEMENT", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(190, 5, "Regalia Credit Card", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 8)
	pdf.CellFormat(95, 4, "Name : RAHUL SHARMA", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Card No. : 4111 XXXX XXXX 3638", "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "Statement Date : 15 Aug, 2026", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Billing Period : 16 Jul, 2026 - 15 Aug, 2026", "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "DUE DATE : 04 Sep, 2026", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Total Amount Due : Rs. 34,500.00", "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "Minimum Amount Due : Rs. 1,725.00", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Credit Limit : Rs. 4,00,000.00 | Available Limit : Rs. 3,65,500.00", "", 1, "R", false, 0, "")

	pdf.CellFormat(190, 4, "Reward Points : 12,450", "", 1, "L", false, 0, "")
	pdf.Ln(4)

	pdf.SetFont("Helvetica", "B", 8)
	pdf.CellFormat(190, 5, "DOMESTIC TRANSACTIONS", "B", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 7)
	pdf.CellFormat(25, 5, "Date", "1", 0, "C", false, 0, "")
	pdf.CellFormat(130, 5, "Transaction Description", "1", 0, "L", false, 0, "")
	pdf.CellFormat(35, 5, "Amount (Rs.)", "1", 1, "R", false, 0, "")

	txs := []struct {
		date, desc, amt string
	}{
		{"18/07/2026", "MAKEMYTRIP FLIGHT TICKETS GURGAON", "12,450.00"},
		{"22/07/2026", "APPLE SERVICES RETAIL STORE MUMBAI", "199.00"},
		{"28/07/2026", "TAJ HOTELS RESORTS MUMBAI", "14,800.00"},
		{"02/08/2026", "SHELL AUTO FUEL BENGALURU", "3,500.00"},
		{"06/08/2026", "ZOMATO RESTAURANT BANGALORE", "1,850.00"},
		{"09/08/2026", "BLINKIT GROCERIES BENGALURU", "1,701.00"},
		{"10/08/2026", "NETBANKING CC PAYMENT RECEIVED", "+ 25,000.00"},
	}

	pdf.SetFont("Helvetica", "", 7)
	for _, tx := range txs {
		pdf.CellFormat(25, 5, tx.date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(130, 5, tx.desc, "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 5, tx.amt, "1", 1, "R", false, 0, "")
	}

	return pdf.OutputFileAndClose(filename)
}

// -----------------------------------------------------------------------------
// 4. HDFC Swiggy CC PDF
// -----------------------------------------------------------------------------
func generateHDFCSwiggyPDF(filename string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(190, 7, "HDFC BANK CREDIT CARD STATEMENT", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(190, 5, "Swiggy HDFC Bank Credit Card", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 8)
	pdf.CellFormat(95, 4, "Name : RAHUL SHARMA", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Card No. : 4111 XXXX XXXX 0826", "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "Statement Date : 20 Aug, 2026", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Billing Period : 21 Jul, 2026 - 20 Aug, 2026", "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "DUE DATE : 09 Sep, 2026", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Total Amount Due : Rs. 18,240.50", "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "Minimum Amount Due : Rs. 920.00", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Credit Limit : Rs. 2,50,000.00 | Available Limit : Rs. 2,31,759.50", "", 1, "R", false, 0, "")

	pdf.CellFormat(190, 4, "Cashback Earned this cycle : Rs. 1,240.00", "", 1, "L", false, 0, "")
	pdf.Ln(4)

	pdf.SetFont("Helvetica", "B", 8)
	pdf.CellFormat(190, 5, "DOMESTIC TRANSACTIONS", "B", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 7)
	pdf.CellFormat(25, 5, "Date", "1", 0, "C", false, 0, "")
	pdf.CellFormat(130, 5, "Transaction Description", "1", 0, "L", false, 0, "")
	pdf.CellFormat(35, 5, "Amount (Rs.)", "1", 1, "R", false, 0, "")

	txs := []struct {
		date, desc, amt string
	}{
		{"22/07/2026", "SWIGGY FOOD DELIVERY BANGALORE", "840.00"},
		{"25/07/2026", "SWIGGY INSTAMART BANGALORE", "1,450.00"},
		{"29/07/2026", "AMAZON INDIA INTERNET", "4,200.00"},
		{"03/08/2026", "FLIPKART PAYMENTS BANGALORE", "3,150.00"},
		{"08/08/2026", "DINE OUT RESTAURANT INDIRANAGAR", "2,600.00"},
		{"12/08/2026", "SWIGGY GENIE COURIER", "250.00"},
		{"14/08/2026", "MYNTRA DESIGNS BANGALORE", "5,750.50"},
		{"16/08/2026", "AUTOPAY CC BILL PAYMENT", "+ 15,000.00"},
	}

	pdf.SetFont("Helvetica", "", 7)
	for _, tx := range txs {
		pdf.CellFormat(25, 5, tx.date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(130, 5, tx.desc, "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 5, tx.amt, "1", 1, "R", false, 0, "")
	}

	return pdf.OutputFileAndClose(filename)
}

// -----------------------------------------------------------------------------
// 5. HDFC RuPay CC PDF
// -----------------------------------------------------------------------------
func generateHDFCRuPayPDF(filename string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(190, 7, "HDFC BANK CREDIT CARD STATEMENT", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(190, 5, "RuPay Virtual UPI Credit Card", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 8)
	pdf.CellFormat(95, 4, "Name : RAHUL SHARMA", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Card No. : 6521 XXXX XXXX 0161", "", 1, "R", false, 0, "")

	pdf.SetFont("Helvetica", "", 8)
	pdf.CellFormat(95, 4, "Statement Date : 10 Aug, 2026", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Billing Period : 11 Jul, 2026 - 10 Aug, 2026", "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "DUE DATE : 30 Aug, 2026", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Total Amount Due : Rs. 9,850.00", "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "Minimum Amount Due : Rs. 500.00", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Credit Limit : Rs. 1,00,000.00 | Available Limit : Rs. 90,150.00", "", 1, "R", false, 0, "")
	pdf.Ln(4)

	pdf.SetFont("Helvetica", "B", 8)
	pdf.CellFormat(190, 5, "DOMESTIC TRANSACTIONS", "B", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 7)
	pdf.CellFormat(25, 5, "Date", "1", 0, "C", false, 0, "")
	pdf.CellFormat(130, 5, "Transaction Description", "1", 0, "L", false, 0, "")
	pdf.CellFormat(35, 5, "Amount (Rs.)", "1", 1, "R", false, 0, "")

	txs := []struct {
		date, desc, amt string
	}{
		{"12/07/2026", "UPI/CHAIPATTI CAFE/chai@upi", "120.00"},
		{"14/07/2026", "UPI/NILGIRIS SUPERMARKET/nilg@icici", "1,420.00"},
		{"19/07/2026", "UPI/APOLLO PHARMACY/apollo@hdfc", "650.00"},
		{"24/07/2026", "UPI/INDIAN OIL PETROL/ioc@axis", "2,100.00"},
		{"02/08/2026", "UPI/LOCAL BAKERY STORE/bake@paytm", "340.00"},
		{"05/08/2026", "UPI/METRO SMART CARD RECHARGE", "500.00"},
		{"08/08/2026", "UPI/VEGETABLE VENDOR/veg@upi", "220.00"},
		{"09/08/2026", "IMPS CC PAYMENT RECEIVED", "+ 12,000.00"},
	}

	pdf.SetFont("Helvetica", "", 7)
	for _, tx := range txs {
		pdf.CellFormat(25, 5, tx.date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(130, 5, tx.desc, "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 5, tx.amt, "1", 1, "R", false, 0, "")
	}

	return pdf.OutputFileAndClose(filename)
}

// -----------------------------------------------------------------------------
// 6. ICICI Amazon Pay CC PDF
// -----------------------------------------------------------------------------
func generateICICIAmazonPDF(filename string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(190, 7, "ICICI Bank Credit Card Statement", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(190, 4, "Amazon Pay ICICI Bank Credit Card", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 8)
	pdf.CellFormat(95, 4, "Name: RAHUL SHARMA", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Card Number: 4315XXXXXXXX0001", "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "Statement period: July 13, 2026 to August 12, 2026", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Payment Due Date: 30/08/2026", "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "Statement Date: 12/08/2026", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Total Amount Due: 44,110.56", "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "Minimum Amount Due: 2,210.00", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Credit Limit: 5,60,000.00 | Available Limit: 5,15,889.44", "", 1, "R", false, 0, "")
	pdf.Ln(4)

	pdf.SetFont("Helvetica", "B", 8)
	pdf.CellFormat(190, 5, "PURCHASES / CHARGES", "B", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 7)
	pdf.CellFormat(25, 5, "Date", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 5, "Reference No", "1", 0, "C", false, 0, "")
	pdf.CellFormat(95, 5, "Details", "1", 0, "L", false, 0, "")
	pdf.CellFormat(40, 5, "Amount (INR)", "1", 1, "R", false, 0, "")

	txs := []struct {
		date, ref, desc, amt string
	}{
		{"15/07/2026", "13777833771", "Amazon Pay IN Utility IN", "8,100.00"},
		{"16/07/2026", "13803091904", "NETFLIX DI SI IN", "199.00"},
		{"17/07/2026", "13808011381", "Amazon Pay IN Grocery IN", "916.00"},
		{"18/07/2026", "13814542193", "Amazon Pay IN E Commerce", "331.00"},
		{"25/07/2026", "13815281667", "SANDHYA FUEL STATION KAMAREDDY", "3,974.03"},
		{"02/08/2026", "13935496199", "BLINK COMMERCE IN", "806.00"},
		{"08/08/2026", "13949195683", "PURE RIDE TECHNOLOGIES PVT NEW IN", "10,000.00"},
		{"10/08/2026", "99887766554", "PAYMENT RECEIVED - THANK YOU", "30,000.00 CR"},
	}

	pdf.SetFont("Helvetica", "", 7)
	for _, tx := range txs {
		pdf.CellFormat(25, 5, tx.date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 5, tx.ref, "1", 0, "C", false, 0, "")
		pdf.CellFormat(95, 5, tx.desc, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 5, tx.amt, "1", 1, "R", false, 0, "")
	}

	return pdf.OutputFileAndClose(filename)
}

// -----------------------------------------------------------------------------
// 7. Axis Flipkart CC PDF
// -----------------------------------------------------------------------------
func generateAxisFlipkartPDF(filename string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(190, 7, "Axis Bank Credit Card Statement", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(190, 4, "Flipkart Axis Bank Credit Card", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 8)
	pdf.CellFormat(95, 4, "Customer Name: RAHUL SHARMA", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Card No: 5241XXXXXXXX0308", "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "Statement Period : 15/06/2026 - 13/07/2026", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Payment Due Date : 02/08/2026", "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "Statement Generation Date : 13/07/2026", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Total Payment Due : 11,000.82", "", 1, "R", false, 0, "")

	pdf.CellFormat(95, 4, "Minimum Payment Due : 3,987.00", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, "Credit Limit : 2,96,000.00 | Available Credit Limit : 2,77,516.18", "", 1, "R", false, 0, "")

	pdf.CellFormat(190, 4, "Cashback Earned: 371.00", "", 1, "L", false, 0, "")
	pdf.Ln(4)

	pdf.SetFont("Helvetica", "B", 8)
	pdf.CellFormat(190, 5, "TRANSACTION DETAILS", "B", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 7)
	pdf.CellFormat(25, 5, "Date", "1", 0, "C", false, 0, "")
	pdf.CellFormat(100, 5, "Transaction Details", "1", 0, "L", false, 0, "")
	pdf.CellFormat(35, 5, "Reference Number", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 5, "Amount (INR)", "1", 1, "R", false, 0, "")

	txs := []struct {
		date, desc, ref, amt string
	}{
		{"13/06/2026", "FLIPKART PAYMENTS, BANGALORE", "70165001", "434.00 Dr"},
		{"18/06/2026", "EMI PRINCIPAL 4/6 REF# 70165012", "70165012", "3,668.00 Dr"},
		{"27/06/2026", "PAYMENT RECEIVED VIA CRED", "70165099", "20,007.26 Cr"},
		{"05/07/2026", "FLIPKART INTERNET PVT NOIDA", "70165110", "7,009.00 Dr"},
		{"09/07/2026", "CASHBACK CREDIT JUN26 MYNTRA", "70165120", "286.00 Cr"},
	}

	pdf.SetFont("Helvetica", "", 7)
	for _, tx := range txs {
		pdf.CellFormat(25, 5, tx.date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(100, 5, tx.desc, "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 5, tx.ref, "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 5, tx.amt, "1", 1, "R", false, 0, "")
	}

	return pdf.OutputFileAndClose(filename)
}

// -----------------------------------------------------------------------------
// 8. HDFC Credit Card CSV
// -----------------------------------------------------------------------------
func generateHDFCCC_CSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	rows := [][]string{
		{"HDFC BANK CREDIT CARD STATEMENT"},
		{"Card Number : 4111XXXXXXXX3638", "Name : RAHUL SHARMA"},
		{"Statement Date : 15/08/2026", "Payment Due Date : 04/09/2026"},
		{"Total Due : 34500.00", "Minimum Due : 1725.00", "Credit Limit : 400000.00"},
		{},
		{"Date", "Transaction Description", "Amount", "Type"},
		{"18/07/2026", "MAKEMYTRIP FLIGHT TICKETS", "12450.00", "DEBIT"},
		{"22/07/2026", "APPLE SERVICES RETAIL", "199.00", "DEBIT"},
		{"28/07/2026", "TAJ HOTELS RESORTS", "14800.00", "DEBIT"},
		{"02/08/2026", "SHELL AUTO FUEL", "3500.00", "DEBIT"},
		{"10/08/2026", "NETBANKING CC PAYMENT RECEIVED", "25000.00", "CREDIT"},
	}

	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// 9. ICICI Bank Savings Synthetic PDF
// -----------------------------------------------------------------------------
func generateICICISavingsSyntheticPDF(filename string) error {
	return writeBankHistorySyntheticPDF(filename, "ICICI BANK", "Statement of Transactions in Saving Account",
		"Statement Period: 10/04/2026 to 11/04/2026",
		[][]string{
			{"Transaction Date", "Transaction Remarks", "Withdrawal", "Deposit", "Balance"},
			{"10.04.2026", "UPI/DEMO CAFE/demo-cafe@upi", "100.00", "", "900.00 Cr"},
			{"11.04.2026", "NEFT CR/EXAMPLE EMPLOYER", "", "250.00", "1,150.00 Cr"},
		},
	)
}

// -----------------------------------------------------------------------------
// 10. Union Bank of India Savings Synthetic PDF
// -----------------------------------------------------------------------------
func generateUnionBankSavingsSyntheticPDF(filename string) error {
	return writeBankHistorySyntheticPDF(filename, "UNION BANK OF INDIA (UBIN)", "DETAILS OF STATEMENT",
		"Period From 10-04-2026 to 11-04-2026",
		[][]string{
			{"Date", "Particulars", "Chq Num", "Withdrawal", "Deposit", "Balance"},
			{"10-04-2026", "UPI/DR/123456789012/DEMO STORE", "", "100.00", "", "900.00 Cr"},
			{"11-04-2026", "NEFT CR/EXAMPLE PAYER", "", "", "250.00", "1,150.00 Cr"},
		},
	)
}

func writeBankHistorySyntheticPDF(path, bank, title, period string, rows [][]string) error {
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
