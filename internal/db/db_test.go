package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"local-finance/internal/models"
)

func TestNewDBWithGooseMigrations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_local_finance.db")

	database, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}
	defer database.Close()

	// Verify categories seeded
	categories, err := database.ListCategories()
	if err != nil {
		t.Fatalf("Failed to list categories: %v", err)
	}
	if len(categories) == 0 {
		t.Errorf("Expected default categories to be seeded, got 0")
	}

	// Verify rules seeded
	rules, err := database.ListRules()
	if err != nil {
		t.Fatalf("Failed to list rules: %v", err)
	}
	if len(rules) == 0 {
		t.Errorf("Expected default rules to be seeded, got 0")
	}

	// Verify accounts table exists and functions
	accounts, err := database.ListAccounts()
	if err != nil {
		t.Fatalf("Failed to list accounts: %v", err)
	}
	if len(accounts) != 0 {
		t.Errorf("Expected 0 initial accounts, got %d", len(accounts))
	}

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file does not exist at %s", dbPath)
	}

	// Test ResetDatabase
	if err := database.ResetDatabase(); err != nil {
		t.Fatalf("ResetDatabase failed: %v", err)
	}

	// Test GetDatabaseInfo
	info, err := database.GetDatabaseInfo()
	if err != nil {
		t.Fatalf("GetDatabaseInfo failed: %v", err)
	}
	if info.Path == "" {
		t.Errorf("Expected non-empty DB path")
	}

	// Test BackupTo
	backupPath := filepath.Join(tempDir, "backup.db")
	if err := database.BackupTo(backupPath); err != nil {
		t.Fatalf("BackupTo failed: %v", err)
	}
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatalf("Backup file was not created at %s", backupPath)
	}

	// Test RestoreFrom
	backupFile, err := os.Open(backupPath)
	if err != nil {
		t.Fatalf("Failed to open backup file: %v", err)
	}
	defer backupFile.Close()

	if err := database.RestoreFrom(backupFile); err != nil {
		t.Fatalf("RestoreFrom failed: %v", err)
	}

	// Test ExportAllDataJSON
	jsonData, err := database.ExportAllDataJSON()
	if err != nil {
		t.Fatalf("ExportAllDataJSON failed: %v", err)
	}
	if len(jsonData.Categories) == 0 {
		t.Errorf("Expected categories in exported JSON")
	}

	// Test ExportTransactionsCSV
	tempCSV := filepath.Join(tempDir, "export.csv")
	csvFile, err := os.Create(tempCSV)
	if err != nil {
		t.Fatalf("Failed to create temp csv: %v", err)
	}
	if err := database.ExportTransactionsCSV(csvFile); err != nil {
		csvFile.Close()
		t.Fatalf("ExportTransactionsCSV failed: %v", err)
	}
	csvFile.Close()

	// Test Category CRUD
	customCat := &models.Category{
		Name:     "Gaming & Tech",
		ColorHex: "#8B5CF6",
		Icon:     "gamepad",
		IsSystem: false,
	}
	if err := database.CreateCategory(customCat); err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}
	if customCat.ID == "" {
		t.Errorf("Expected category ID to be assigned")
	}

	// Test Rule CRUD
	customRule := &models.CategorizationRule{
		Priority:         90,
		MatchField:       "cleaned_payee",
		MatchType:        "CONTAINS",
		MatchPattern:     "STEAM",
		TargetCategoryID: customCat.ID,
		AssignTags:       "gaming,software",
		IsActive:         true,
	}
	if err := database.CreateRule(customRule); err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Test UpdateRule
	customRule.MatchPattern = "STEAM GAMES"
	if err := database.UpdateRule(customRule); err != nil {
		t.Fatalf("Failed to update rule: %v", err)
	}

	// Test ReapplyRules
	updatedCount, err := database.ReapplyRules()
	if err != nil {
		t.Fatalf("Failed to reapply rules: %v", err)
	}
	t.Logf("ReapplyRules returned %d updated transactions", updatedCount)

	// Test DeleteRule
	if err := database.DeleteRule(customRule.ID); err != nil {
		t.Fatalf("Failed to delete rule: %v", err)
	}

	// Test DeleteCategory
	if err := database.DeleteCategory(customCat.ID); err != nil {
		t.Fatalf("Failed to delete category: %v", err)
	}

	// ----------------------------------------------------
	// Test Category Budgeting & Overspend Warning System
	// ----------------------------------------------------
	allCats, err := database.ListCategories()
	if err != nil || len(allCats) == 0 {
		t.Fatalf("Failed to list categories: %v", err)
	}

	var catFood *models.Category
	for _, c := range allCats {
		if c.Name == "Food & Dining" {
			catFood = &c
			break
		}
	}
	if catFood == nil {
		catFood = &allCats[0]
	}

	currentMonth := "2026-08"

	// 1. Set budget target of ₹10,000 for Food & Dining
	if err := database.UpsertCategoryBudget(catFood.ID, currentMonth, 10000.0); err != nil {
		t.Fatalf("UpsertCategoryBudget failed: %v", err)
	}

	// 2. Fetch Budget Summary
	summary, err := database.GetCategoryBudgetSummary(currentMonth)
	if err != nil {
		t.Fatalf("GetCategoryBudgetSummary failed: %v", err)
	}
	if summary.CategoriesWithBudgets == 0 {
		t.Errorf("Expected at least 1 category with budget, got %d", summary.CategoriesWithBudgets)
	}
	if summary.TotalBudget != 10000.0 {
		t.Errorf("Expected TotalBudget ₹10,000, got %.2f", summary.TotalBudget)
	}

	var foodBudget *models.CategoryBudget
	for _, b := range summary.Budgets {
		if b.CategoryID == catFood.ID {
			foodBudget = &b
			break
		}
	}
	if foodBudget == nil {
		t.Fatalf("Food & Dining budget not found in summary")
	}
	if foodBudget.MonthlyLimit != 10000.0 {
		t.Errorf("Expected MonthlyLimit 10000, got %.2f", foodBudget.MonthlyLimit)
	}
	if foodBudget.Status != "SAFE" {
		t.Errorf("Expected initial status SAFE, got %s", foodBudget.Status)
	}

	// 3. Delete Budget
	if err := database.DeleteCategoryBudget(catFood.ID, currentMonth); err != nil {
		t.Fatalf("DeleteCategoryBudget failed: %v", err)
	}
	summaryAfterDel, err := database.GetCategoryBudgetSummary(currentMonth)
	if err != nil {
		t.Fatalf("GetCategoryBudgetSummary after delete failed: %v", err)
	}
	if summaryAfterDel.CategoriesWithBudgets != 0 {
		t.Errorf("Expected 0 categories with budget after delete, got %d", summaryAfterDel.CategoriesWithBudgets)
	}
}

func TestGetCashFlowIntelligence(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_cashflow.db")

	database, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}
	defer database.Close()

	// 1. Create a Savings Account and a Credit Card Account
	savingsAcc, err := database.GetOrCreateAccount("HDFC Bank", models.AccountTypeSavings, "", "XXXX1234", "", "", "", "", "", "RAHUL SHARMA", nil)
	if err != nil {
		t.Fatalf("Failed to create savings account: %v", err)
	}
	cardAcc, err := database.GetOrCreateAccount("HDFC Bank", models.AccountTypeCreditCard, "", "XXXX5678", "", "", "", "VISA", "Regalia", "RAHUL SHARMA", nil)
	if err != nil {
		t.Fatalf("Failed to create card account: %v", err)
	}

	// 2. Insert Inflows & Outflows for Feb 2026 and Mar 2026
	foodCat := "cat_food"
	shoppingCat := "cat_shopping"

	// Feb 2026 (Month 1):
	// Income: ₹80,000 Salary
	// Expense: ₹5,000 Food (Bank), ₹10,000 Shopping (Card) -> Total Outflow ₹15,000
	febTxs := []models.Transaction{
		{
			AccountID: savingsAcc.ID, TxHash: "hash_feb_sal",
			TxDate: "2026-02-01", RawNarration: "SALARY CREDITED INFOSYS", CleanedPayee: "Infosys Salary",
			PaymentMode: models.PaymentModeSalary, TxType: models.TxTypeCredit, Amount: 80000.0,
		},
		{
			AccountID: savingsAcc.ID, TxHash: "hash_feb_food",
			TxDate: "2026-02-05", RawNarration: "UPI/DR/SWIGGY/BANGALORE", CleanedPayee: "Swiggy",
			PaymentMode: models.PaymentModeUPI, TxType: models.TxTypeDebit, Amount: 5000.0, CategoryID: &foodCat,
		},
		{
			AccountID: cardAcc.ID, TxHash: "hash_feb_shop",
			TxDate: "2026-02-10", RawNarration: "AMAZON RETAIL MUMBAI", CleanedPayee: "Amazon",
			PaymentMode: models.PaymentModeCardOnline, TxType: models.TxTypeDebit, Amount: 10000.0, CategoryID: &shoppingCat,
		},
	}
	for i := range febTxs {
		_, err := database.UpsertTransaction(&febTxs[i])
		if err != nil {
			t.Fatalf("Insert feb transaction failed: %v", err)
		}
	}

	// Mar 2026 (Month 2):
	// Income: ₹90,000 Salary + ₹2,000 Refund -> Total Inflow ₹92,000
	// Expense: ₹12,000 Food (Bank, Spiked by 140%), ₹2,000 Shopping (Card, Dropped by 80%) -> Total Outflow ₹14,000
	marTxs := []models.Transaction{
		{
			AccountID: savingsAcc.ID, TxHash: "hash_mar_sal",
			TxDate: "2026-03-01", RawNarration: "SALARY FOR FEB INFOSYS", CleanedPayee: "Infosys Salary",
			PaymentMode: models.PaymentModeSalary, TxType: models.TxTypeCredit, Amount: 90000.0,
		},
		{
			AccountID: savingsAcc.ID, TxHash: "hash_mar_ref",
			TxDate: "2026-03-03", RawNarration: "SWIGGY REFUND TX1234", CleanedPayee: "Swiggy Refund",
			PaymentMode: models.PaymentModeUPI, TxType: models.TxTypeCredit, Amount: 2000.0,
		},
		{
			AccountID: savingsAcc.ID, TxHash: "hash_mar_food",
			TxDate: "2026-03-08", RawNarration: "UPI/DR/SWIGGY/BANGALORE", CleanedPayee: "Swiggy",
			PaymentMode: models.PaymentModeUPI, TxType: models.TxTypeDebit, Amount: 12000.0, CategoryID: &foodCat,
		},
		{
			AccountID: cardAcc.ID, TxHash: "hash_mar_shop",
			TxDate: "2026-03-12", RawNarration: "AMAZON RETAIL MUMBAI", CleanedPayee: "Amazon",
			PaymentMode: models.PaymentModeCardOnline, TxType: models.TxTypeDebit, Amount: 2000.0, CategoryID: &shoppingCat,
		},
	}
	for i := range marTxs {
		_, err := database.UpsertTransaction(&marTxs[i])
		if err != nil {
			t.Fatalf("Insert mar transaction failed: %v", err)
		}
	}

	// 3. Test GetCashFlowIntelligence for March 2026
	res, err := database.GetCashFlowIntelligence("2026-03")
	if err != nil {
		t.Fatalf("GetCashFlowIntelligence failed: %v", err)
	}

	if res.Period != "2026-03" {
		t.Errorf("Expected period 2026-03, got %s", res.Period)
	}
	if res.PreviousMonth != "2026-02" {
		t.Errorf("Expected previous month 2026-02, got %s", res.PreviousMonth)
	}
	if len(res.AvailableMonths) != 2 {
		t.Errorf("Expected 2 available months, got %d", len(res.AvailableMonths))
	}
	if res.Summary.TotalInflow != 92000.0 {
		t.Errorf("Expected TotalInflow 92,000, got %.2f", res.Summary.TotalInflow)
	}
	if res.Summary.TotalOutflow != 14000.0 {
		t.Errorf("Expected TotalOutflow 14,000, got %.2f", res.Summary.TotalOutflow)
	}
	if res.Summary.NetSurplus != 78000.0 {
		t.Errorf("Expected NetSurplus 78,000, got %.2f", res.Summary.NetSurplus)
	}

	// 4. Verify Sankey Nodes and Links
	if len(res.Sankey.Nodes) == 0 {
		t.Errorf("Expected Sankey nodes, got 0")
	}
	if len(res.Sankey.Links) == 0 {
		t.Errorf("Expected Sankey links, got 0")
	}

	// Verify Surplus Node exists in Sankey
	hasSurplusNode := false
	for _, n := range res.Sankey.Nodes {
		if n.Type == models.SankeyNodeSurplus {
			hasSurplusNode = true
			if n.TotalValue != 78000.0 {
				t.Errorf("Expected surplus node value 78,000, got %.2f", n.TotalValue)
			}
		}
	}
	if !hasSurplusNode {
		t.Errorf("Expected surplus node in Sankey")
	}

	// 5. Verify MoM Anomalies (Food Surge + Shopping Drop + High Savings)
	hasFoodSpike := false
	hasShopDrop := false
	hasHighSavings := false

	for _, a := range res.Anomalies {
		if a.Type == "SPIKE" && a.CategoryName == "Food & Dining" {
			hasFoodSpike = true
			if a.DeltaAmount != 7000.0 {
				t.Errorf("Expected food spike delta 7000, got %.2f", a.DeltaAmount)
			}
		}
		if a.Type == "DROP" && a.CategoryName == "Shopping & E-Commerce" {
			hasShopDrop = true
		}
		if a.Type == "SAVINGS_MILESTONE" {
			hasHighSavings = true
		}
	}

	if !hasFoodSpike {
		t.Errorf("Expected food spike anomaly to be detected")
	}
	if !hasShopDrop {
		t.Errorf("Expected shopping drop anomaly to be detected")
	}
	if !hasHighSavings {
		t.Errorf("Expected high savings milestone anomaly to be detected")
	}
}

func TestGetSalaryInsights(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_salary.db")

	database, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to init DB: %v", err)
	}
	defer database.Close()

	// 1. Create account
	acc, err := database.GetOrCreateAccount("HDFC Bank", models.AccountTypeSavings, "", "XXXX1234", "", "", "", "", "", "TEST USER", nil)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	catSalary := "cat_salary"

	// 2. Insert salary transactions
	txs := []models.Transaction{
		{
			ID:           "tx_s1",
			AccountID:    acc.ID,
			TxHash:       "hash_s1",
			TxDate:       "2023-01-31",
			RawNarration: "SALARY CREDIT ACME CORP",
			CleanedPayee: "ACME Corp",
			PaymentMode:  models.PaymentModeSalary,
			TxType:       models.TxTypeCredit,
			Amount:       50000.0,
			CategoryID:   &catSalary,
		},
		{
			ID:           "tx_s2",
			AccountID:    acc.ID,
			TxHash:       "hash_s2",
			TxDate:       "2023-02-28",
			RawNarration: "SALARY CREDIT ACME CORP",
			CleanedPayee: "ACME Corp",
			PaymentMode:  models.PaymentModeSalary,
			TxType:       models.TxTypeCredit,
			Amount:       50000.0,
			CategoryID:   &catSalary,
		},
		{
			ID:           "tx_s3",
			AccountID:    acc.ID,
			TxHash:       "hash_s3",
			TxDate:       "2023-03-31",
			RawNarration: "SALARY CREDIT ACME CORP",
			CleanedPayee: "ACME Corp",
			PaymentMode:  models.PaymentModeSalary,
			TxType:       models.TxTypeCredit,
			Amount:       60000.0, // Hike 20%
			CategoryID:   &catSalary,
		},
		{
			ID:           "tx_s4",
			AccountID:    acc.ID,
			TxHash:       "hash_s4",
			TxDate:       "2024-01-31",
			RawNarration: "PAYROLL GLOBEX LTD",
			CleanedPayee: "Globex Ltd",
			PaymentMode:  models.PaymentModeSalary,
			TxType:       models.TxTypeCredit,
			Amount:       90000.0,
			CategoryID:   &catSalary,
		},
	}

	for _, tx := range txs {
		if _, err := database.UpsertTransaction(&tx); err != nil {
			t.Fatalf("Failed to insert tx %s: %v", tx.ID, err)
		}
	}

	// 3. Query salary insights
	insights, err := database.GetSalaryInsights()
	if err != nil {
		t.Fatalf("GetSalaryInsights failed: %v", err)
	}

	if insights.TotalPaychecks != 4 {
		t.Errorf("Expected 4 paychecks, got %d", insights.TotalPaychecks)
	}
	expectedLifetime := 50000.0 + 50000.0 + 60000.0 + 90000.0
	if insights.LifetimeEarned != expectedLifetime {
		t.Errorf("Expected lifetime earned %.2f, got %.2f", expectedLifetime, insights.LifetimeEarned)
	}
	if insights.FirstSalaryAmount != 50000.0 {
		t.Errorf("Expected first salary 50000, got %.2f", insights.FirstSalaryAmount)
	}
	if insights.LatestSalaryAmount != 90000.0 {
		t.Errorf("Expected latest salary 90000, got %.2f", insights.LatestSalaryAmount)
	}
	if insights.PeakSalaryAmount != 90000.0 {
		t.Errorf("Expected peak salary 90000, got %.2f", insights.PeakSalaryAmount)
	}
	if insights.PeakEmployer != "Globex Ltd" {
		t.Errorf("Expected peak employer Globex Ltd, got %s", insights.PeakEmployer)
	}
	if len(insights.YearlyProgress) != 2 {
		t.Errorf("Expected 2 yearly records (2023, 2024), got %d", len(insights.YearlyProgress))
	}
	if len(insights.Employers) != 2 {
		t.Errorf("Expected 2 employers, got %d", len(insights.Employers))
	}
	if len(insights.RecentPaychecks) != 4 {
		t.Errorf("Expected 4 recent paychecks, got %d", len(insights.RecentPaychecks))
	}
}

func TestRuleExceptionsAndTxType(t *testing.T) {
	// 1. Unit test MatchesTxType and MatchesException directly
	rule := &models.CategorizationRule{
		TxType:         "CREDIT",
		ExcludePattern: "maid, driver, cook, helper, advance",
	}

	if !rule.MatchesTxType("CREDIT") {
		t.Errorf("Expected rule to match CREDIT")
	}
	if rule.MatchesTxType("DEBIT") {
		t.Errorf("Expected rule to NOT match DEBIT")
	}

	allRule := &models.CategorizationRule{TxType: "ALL"}
	if !allRule.MatchesTxType("CREDIT") || !allRule.MatchesTxType("DEBIT") {
		t.Errorf("Expected ALL rule to match both DEBIT and CREDIT")
	}

	if !rule.MatchesException("UPI-MAID SALARY-123", "") {
		t.Errorf("Expected narration containing 'MAID' to match exception")
	}
	if !rule.MatchesException("", "DRIVER RAMESH") {
		t.Errorf("Expected payee containing 'DRIVER' to match exception")
	}
	if rule.MatchesException("SALARY CREDIT TECH CORP", "TECH CORP") {
		t.Errorf("Expected standard salary to NOT match exception")
	}

	// 2. Integration test with DB and ReapplyRules
	testDBPath := filepath.Join(os.TempDir(), "test_rule_exceptions.db")
	_ = os.Remove(testDBPath)
	defer os.Remove(testDBPath)

	database, err := NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize test DB: %v", err)
	}
	defer database.Close()

	// Verify default salary rule seeded correctly
	rules, err := database.ListRules()
	if err != nil {
		t.Fatalf("Failed to list rules: %v", err)
	}
	var salaryRule *models.CategorizationRule
	for i := range rules {
		if rules[i].ID == "rule_salary" {
			salaryRule = &rules[i]
			break
		}
	}
	if salaryRule == nil {
		t.Fatalf("Expected default rule_salary to be seeded")
	}
	if salaryRule.TxType != "CREDIT" {
		t.Errorf("Expected rule_salary TxType to be CREDIT, got %s", salaryRule.TxType)
	}
	if !strings.Contains(salaryRule.ExcludePattern, "maid") {
		t.Errorf("Expected rule_salary ExcludePattern to contain 'maid', got %s", salaryRule.ExcludePattern)
	}

	// Create test account
	_, err = database.conn.Exec(`
		INSERT INTO accounts (id, bank_name, account_type, currency)
		VALUES ('acc_test_rule', 'HDFC Bank', 'SAVINGS', 'INR')
	`)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Insert test transactions
	txList := []struct {
		id        string
		narration string
		payee     string
		txType    models.TxType
	}{
		{"tx_sal_legit", "ACH SALARY CREDIT TECH CORP", "TECH CORP", models.TxTypeCredit},
		{"tx_sal_maid", "UPI-MAID SALARY-MONTHLY", "SHANTI MAID", models.TxTypeDebit},
		{"tx_sal_reimb", "REIMBURSEMENT MAID SALARY EXP", "MAID REIMBURSE", models.TxTypeCredit},
		{"tx_sal_legit_2", "NEFT SALARY CREDIT ACME LTD", "ACME LTD", models.TxTypeCredit},
	}

	for _, item := range txList {
		// Pre-populate with 'cat_salary' to simulate previously imported miscategorized records
		_, err := database.conn.Exec(`
			INSERT INTO transactions (id, account_id, tx_hash, tx_date, raw_narration, cleaned_payee, payment_mode, reference_number, tx_type, amount, category_id)
			VALUES (?, 'acc_test_rule', ?, '2026-03-01', ?, ?, 'UPI', 'REF123', ?, 5000.0, 'cat_salary')
		`, item.id, item.id+"_hash", item.narration, item.payee, item.txType)
		if err != nil {
			t.Fatalf("Failed to insert test transaction: %v", err)
		}
	}

	// Reapply rules
	updated, err := database.ReapplyRules()
	if err != nil {
		t.Fatalf("Failed to reapply rules: %v", err)
	}
	t.Logf("ReapplyRules updated %d transactions", updated)

	// Verify categories of transactions
	checkCategory := func(txID string) string {
		var catID sql.NullString
		_ = database.conn.QueryRow("SELECT category_id FROM transactions WHERE id = ?", txID).Scan(&catID)
		if catID.Valid {
			return catID.String
		}
		return ""
	}

	if checkCategory("tx_sal_legit") != "cat_salary" {
		t.Errorf("Expected tx_sal_legit to be cat_salary, got '%s'", checkCategory("tx_sal_legit"))
	}
	if checkCategory("tx_sal_maid") != "cat_others" {
		t.Errorf("Expected tx_sal_maid (DEBIT) to be reset to cat_others, got '%s'", checkCategory("tx_sal_maid"))
	}
	if checkCategory("tx_sal_reimb") != "cat_others" {
		t.Errorf("Expected tx_sal_reimb (contains exception 'maid') to be reset to cat_others, got '%s'", checkCategory("tx_sal_reimb"))
	}
	if checkCategory("tx_sal_legit_2") != "cat_salary" {
		t.Errorf("Expected tx_sal_legit_2 to be cat_salary, got '%s'", checkCategory("tx_sal_legit_2"))
	}
}

func TestManualCategoryEditingAndReapplyRules(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testdb_manual_cat_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer database.Close()

	// Insert account
	_, err = database.conn.Exec(`
		INSERT INTO accounts (id, bank_name, account_type, currency)
		VALUES ('acc_manual_test', 'HDFC Bank', 'SAVINGS', 'INR')
	`)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// 1. Insert a transaction matching Amazon rule (default category would be cat_shopping)
	tx := &models.Transaction{
		AccountID:    "acc_manual_test",
		TxHash:       "hash_amazon_manual_test",
		TxDate:       "2026-03-01",
		RawNarration: "AMAZON RETAIL MUMBAI",
		CleanedPayee: "AMAZON",
		PaymentMode:  "CARD_ONLINE",
		TxType:       models.TxTypeDebit,
		Amount:       2499.0,
	}
	isNew, err := database.UpsertTransaction(tx)
	if err != nil || !isNew {
		t.Fatalf("Failed to insert transaction: %v", err)
	}

	// 2. Update category manually to cat_entertainment
	newCat := "cat_entertainment"
	updatedTx, err := database.UpdateTransaction(tx.ID, models.UpdateTransactionRequest{
		CategoryID: &newCat,
	})
	if err != nil {
		t.Fatalf("Failed to update transaction category: %v", err)
	}
	if updatedTx.CategoryID == nil || *updatedTx.CategoryID != "cat_entertainment" {
		t.Fatalf("Expected CategoryID to be cat_entertainment, got %v", updatedTx.CategoryID)
	}
	if !updatedTx.IsManualCategory {
		t.Fatalf("Expected IsManualCategory to be true, got %v", updatedTx.IsManualCategory)
	}

	// 3. Run ReapplyRules - should NOT overwrite because is_manual_category is 1
	reappliedCount, err := database.ReapplyRules()
	if err != nil {
		t.Fatalf("Failed to reapply rules: %v", err)
	}
	t.Logf("Reapplied count: %d", reappliedCount)

	fetchedTx, err := database.GetTransaction(tx.ID)
	if err != nil {
		t.Fatalf("Failed to get transaction: %v", err)
	}
	if fetchedTx.CategoryID == nil || *fetchedTx.CategoryID != "cat_entertainment" {
		t.Errorf("ReapplyRules overwrote manual category! Got: %v, expected cat_entertainment", fetchedTx.CategoryID)
	}
	if !fetchedTx.IsManualCategory {
		t.Errorf("Expected IsManualCategory to remain true")
	}

	// 4. Re-import statement (UpsertTransaction with same tx_hash) - should preserve manual category
	reimportTx := &models.Transaction{
		AccountID:    "acc_manual_test",
		TxHash:       "hash_amazon_manual_test",
		TxDate:       "2026-03-01",
		RawNarration: "AMAZON RETAIL MUMBAI",
		CleanedPayee: "AMAZON",
		PaymentMode:  "CARD_ONLINE",
		TxType:       models.TxTypeDebit,
		Amount:       2499.0,
	}
	isNew2, err := database.UpsertTransaction(reimportTx)
	if err != nil || isNew2 {
		t.Fatalf("Expected existing transaction on reimport, got isNew=%v, err=%v", isNew2, err)
	}

	fetchedTx2, err := database.GetTransaction(tx.ID)
	if err != nil {
		t.Fatalf("Failed to get transaction: %v", err)
	}
	if fetchedTx2.CategoryID == nil || *fetchedTx2.CategoryID != "cat_entertainment" {
		t.Errorf("Reimport overwrote manual category! Got: %v, expected cat_entertainment", fetchedTx2.CategoryID)
	}
	if !fetchedTx2.IsManualCategory {
		t.Errorf("Expected IsManualCategory to remain true after reimport")
	}
}
