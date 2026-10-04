package models

// ReviewPeriod uses inclusive dates; current months compare equivalent elapsed days.
type ReviewPeriod struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type ReviewCoverage struct {
	AccountID     string `json:"account_id"`
	Name          string `json:"name"`
	LatestEnd     string `json:"latest_end"`
	CurrentDays   int    `json:"current_days"`
	PreviousDays  int    `json:"previous_days"`
	CurrentTotal  int    `json:"current_total"`
	PreviousTotal int    `json:"previous_total"`
}

type ReviewSpend struct {
	Amount float64 `json:"amount"`
	Count  int     `json:"count"`
}

type ReviewMerchant struct {
	Name     string      `json:"name"`
	Current  ReviewSpend `json:"current"`
	Previous ReviewSpend `json:"previous"`
	Delta    float64     `json:"delta"`
}

type ReviewCategory struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Current   ReviewSpend      `json:"current"`
	Previous  ReviewSpend      `json:"previous"`
	Delta     float64          `json:"delta"`
	Merchants []ReviewMerchant `json:"merchants"`
}

type MonthlyReview struct {
	Month            string           `json:"month"`
	NextMonth        string           `json:"next_month"`
	AvailableMonths  []string         `json:"available_months"`
	CurrentPeriod    ReviewPeriod     `json:"current_period"`
	PreviousPeriod   ReviewPeriod     `json:"previous_period"`
	IsPartialMonth   bool             `json:"is_partial_month"`
	CoverageComplete bool             `json:"coverage_complete"`
	Coverage         []ReviewCoverage `json:"coverage"`
	Current          ReviewSpend      `json:"current"`
	Previous         ReviewSpend      `json:"previous"`
	Delta            float64          `json:"delta"`
	Categories       []ReviewCategory `json:"categories"`
}

type ReviewTransaction struct {
	ID      string  `json:"id"`
	Date    string  `json:"date"`
	Payee   string  `json:"payee"`
	Account string  `json:"account"`
	Amount  float64 `json:"amount"`
}

type ReviewEvidence struct {
	Items    []ReviewTransaction `json:"items"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	Period   ReviewPeriod        `json:"period"`
}
