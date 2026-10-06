package models

// InvestmentSnapshot is a complete, dated portfolio, never a bank transaction.
// Providers normalize into this model; original statement fields remain available.
type InvestmentSnapshot struct {
	ID               string              `json:"id"`
	Provider         string              `json:"provider"`
	ParserID         string              `json:"parser_id"`
	AccountRef       string              `json:"account_ref"`
	AsOf             string              `json:"as_of"`
	Currency         string              `json:"currency"`
	Filename         string              `json:"filename"`
	ImportedAt       string              `json:"imported_at"`
	InvestedValue    float64             `json:"invested_value"`
	CurrentValue     float64             `json:"current_value"`
	UnrealizedReturn float64             `json:"unrealized_return"`
	ReturnPercent    *float64            `json:"return_percent"`
	Holdings         []InvestmentHolding `json:"holdings"`
	Sheets           []InvestmentSheet   `json:"sheets"`
	Warnings         []string            `json:"warnings"`
}

type InvestmentHolding struct {
	Symbol           string            `json:"symbol"`
	ISIN             string            `json:"isin"`
	AssetClass       string            `json:"asset_class"`
	Quantity         float64           `json:"quantity"`
	AveragePrice     float64           `json:"average_price"`
	ClosingPrice     float64           `json:"closing_price"`
	InvestedValue    float64           `json:"invested_value"`
	CurrentValue     float64           `json:"current_value"`
	UnrealizedReturn float64           `json:"unrealized_return"`
	ReturnPercent    *float64          `json:"return_percent"`
	Fields           map[string]string `json:"fields"`
}

type InvestmentSheet struct {
	Name string     `json:"name"`
	Rows [][]string `json:"rows"`
}

func InvestmentReturnPercent(cost, gain float64) *float64 {
	if cost <= 0 {
		return nil
	}
	value := gain / cost * 100
	return &value
}
