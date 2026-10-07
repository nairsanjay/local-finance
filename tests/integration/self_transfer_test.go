package integration_test

import (
	"strings"
	"testing"

	"local-finance/internal/db"
	"local-finance/internal/models"
	"local-finance/internal/service"
)

func TestReferenceTransfersReconciledOnImportInEitherOrder(t *testing.T) {
	const header = "Date,Narration,Chq/Ref,Value Dt,Withdrawal,Deposit,Closing Balance\n"
	statements := []string{
		header + "01/04/2026,UPIAR/123456789012/DR/PERSON/BANK/person,,01/04/2026,500,,500\n02/04/2026,SHOP,123456789015,02/04/2026,100,,400\n",
		header + "01/04/2026,UPI-PERSON-person@bank-BANK-123456789012-PAYMENT,123456789012,01/04/2026,,500,500\n02/04/2026,EMPLOYER SALARY,123456789016,02/04/2026,,1000,1500\n",
	}
	for _, order := range [][]int{{0, 1}, {1, 0}} {
		database := testDatabase(t)
		var accountIDs []string
		for _, mask := range []string{"XX1001", "XX1002"} {
			a, err := database.GetOrCreateAccount("HDFC Bank", models.AccountTypeSavings, "", mask, "", "", "", "", "", "", nil)
			if err != nil {
				t.Fatal(err)
			}
			accountIDs = append(accountIDs, a.ID)
		}
		svc := service.NewTransactionService(database)
		for attempt := 0; attempt < 2; attempt++ {
			for _, index := range order {
				result, err := svc.ImportStatement("hdfc.csv", strings.NewReader(statements[index]), accountIDs[index], "hdfc_savings_csv_v1", "")
				if err != nil {
					t.Fatal(err)
				}
				if attempt == 1 && (result.InsertedCount != 0 || result.DuplicateCount != 2) {
					t.Fatalf("reupload not idempotent: %+v", result)
				}
			}
			overview, err := database.GetAnalyticsOverview()
			if err != nil {
				t.Fatal(err)
			}
			if overview.TotalExpense != 100 || overview.TotalIncome != 1000 {
				t.Fatalf("import order %v included transfers: %+v", order, overview)
			}
		}
	}
}

func TestSelfTransfersExcludedFromIncomeAndExpense(t *testing.T) {
	database := testDatabase(t)
	svc := service.NewTransactionService(database)
	const statement = `Date,Narration,Chq/Ref,Value Dt,Withdrawal,Deposit,Closing Balance
01/04/2026,NEFT CR-ABC123-EMPLOYER-SALARY,ABC123,01/04/2026,,1000,1000
02/04/2026,UPI-ALEX-alex@okhdfcbank-HDFC-123456789012-SELF TRANSFER,123456789012,02/04/2026,500,,500
03/04/2026,UPI/CR/123456789013/ALEX/HDFC/alex@okhdfcbank/SELF TRANSFER,123456789013,03/04/2026,,500,1000
04/04/2026,UPI-SHOP-shop@okhdfcbank-HDFC-123456789014-PAYMENT,123456789014,04/04/2026,100,,900
`
	for attempt := 0; attempt < 2; attempt++ {
		result, err := svc.ImportStatement("hdfc.csv", strings.NewReader(statement), "", "hdfc_savings_csv_v1", "")
		if err != nil {
			t.Fatal(err)
		}
		if attempt == 0 && result.InsertedCount != 4 || attempt == 1 && (result.InsertedCount != 0 || result.DuplicateCount != 4) {
			t.Fatalf("unexpected import counts: %+v", result)
		}
		overview, err := database.GetAnalyticsOverview()
		if err != nil {
			t.Fatal(err)
		}
		if overview.TotalIncome != 1000 || overview.TotalExpense != 100 || overview.NetSavings != 900 {
			t.Fatalf("self transfers affected totals: %+v", overview)
		}
		if len(overview.MonthlyTrends) != 1 || overview.MonthlyTrends[0].Income != 1000 || overview.MonthlyTrends[0].Expense != 100 {
			t.Fatalf("self transfers affected monthly cash flow: %+v", overview.MonthlyTrends)
		}
		if len(overview.TopPayees) != 1 || overview.TopPayees[0].TotalSpent != 100 {
			t.Fatalf("self transfers affected payee spending: %+v", overview.TopPayees)
		}
		isTransfer := true
		_, count, err := database.ListTransactions(db.TransactionFilter{IsTransfer: &isTransfer, Limit: 10})
		if err != nil || count != 2 {
			t.Fatalf("want both self-transfer rows retained; count=%d, err=%v", count, err)
		}
	}
}
