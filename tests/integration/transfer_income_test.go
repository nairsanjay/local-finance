package integration_test

import (
	"testing"

	"local-finance/internal/db"
	"local-finance/internal/models"
)

func TestTransfersAreNeitherIncomeNorExpense(t *testing.T) {
	for _, tc := range []struct {
		name, category string
		flag           bool
	}{
		{"category only", "cat_transfers", false},
		{"transfer flag only", "cat_salary", true},
		{"category and flag", "cat_transfers", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			database := testDatabase(t)
			a, err := database.GetOrCreateAccount("Test Bank A", models.AccountTypeSavings, "", "XX1001", "", "", "", "", "", "", nil)
			if err != nil {
				t.Fatal(err)
			}
			b, err := database.GetOrCreateAccount("Test Bank B", models.AccountTypeSavings, "", "XX1002", "", "", "", "", "", "", nil)
			if err != nil {
				t.Fatal(err)
			}
			salaryCategory := "cat_salary"
			for _, tx := range []*models.Transaction{
				{AccountID: a.ID, TxHash: "salary", TxType: models.TxTypeCredit, Amount: 1000, CategoryID: &salaryCategory, CleanedPayee: "Employer Salary"},
				{AccountID: a.ID, TxHash: "shop", TxType: models.TxTypeDebit, Amount: 100, CleanedPayee: "Shop"},
				{AccountID: a.ID, TxHash: "transfer-credit", TxType: models.TxTypeCredit, Amount: 500, CategoryID: &tc.category, IsTransfer: tc.flag, CleanedPayee: "Payroll transfer", Notes: "keep note"},
				{AccountID: b.ID, TxHash: "transfer-debit", TxType: models.TxTypeDebit, Amount: 500, CategoryID: &tc.category, IsTransfer: tc.flag},
			} {
				tx.TxDate = "2026-04-01"
				if _, err := database.UpsertTransaction(tx); err != nil {
					t.Fatal(err)
				}
			}
			overview, err := database.GetAnalyticsOverview()
			if err != nil {
				t.Fatal(err)
			}
			if overview.TotalIncome != 1000 || overview.TotalExpense != 100 || len(overview.MonthlyTrends) != 1 || overview.MonthlyTrends[0].Income != 1000 {
				t.Fatalf("transfer in overview: %+v", overview)
			}
			flow, err := database.GetCashFlowIntelligence("2026-04")
			if err != nil || flow.Summary.TotalInflow != 1000 || flow.Summary.TotalOutflow != 100 {
				t.Fatalf("transfer in cash flow: %+v %v", flow, err)
			}
			wrapped, err := database.GetWrappedStory("2026")
			if err != nil || wrapped.TotalIncome != 1000 || wrapped.TotalExpense != 100 {
				t.Fatalf("transfer in year totals: %+v %v", wrapped, err)
			}
			salary, err := database.GetSalaryInsights()
			if err != nil || salary.TotalPaychecks != 1 || salary.LifetimeEarned != 1000 {
				t.Fatalf("transfer in salary: %+v %v", salary, err)
			}
			profile, err := database.GetMerchantProfile("Payroll transfer")
			if err != nil || profile.TotalCredits != 0 {
				t.Fatalf("transfer in merchant credits: %+v %v", profile, err)
			}
			rows, count, err := database.ListTransactions(db.TransactionFilter{Limit: 10})
			if err != nil || count != 4 {
				t.Fatalf("ledger lost rows: %d %v", count, err)
			}
			for _, row := range rows {
				if row.TxHash == "transfer-credit" && (row.IsTransfer != tc.flag || row.Notes != "keep note" || row.CategoryID == nil || *row.CategoryID != tc.category) {
					t.Fatal("ledger mutated")
				}
			}
			if err := database.RecalculateAccountBalance(a.ID); err != nil {
				t.Fatal(err)
			}
			accounts, err := database.ListAccounts()
			if err != nil {
				t.Fatal(err)
			}
			for _, account := range accounts {
				if account.ID == a.ID && account.CurrentBalance != 1400 {
					t.Fatalf("transfer removed from bank balance: %v", account.CurrentBalance)
				}
			}
		})
	}
}
