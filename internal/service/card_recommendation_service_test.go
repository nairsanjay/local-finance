package service

import (
	"path/filepath"
	"testing"
	"time"

	"local-finance/internal/db"
	"local-finance/internal/models"
)

func TestCardRecommendationEngine(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_cards.db")

	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}
	defer database.Close()

	// 1. Create Credit Card Accounts
	credLimit := 500000.0
	swiggyCard, err := database.GetOrCreateAccount("HDFC Bank", models.AccountTypeCreditCard, "", "XX7496", "", "", "", "VISA", "HDFC Swiggy Card", "RAHUL SHARMA", &credLimit)
	if err != nil {
		t.Fatalf("Failed to create swiggy card: %v", err)
	}

	amazonCard, err := database.GetOrCreateAccount("ICICI Bank", models.AccountTypeCreditCard, "", "XX0001", "", "", "", "VISA", "Amazon Pay ICICI", "RAHUL SHARMA", &credLimit)
	if err != nil {
		t.Fatalf("Failed to create amazon card: %v", err)
	}

	axisCard, err := database.GetOrCreateAccount("Axis Bank", models.AccountTypeCreditCard, "", "XX0308", "", "", "", "MASTERCARD", "Flipkart Axis Card", "RAHUL SHARMA", &credLimit)
	if err != nil {
		t.Fatalf("Failed to create axis card: %v", err)
	}

	// 2. Fetch Overview (triggers auto-seeding of Indian default rules)
	overview, err := database.GetCardPortfolioOverview()
	if err != nil {
		t.Fatalf("Failed to get portfolio overview: %v", err)
	}
	if len(overview.Cards) != 3 {
		t.Fatalf("Expected 3 cards in portfolio, got %d", len(overview.Cards))
	}

	// 3. Test Service Recommendation
	recService := NewCardRecommendationService(database)

	// Test Swiggy Query
	swiggyRecs, err := recService.RecommendBestCards(models.CardRecommendationRequest{
		Merchant: "Swiggy",
		Amount:   1000,
	})
	if err != nil {
		t.Fatalf("RecommendBestCards failed: %v", err)
	}
	if len(swiggyRecs) == 0 {
		t.Fatalf("Expected recommendations for Swiggy, got 0")
	}
	if swiggyRecs[0].AccountID != swiggyCard.ID {
		t.Errorf("Expected HDFC Swiggy card to rank #1 for Swiggy, got %s (%s)", swiggyRecs[0].CardVariant, swiggyRecs[0].BankName)
	}
	if swiggyRecs[0].RewardRate != 10.0 {
		t.Errorf("Expected 10%% reward rate for Swiggy on Swiggy Card, got %.1f%%", swiggyRecs[0].RewardRate)
	}
	if swiggyRecs[0].EstimatedReward != 100.0 {
		t.Errorf("Expected ₹100 cashback on ₹1000 spend, got %.2f", swiggyRecs[0].EstimatedReward)
	}

	// Test Amazon Query
	amazonRecs, err := recService.RecommendBestCards(models.CardRecommendationRequest{
		Merchant: "Amazon.in Shopping",
		Amount:   2000,
	})
	if err != nil {
		t.Fatalf("RecommendBestCards for Amazon failed: %v", err)
	}
	if len(amazonRecs) == 0 {
		t.Fatalf("Expected recommendations for Amazon, got 0")
	}
	if amazonRecs[0].RewardRate != 5.0 {
		t.Errorf("Expected top reward rate for Amazon to be 5%%, got %.1f%%", amazonRecs[0].RewardRate)
	}

	// Test Flipkart Query
	flipkartRecs, err := recService.RecommendBestCards(models.CardRecommendationRequest{
		Merchant: "Flipkart",
		Amount:   5000,
	})
	if err != nil {
		t.Fatalf("RecommendBestCards for Flipkart failed: %v", err)
	}
	if flipkartRecs[0].AccountID != axisCard.ID && flipkartRecs[0].AccountID != swiggyCard.ID {
		t.Errorf("Expected Axis Flipkart or Swiggy Card to rank top for Flipkart, got %s", flipkartRecs[0].CardVariant)
	}

	// 4. Test Card Metadata Update
	newFee := 500.0
	newWaiver := 200000.0
	bDay := 15
	err = database.UpdateCardMetadata(swiggyCard.ID, models.UpdateCardMetadataRequest{
		AnnualFee:          &newFee,
		FeeWaiverThreshold: &newWaiver,
		BillingDay:         &bDay,
	})
	if err != nil {
		t.Fatalf("UpdateCardMetadata failed: %v", err)
	}

	// Verify updated portfolio metrics
	updatedOverview, err := database.GetCardPortfolioOverview()
	if err != nil {
		t.Fatalf("Failed to fetch updated portfolio: %v", err)
	}
	var updatedSwiggy *models.CardDetails
	for _, c := range updatedOverview.Cards {
		if c.ID == swiggyCard.ID {
			updatedSwiggy = &c
			break
		}
	}
	if updatedSwiggy == nil {
		t.Fatalf("Could not find swiggy card in updated overview")
	}
	if updatedSwiggy.AnnualFee != 500.0 {
		t.Errorf("Expected annual fee 500, got %.2f", updatedSwiggy.AnnualFee)
	}
	if updatedSwiggy.FeeWaiverThreshold != 200000.0 {
		t.Errorf("Expected waiver threshold 200000, got %.2f", updatedSwiggy.FeeWaiverThreshold)
	}

	_ = amazonCard
	_ = time.Now()
}
