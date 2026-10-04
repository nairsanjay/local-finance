package service

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"local-finance/internal/db"
	"local-finance/internal/models"
)

type CardRecommendationService struct {
	db *db.DB
}

func NewCardRecommendationService(database *db.DB) *CardRecommendationService {
	return &CardRecommendationService{
		db: database,
	}
}

func (s *CardRecommendationService) RecommendBestCards(req models.CardRecommendationRequest) ([]models.CardRecommendation, error) {
	overview, err := s.db.GetCardPortfolioOverview()
	if err != nil {
		return nil, err
	}

	if len(overview.Cards) == 0 {
		return []models.CardRecommendation{}, nil
	}

	amount := req.Amount
	if amount <= 0 {
		amount = 1000.0 // Default reference spend for comparison
	}

	searchTerm := strings.TrimSpace(strings.ToUpper(req.Merchant))
	categoryTerm := strings.TrimSpace(strings.ToUpper(req.Category))

	var recommendations []models.CardRecommendation

	for _, card := range overview.Cards {
		variant := "Credit Card"
		if card.CardVariant != nil && *card.CardVariant != "" {
			variant = *card.CardVariant
		}
		network := "VISA"
		if card.CardNetwork != nil && *card.CardNetwork != "" {
			network = *card.CardNetwork
		}

		rec := models.CardRecommendation{
			AccountID:        card.ID,
			BankName:         card.BankName,
			CardVariant:      variant,
			CardNetwork:      network,
			CardColor:        card.CardColor,
			RewardType:       card.RewardType,
			InterestFreeDays: card.InterestFreeDaysRemaining,
			RewardRate:       card.BaseRewardRate,
			RewardDescription: fmt.Sprintf("%.1f%% Base reward on general spends", card.BaseRewardRate),
		}

		// Find best matching reward rule
		var matchedRule *models.CardRewardRule
		for _, rule := range card.RewardRules {
			if rule.MerchantPattern == "" {
				continue
			}

			// Check exact or regex match
			pattern := "(?i)(" + rule.MerchantPattern + ")"
			re, err := regexp.Compile(pattern)
			if err == nil {
				if searchTerm != "" && re.MatchString(searchTerm) {
					matchedRule = &rule
					break
				}
				if categoryTerm != "" && re.MatchString(categoryTerm) {
					matchedRule = &rule
					break
				}
				if rule.CategoryName != "" && categoryTerm != "" && strings.Contains(strings.ToUpper(rule.CategoryName), categoryTerm) {
					matchedRule = &rule
					break
				}
			}
		}

		if matchedRule != nil {
			rec.RewardRate = matchedRule.RewardPercentage
			rec.RewardDescription = matchedRule.RewardDescription
			rec.Notes = fmt.Sprintf("Matched category: %s", matchedRule.CategoryName)
		} else {
			rec.Notes = "Standard baseline reward rate"
		}

		// Calculate estimated monetary reward in INR
		if strings.ToUpper(rec.RewardType) == "REWARD_POINTS" {
			// Estimate reward points value (₹0.25 standard in India, ₹0.50 on super-premium travel cards)
			pointValue := 0.25
			varName := strings.ToUpper(variant)
			if strings.Contains(varName, "INFINIA") || strings.Contains(varName, "MAGNUS") || strings.Contains(varName, "DINERS") {
				pointValue = 0.50
			}
			pts := (amount * rec.RewardRate) / 100.0
			rec.EstimatedReward = pts * pointValue
			rec.RewardDescription = fmt.Sprintf("%s (~₹%.2f value @ ₹%.2f/pt)", rec.RewardDescription, rec.EstimatedReward, pointValue)
		} else {
			// Direct cashback INR
			rec.EstimatedReward = (amount * rec.RewardRate) / 100.0
		}

		recommendations = append(recommendations, rec)
	}

	// Sort recommendations: Highest estimated monetary reward first, tie-break with longest interest-free runway
	sort.Slice(recommendations, func(i, j int) bool {
		if recommendations[i].EstimatedReward != recommendations[j].EstimatedReward {
			return recommendations[i].EstimatedReward > recommendations[j].EstimatedReward
		}
		if recommendations[i].RewardRate != recommendations[j].RewardRate {
			return recommendations[i].RewardRate > recommendations[j].RewardRate
		}
		return recommendations[i].InterestFreeDays > recommendations[j].InterestFreeDays
	})

	// Assign Ranks
	for i := range recommendations {
		recommendations[i].Rank = i + 1
	}

	return recommendations, nil
}
