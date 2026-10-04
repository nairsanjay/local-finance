package service

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"local-finance/internal/db"
	"local-finance/internal/models"
	"local-finance/internal/parser"
	"local-finance/internal/parser/extractor"
)

func TestPreviewStatementService(t *testing.T) {
	testDBPath := filepath.Join(os.TempDir(), "test_preview.db")
	_ = os.Remove(testDBPath)
	defer os.Remove(testDBPath)

	database, err := db.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize test db: %v", err)
	}
	defer database.Close()

	svc := NewTransactionService(database)
	samplesDir := filepath.Join("..", "..", "samples")

	// Test 1: HDFC Swiggy CC (Unencrypted Sample)
	swiggyPath := filepath.Join(samplesDir, "credit_cards", "HDFC_Swiggy_Credit_Card.pdf")
	f1, err := os.Open(swiggyPath)
	if err != nil {
		t.Fatalf("Sample statement not accessible: %v", err)
	}
	defer f1.Close()

	if extractor.IsPDFEncrypted(f1) {
		t.Errorf("Expected HDFC_Swiggy_Credit_Card.pdf to NOT be encrypted")
	}
	_, _ = f1.Seek(0, 0)

	res1, err := svc.PreviewStatement("HDFC_Swiggy_Credit_Card.pdf", f1, "", "", "")
	if err != nil {
		t.Fatalf("Preview failed for HDFC_Swiggy_Credit_Card.pdf: %v", err)
	}

	if res1.TotalTransactions < 5 {
		t.Errorf("Expected at least 5 transactions, got %d", res1.TotalTransactions)
	}
	if res1.BankName != "HDFC Bank" {
		t.Errorf("Expected HDFC Bank, got %s", res1.BankName)
	}
	if res1.AccountType != "CREDIT_CARD" {
		t.Errorf("Expected CREDIT_CARD, got %s", res1.AccountType)
	}

	// Test 2: Encrypted HDFC Savings PDF without password -> should report requires_password
	savingsPath := filepath.Join(samplesDir, "savings", "HDFC_Savings_Account_Statement.pdf")
	rawSavings, err := os.ReadFile(savingsPath)
	if err != nil {
		t.Fatalf("Failed to read savings sample: %v", err)
	}

	savingsPass := "TestPass123"
	confSavings := model.NewAESConfiguration(savingsPass, savingsPass, 256)
	var encSavingsBuf bytes.Buffer
	err = api.Encrypt(bytes.NewReader(rawSavings), &encSavingsBuf, confSavings)
	if err != nil {
		t.Fatalf("Failed to encrypt savings PDF: %v", err)
	}
	encSavingsBytes := encSavingsBuf.Bytes()

	res2, err := svc.PreviewStatement("HDFC_Savings_Protected.pdf", bytes.NewReader(encSavingsBytes), "", "", "")
	if err != nil {
		t.Fatalf("Preview failed on encrypted PDF: %v", err)
	}
	if !res2.RequiresPassword {
		t.Errorf("Expected encrypted PDF to flag RequiresPassword=true")
	}

	// Unlock with password
	res2Unlocked, err := svc.PreviewStatement("HDFC_Savings_Protected.pdf", bytes.NewReader(encSavingsBytes), "", "", savingsPass)
	if err != nil || res2Unlocked.TotalTransactions == 0 {
		t.Errorf("Expected transactions with password, got %d (err=%v)", res2Unlocked.TotalTransactions, err)
	}

	// Test 3: Amazon Pay ICICI encrypted PDF
	amazonPath := filepath.Join(samplesDir, "credit_cards", "ICICI_Amazon_Pay_Credit_Card.pdf")
	rawAmazon, err := os.ReadFile(amazonPath)
	if err != nil {
		t.Fatalf("Failed to read amazon CC sample: %v", err)
	}

	amazonPass := "mypassword"
	confAmazon := model.NewAESConfiguration(amazonPass, amazonPass, 256)
	var encBuf bytes.Buffer
	err = api.Encrypt(bytes.NewReader(rawAmazon), &encBuf, confAmazon)
	if err != nil {
		t.Fatalf("Failed to encrypt amazon PDF: %v", err)
	}
	encBytes := encBuf.Bytes()

	// Test without password -> should report requires_password
	res3, err := svc.PreviewStatement("ICICI_Amazon_Pay_Credit_Card.pdf", bytes.NewReader(encBytes), "", "", "")
	if err != nil {
		t.Fatalf("Preview failed: %v", err)
	}
	if !res3.RequiresPassword {
		t.Errorf("Expected encrypted amazon PDF to require password")
	}

	// Test with password -> should decrypt and parse transactions
	res3Unlocked, err := svc.PreviewStatement("ICICI_Amazon_Pay_Credit_Card.pdf", bytes.NewReader(encBytes), "", "", amazonPass)
	if err != nil {
		t.Fatalf("Preview with password failed: %v", err)
	}
	if res3Unlocked.TotalTransactions == 0 {
		t.Errorf("Expected transactions with password, got 0 (Error: %s)", res3Unlocked.Error)
	}
	if res3Unlocked.BankName != "ICICI Bank" {
		t.Errorf("Expected ICICI Bank, got %s", res3Unlocked.BankName)
	}
	if res3Unlocked.AccountType != "CREDIT_CARD" {
		t.Errorf("Expected CREDIT_CARD, got %s", res3Unlocked.AccountType)
	}

	// Test import with password
	impRes, err := svc.ImportStatement("ICICI_Amazon_Pay_Credit_Card.pdf", bytes.NewReader(encBytes), "", "", amazonPass)
	if err != nil {
		t.Fatalf("ImportStatement failed: %v", err)
	}
	if impRes.TotalParsed == 0 {
		t.Errorf("Expected total parsed transactions > 0, got %d", impRes.TotalParsed)
	}
	if impRes.InsertedCount == 0 {
		t.Errorf("Expected inserted transactions > 0, got %d", impRes.InsertedCount)
	}
}

func TestSavingsAndCurrentBulkAccountIsolation(t *testing.T) {
	samplesDir := filepath.Join("..", "..", "samples")
	savingsPath := filepath.Join(samplesDir, "savings", "HDFC_Savings_Account_Statement.pdf")
	currentPath := filepath.Join(samplesDir, "savings", "HDFC_Current_Account_Statement.pdf")

	sData, err1 := os.ReadFile(savingsPath)
	cData, err2 := os.ReadFile(currentPath)
	if err1 != nil || err2 != nil {
		t.Skip("Sample statements not accessible")
	}

	testDBPath := filepath.Join(os.TempDir(), "test_isolation.db")
	_ = os.Remove(testDBPath)
	defer os.Remove(testDBPath)

	database, err := db.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to init DB: %v", err)
	}
	defer database.Close()

	svc := NewTransactionService(database)

	// 1. Import Savings Statement
	sRes, err := svc.ImportStatement(filepath.Base(savingsPath), bytes.NewReader(sData), "", "", "")
	if err != nil {
		t.Fatalf("Failed to import savings: %v", err)
	}
	if sRes.TotalParsed == 0 {
		t.Errorf("Expected savings transactions, got %d", sRes.TotalParsed)
	}

	// 2. Import Current Statement
	cRes, err := svc.ImportStatement(filepath.Base(currentPath), bytes.NewReader(cData), "", "", "")
	if err != nil {
		t.Fatalf("Failed to import current: %v", err)
	}
	if cRes.TotalParsed == 0 {
		t.Errorf("Expected current transactions, got %d", cRes.TotalParsed)
	}

	// 3. Verify accounts are separate and dedicated
	accounts, err := database.ListAccounts()
	if err != nil {
		t.Fatalf("Failed to list accounts: %v", err)
	}
	if len(accounts) != 2 {
		t.Fatalf("Expected 2 separate accounts, got %d", len(accounts))
	}

	var savingsAcc, currentAcc *models.Account
	for i := range accounts {
		if accounts[i].AccountType == models.AccountTypeSavings {
			savingsAcc = &accounts[i]
		} else if accounts[i].AccountType == models.AccountTypeCurrent {
			currentAcc = &accounts[i]
		}
	}

	if savingsAcc == nil {
		t.Fatal("Savings account not found in DB")
	}
	if currentAcc == nil {
		t.Fatal("Current account not found in DB")
	}

	if savingsAcc.ID == currentAcc.ID {
		t.Errorf("Savings and Current accounts have the SAME ID: %s", savingsAcc.ID)
	}
	if savingsAcc.AccountNumberMask != "XX0099" {
		t.Errorf("Savings AccountNumberMask mismatch: %s", savingsAcc.AccountNumberMask)
	}
	if currentAcc.AccountNumberMask != "XX7496" {
		t.Errorf("Current AccountNumberMask mismatch: %s", currentAcc.AccountNumberMask)
	}
}

func TestMatchCategoryExceptionsAndTxType(t *testing.T) {
	svc := &TransactionService{}

	rules := []models.CategorizationRule{
		{
			ID:               "rule_salary",
			Priority:         100,
			MatchField:       "raw_narration",
			MatchType:        "REGEX",
			MatchPattern:     "(?i)SALARY|SAL CREDIT",
			ExcludePattern:   "maid, driver, cook, helper, staff, advance",
			TxType:           "CREDIT",
			TargetCategoryID: "cat_salary",
			IsActive:         true,
		},
		{
			ID:               "rule_swiggy_food",
			Priority:         90,
			MatchField:       "cleaned_payee",
			MatchType:        "CONTAINS",
			MatchPattern:     "SWIGGY",
			ExcludePattern:   "INSTAMART, GENIE",
			TxType:           "DEBIT",
			TargetCategoryID: "cat_food",
			IsActive:         true,
		},
		{
			ID:               "rule_instamart",
			Priority:         85,
			MatchField:       "cleaned_payee",
			MatchType:        "CONTAINS",
			MatchPattern:     "INSTAMART",
			TxType:           "ALL",
			TargetCategoryID: "cat_groceries",
			IsActive:         true,
		},
	}

	// 1. Legit salary credit -> cat_salary
	res1 := svc.matchCategory(parser.ParsedTransaction{
		RawNarration: "ACH SALARY CREDIT TECH CORP",
		CleanedPayee: "TECH CORP",
		TxType:       models.TxTypeCredit,
	}, rules)
	if res1 == nil || *res1 != "cat_salary" {
		t.Errorf("Expected cat_salary for legit credit, got %v", res1)
	}

	// 2. Maid salary debit -> should NOT match salary rule (debit, plus exception 'maid')
	res2 := svc.matchCategory(parser.ParsedTransaction{
		RawNarration: "UPI-MAID SALARY-MONTHLY",
		CleanedPayee: "SHANTI MAID",
		TxType:       models.TxTypeDebit,
	}, rules)
	if res2 != nil && *res2 == "cat_salary" {
		t.Errorf("Expected maid salary debit to NOT match cat_salary, got %s", *res2)
	}

	// 3. Driver salary debit -> should NOT match salary rule
	res3 := svc.matchCategory(parser.ParsedTransaction{
		RawNarration: "NETBANKING DRIVER SALARY ADVANCE",
		CleanedPayee: "RAMESH DRIVER",
		TxType:       models.TxTypeDebit,
	}, rules)
	if res3 != nil && *res3 == "cat_salary" {
		t.Errorf("Expected driver salary debit to NOT match cat_salary, got %s", *res3)
	}

	// 4. Swiggy normal food debit -> cat_food
	res4 := svc.matchCategory(parser.ParsedTransaction{
		RawNarration: "UPI/SWIGGY/123456",
		CleanedPayee: "SWIGGY",
		TxType:       models.TxTypeDebit,
	}, rules)
	if res4 == nil || *res4 != "cat_food" {
		t.Errorf("Expected cat_food for Swiggy, got %v", res4)
	}

	// 5. Swiggy Instamart -> should skip cat_food (exception 'INSTAMART') and match cat_groceries
	res5 := svc.matchCategory(parser.ParsedTransaction{
		RawNarration: "POS SWIGGY INSTAMART BANGALORE",
		CleanedPayee: "SWIGGY INSTAMART",
		TxType:       models.TxTypeDebit,
	}, rules)
	if res5 == nil || *res5 != "cat_groceries" {
		t.Errorf("Expected cat_groceries for Swiggy Instamart, got %v", res5)
	}
}
