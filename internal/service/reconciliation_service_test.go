package service

import (
	"path/filepath"
	"testing"
	"time"

	"local-finance/internal/db"
	"local-finance/internal/models"
)

func TestTransferReconciliationAndMerchants(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_reconcile.db")

	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}
	defer database.Close()

	// 1. Create Savings & Credit Card accounts
	savingsAcc, err := database.GetOrCreateAccount("HDFC Bank", models.AccountTypeSavings, "", "XX0099", "", "", "", "", "", "RAHUL SHARMA", nil)
	if err != nil {
		t.Fatalf("Failed to create savings account: %v", err)
	}

	limit := 500000.0
	ccAcc, err := database.GetOrCreateAccount("ICICI Bank", models.AccountTypeCreditCard, "", "XX0001", "", "", "", "VISA", "Amazon Pay ICICI", "RAHUL SHARMA", &limit)
	if err != nil {
		t.Fatalf("Failed to create credit card account: %v", err)
	}

	// 2. Insert paired CC Bill Payment transactions
	// Savings Debit: ₹12,850 on 2026-08-10 with narration "CRED CLUB BILL PAYMENT"
	txDebit := &models.Transaction{
		AccountID:       savingsAcc.ID,
		TxHash:          "hash_savings_debit_12850",
		TxDate:          "2026-08-10",
		RawNarration:    "UPI/CRED CLUB/12850/CC PAYMENT",
		CleanedPayee:    "Cred Club",
		PaymentMode:     models.PaymentModeUPI,
		TxType:          models.TxTypeDebit,
		Amount:          12850.0,
		IsTransfer:      false,
		IsExcluded:      false,
		CreatedAt:       time.Now(),
	}
	if _, err := database.UpsertTransaction(txDebit); err != nil {
		t.Fatalf("Failed to insert debit tx: %v", err)
	}

	// CC Credit: ₹12,850 on 2026-08-11 with narration "PAYMENT RECEIVED THANKS"
	txCredit := &models.Transaction{
		AccountID:       ccAcc.ID,
		TxHash:          "hash_cc_credit_12850",
		TxDate:          "2026-08-11",
		RawNarration:    "PAYMENT RECEIVED VIA IMPS",
		CleanedPayee:    "Credit Card Bill Payment",
		PaymentMode:     models.PaymentModeCardPOS,
		TxType:          models.TxTypeCredit,
		Amount:          12850.0,
		IsTransfer:      false,
		IsExcluded:      false,
		CreatedAt:       time.Now(),
	}
	if _, err := database.UpsertTransaction(txCredit); err != nil {
		t.Fatalf("Failed to insert credit tx: %v", err)
	}

	// 3. Insert merchant transactions (Swiggy orders)
	for i := 1; i <= 3; i++ {
		swiggyTx := &models.Transaction{
			AccountID:       ccAcc.ID,
			TxHash:          "hash_swiggy_" + string(rune('0'+i)),
			TxDate:          "2026-08-1" + string(rune('0'+i)),
			RawNarration:    "POS 40124300 SWIGGY BANGALORE IN",
			CleanedPayee:    "Swiggy",
			PaymentMode:     models.PaymentModeCardPOS,
			TxType:          models.TxTypeDebit,
			Amount:          500.0 * float64(i),
			IsTransfer:      false,
			IsExcluded:      false,
			CreatedAt:       time.Now(),
		}
		if _, err := database.UpsertTransaction(swiggyTx); err != nil {
			t.Fatalf("Failed to insert swiggy tx: %v", err)
		}
	}

	// 4. Test Reconciliation Service
	reconService := NewReconciliationService(database)
	autoLinked, _, err := reconService.ScanAndAutoReconcile()
	if err != nil {
		t.Fatalf("ScanAndAutoReconcile failed: %v", err)
	}
	if autoLinked != 1 {
		t.Errorf("Expected 1 auto-linked CC payment pair, got %d", autoLinked)
	}

	summary, err := reconService.GetSummary()
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}
	if summary.TotalPairedTransfers != 1 {
		t.Errorf("Expected 1 paired transfer in summary, got %d", summary.TotalPairedTransfers)
	}
	if summary.DoubleCountPreventedAmount != 12850.0 {
		t.Errorf("Expected ₹12,850 double count prevented, got %.2f", summary.DoubleCountPreventedAmount)
	}

	// 5. Test Merchant Intelligence
	merchantsList, err := database.ListMerchants("", "", "total_spend")
	if err != nil {
		t.Fatalf("ListMerchants failed: %v", err)
	}
	if len(merchantsList.Merchants) == 0 {
		t.Fatalf("Expected merchants in list, got 0")
	}

	var swiggySummary *models.MerchantSummaryItem
	for _, m := range merchantsList.Merchants {
		if m.CleanedPayee == "Swiggy" {
			swiggySummary = &m
			break
		}
	}
	if swiggySummary == nil {
		t.Fatalf("Swiggy merchant summary not found")
	}
	if swiggySummary.TxCount != 3 {
		t.Errorf("Expected 3 Swiggy transactions, got %d", swiggySummary.TxCount)
	}
	if swiggySummary.TotalSpend != 3000.0 { // 500 + 1000 + 1500 = 3000
		t.Errorf("Expected total spend 3000, got %.2f", swiggySummary.TotalSpend)
	}

	// Test Merchant Deep-Dive Profile
	profile, err := database.GetMerchantProfile("Swiggy")
	if err != nil {
		t.Fatalf("GetMerchantProfile failed: %v", err)
	}
	if profile.CleanedPayee != "Swiggy" {
		t.Errorf("Expected CleanedPayee Swiggy, got %s", profile.CleanedPayee)
	}
	if profile.TotalSpend != 3000.0 {
		t.Errorf("Expected TotalSpend 3000, got %.2f", profile.TotalSpend)
	}
	if len(profile.RecentTransactions) != 3 {
		t.Errorf("Expected 3 recent transactions in profile, got %d", len(profile.RecentTransactions))
	}
}
