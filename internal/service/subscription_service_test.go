package service

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"local-finance/internal/db"
	"local-finance/internal/models"
)

func TestSubscriptionDetection(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_subs.db")

	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer database.Close()

	// 1. Create a test account
	acc, err := database.GetOrCreateAccount("HDFC Bank", models.AccountTypeCreditCard, "", "XX1234", "", "", "", "", "", "RAHUL SHARMA", nil)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// 2. Insert recurring debit transactions for Netflix, YouTube, Apple
	txs := []models.Transaction{
		{
			AccountID:       acc.ID,
			TxDate:          "2026-07-16",
			TxType:          models.TxTypeDebit,
			Amount:          199.00,
			CleanedPayee:    "Netflix Di Si In",
			RawNarration:    "NETFLIX DI SI MUMBAI IN",
			PaymentMode:     models.PaymentModeCardPOS,
			ReferenceNumber: "REF1",
		},
		{
			AccountID:       acc.ID,
			TxDate:          "2026-08-16",
			TxType:          models.TxTypeDebit,
			Amount:          199.00,
			CleanedPayee:    "Netflix Di Si In",
			RawNarration:    "NETFLIX DI SI MUMBAI IN",
			PaymentMode:     models.PaymentModeCardPOS,
			ReferenceNumber: "REF2",
		},
		{
			AccountID:       acc.ID,
			TxDate:          "2026-08-05",
			TxType:          models.TxTypeDebit,
			Amount:          75.00,
			CleanedPayee:    "Applemedia Services",
			RawNarration:    "UPI-AUTOPAY-APPLEMEDIA SERVICES-APPLESERVICES.BDSI@HDFCBANK",
			PaymentMode:     models.PaymentModeUPI,
			ReferenceNumber: "REF3",
		},
		{
			AccountID:       acc.ID,
			TxDate:          "2026-08-27",
			TxType:          models.TxTypeDebit,
			Amount:          299.00,
			CleanedPayee:    "Youtube",
			RawNarration:    "UPI-AUTOPAY-YOUTUBE-YOUTUBE1.BD@AXISBANK-MANDATEEXECUTE",
			PaymentMode:     models.PaymentModeUPI,
			ReferenceNumber: "REF4",
		},
	}

	for i := range txs {
		txs[i].TxHash = fmt.Sprintf("test_hash_%d", i)
		_, err := database.UpsertTransaction(&txs[i])
		if err != nil {
			t.Fatalf("Failed to upsert tx: %v", err)
		}
	}

	// 3. Run Subscription Service Scanner
	subService := NewSubscriptionService(database)
	summary, err := subService.ScanAndDetectSubscriptions()
	if err != nil {
		t.Fatalf("ScanAndDetectSubscriptions failed: %v", err)
	}

	if summary.TotalActive < 3 {
		t.Errorf("Expected at least 3 active subscriptions, got %d", summary.TotalActive)
	}

	// Verify burn rate (199 + 75 + 299 = 573)
	if summary.MonthlyBurnRate < 500 {
		t.Errorf("Expected monthly burn rate around ~573, got %.2f", summary.MonthlyBurnRate)
	}

	// 4. Test manual subscription creation
	customSub := &models.Subscription{
		Name:            "Gym Membership",
		MerchantPattern: "CULT FIT",
		Frequency:       models.FrequencyMonthly,
		ExpectedAmount:  1500.0,
		Currency:        "INR",
		BillingDay:      1,
		Status:          models.SubscriptionStatusActive,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := database.CreateSubscription(customSub); err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}

	updatedSummary, err := database.GetSubscriptionsSummary()
	if err != nil {
		t.Fatalf("GetSubscriptionsSummary failed: %v", err)
	}
	if updatedSummary.TotalActive < 4 {
		t.Errorf("Expected at least 4 active subscriptions after manual add, got %d", updatedSummary.TotalActive)
	}
}
