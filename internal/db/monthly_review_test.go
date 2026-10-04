package db_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"local-finance/internal/db"
	"local-finance/internal/models"
)

func reviewFixture(t *testing.T) (*db.DB, string) {
	t.Helper()
	database, err := db.NewDB(filepath.Join(t.TempDir(), "review.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	account, err := database.GetOrCreateAccount("Review Bank", models.AccountTypeSavings, "", "1234", "", "", "", "", "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	return database, account.ID
}

func reviewStatement(t *testing.T, database *db.DB, account, id, from, to string) {
	t.Helper()
	if err := database.CreateStatementImport(&models.StatementImport{ID: id, AccountID: account,
		Filename: id, FileHash: id, StatementFormat: "CSV", ParserUsed: "test", StartDate: &from, EndDate: &to,
		ImportedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
}

func reviewTransaction(t *testing.T, database *db.DB, account, hash, date, category, payee string, amount float64, kind models.TxType, transfer, excluded bool) *models.Transaction {
	t.Helper()
	transaction := &models.Transaction{AccountID: account, TxHash: hash, TxDate: date, RawNarration: payee,
		CleanedPayee: payee, Amount: amount, TxType: kind, IsTransfer: transfer, IsExcluded: excluded,
		Notes: "Keep my note", Tags: "personal", IsManualCategory: true}
	if category != "" {
		transaction.CategoryID = &category
	}
	if _, err := database.UpsertTransaction(transaction); err != nil {
		t.Fatal(err)
	}
	return transaction
}

func TestMonthlyReviewExplainsSpendingWithoutChangingLedger(t *testing.T) {
	database, account := reviewFixture(t)
	food := &models.Category{Name: "Review food"}
	travel := &models.Category{Name: "Review travel"}
	for _, category := range []*models.Category{food, travel} {
		if err := database.CreateCategory(category); err != nil {
			t.Fatal(err)
		}
	}
	reviewStatement(t, database, account, "all", "2026-07-01", "2026-08-31")
	reviewTransaction(t, database, account, "food-before", "2026-07-03", food.ID, "Cafe", 100, models.TxTypeDebit, false, false)
	reviewTransaction(t, database, account, "travel-before", "2026-07-04", travel.ID, "Train", 500, models.TxTypeDebit, false, false)
	reviewTransaction(t, database, account, "food-after", "2026-08-03", food.ID, "Cafe", 160.25, models.TxTypeDebit, false, false)
	reviewTransaction(t, database, account, "unknown", "2026-08-04", "", "", 20.10, models.TxTypeDebit, false, false)
	reviewTransaction(t, database, account, "transfer", "2026-08-04", food.ID, "Card payment", 9000, models.TxTypeDebit, true, false)
	reviewTransaction(t, database, account, "excluded", "2026-08-05", food.ID, "Wallet", 5000, models.TxTypeDebit, false, true)
	reviewTransaction(t, database, account, "refund", "2026-08-06", food.ID, "Cafe", 60, models.TxTypeCredit, false, false)
	before, _, err := database.ListTransactions(db.TransactionFilter{Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	review, err := database.GetMonthlyReview("2026-08", now)
	if err != nil {
		t.Fatal(err)
	}
	if review.Current.Amount != 180.35 || review.Previous.Amount != 600 || review.Delta != -419.65 || review.Current.Count != 2 {
		t.Fatalf("unexpected totals: %+v", review)
	}
	if !review.CoverageComplete || review.IsPartialMonth || review.NextMonth != "2026-09" {
		t.Fatalf("unexpected review metadata: %+v", review)
	}
	if len(review.Categories) != 3 || review.Categories[0].ID != travel.ID || review.Categories[0].Delta != -500 {
		t.Fatalf("dropped-to-zero category must lead: %+v", review.Categories)
	}
	if review.Categories[1].Merchants[0].Delta != 60.25 {
		t.Fatal("merchant explanation differs from spending")
	}
	legacy, err := database.GetCashFlowIntelligence("2026-08")
	if err != nil {
		t.Fatal(err)
	}
	if legacy.Summary.TotalOutflow != review.Current.Amount {
		t.Fatalf("review disagrees with existing cash flow: %v vs %v", review.Current.Amount, legacy.Summary.TotalOutflow)
	}
	for _, category := range review.Categories {
		for _, previous := range []bool{false, true} {
			evidence, err := database.GetMonthlyReviewEvidence(review.Month, category.ID, previous, 1, now)
			if err != nil {
				t.Fatal(err)
			}
			want := category.Current
			if previous {
				want = category.Previous
			}
			var amount float64
			for _, transaction := range evidence.Items {
				amount += transaction.Amount
			}
			if amount != want.Amount || evidence.Total != want.Count {
				t.Fatalf("evidence doesn't explain %+v: %+v", category, evidence)
			}
		}
	}
	after, _, err := database.ListTransactions(db.TransactionFilter{Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("review changed ledger data")
	}
}

func TestMonthlyReviewCoverageUnionsOverlapsButRetainsGaps(t *testing.T) {
	database, account := reviewFixture(t)
	reviewStatement(t, database, account, "previous", "2026-07-01", "2026-07-31")
	reviewStatement(t, database, account, "first", "2026-08-01", "2026-08-10")
	reviewStatement(t, database, account, "overlap", "2026-08-05", "2026-08-10")
	reviewStatement(t, database, account, "last", "2026-08-12", "2026-08-31")
	reviewStatement(t, database, account, "invalid", "not-a-date", "2026-09-30")
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	review, err := database.GetMonthlyReview("2026-08", now)
	if err != nil {
		t.Fatal(err)
	}
	coverage := review.Coverage[0]
	if review.CoverageComplete || coverage.CurrentDays != 30 || coverage.PreviousDays != 31 || coverage.LatestEnd != "2026-08-31" {
		t.Fatalf("overlaps or invalid dates filled a gap: %+v", coverage)
	}
	reviewStatement(t, database, account, "gap", "2026-08-11", "2026-08-11")
	review, err = database.GetMonthlyReview("2026-08", now)
	if err != nil || !review.CoverageComplete {
		t.Fatalf("filled coverage: %+v, %v", review, err)
	}
	// An account without statements must not appear as complete.
	_, err = database.GetOrCreateAccount("Other Bank", models.AccountTypeCreditCard, "", "9876", "", "", "", "", "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	review, err = database.GetMonthlyReview("2026-08", now)
	if err != nil || review.CoverageComplete || len(review.Coverage) != 2 {
		t.Fatalf("missing account coverage: %+v, %v", review, err)
	}
}

func TestMonthlyReviewComparesElapsedDaysAndCalendarBoundaries(t *testing.T) {
	database, account := reviewFixture(t)
	reviewTransaction(t, database, account, "before-cutoff", "2026-08-12", "", "Cafe", 100, models.TxTypeDebit, false, false)
	reviewTransaction(t, database, account, "after-cutoff", "2026-08-13", "", "Cafe", 500, models.TxTypeDebit, false, false)
	reviewTransaction(t, database, account, "today", "2026-09-12", "", "Cafe", 120, models.TxTypeDebit, false, false)
	reviewTransaction(t, database, account, "future", "2026-09-13", "", "Cafe", 900, models.TxTypeDebit, false, false)
	now := time.Date(2026, 9, 12, 23, 0, 0, 0, time.FixedZone("IST", 19800))
	review, err := database.GetMonthlyReview("2026-09", now)
	if err != nil {
		t.Fatal(err)
	}
	if review.PreviousPeriod.End != "2026-08-12" || review.Current.Amount != 120 || review.Previous.Amount != 100 || !review.IsPartialMonth {
		t.Fatalf("unequal elapsed periods: %+v", review)
	}
	for _, test := range []struct {
		month, previousEnd string
		now                time.Time
	}{
		{"2024-03", "2024-02-29", time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)},
		{"2025-03", "2025-02-28", time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC)},
		{"2026-01", "2025-12-31", now},
	} {
		review, err := database.GetMonthlyReview(test.month, test.now)
		if err != nil || review.PreviousPeriod.End != test.previousEnd {
			t.Fatalf("calendar boundary: %+v, %v", review, err)
		}
	}
	for _, month := range []string{"2026-13", "2026-9", "ALL", "2027-01", "0001-01"} {
		if _, err := database.GetMonthlyReview(month, now); !errors.Is(err, db.ErrInvalidReviewPeriod) {
			t.Errorf("accepted %q: %v", month, err)
		}
	}
}

func TestMonthlyReviewEmptyHistoryAndEvidencePagination(t *testing.T) {
	database, account := reviewFixture(t)
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	review, err := database.GetMonthlyReview("", now)
	if err != nil {
		t.Fatal(err)
	}
	if review.CoverageComplete || review.Current.Amount != 0 || review.Categories == nil || review.AvailableMonths == nil {
		t.Fatalf("empty review: %+v", review)
	}
	for i := 0; i < 51; i++ {
		reviewTransaction(t, database, account, fmt.Sprintf("row-%d", i), "2026-08-01", "", "Cafe", 0.10, models.TxTypeDebit, false, false)
	}
	review, err = database.GetMonthlyReview("", now)
	if err != nil || review.Month != "2026-08" || review.Current.Amount != 5.10 {
		t.Fatalf("default period or decimal arithmetic: %+v, %v", review, err)
	}
	first, err := database.GetMonthlyReviewEvidence(review.Month, "", false, 1, now)
	if err != nil {
		t.Fatal(err)
	}
	second, err := database.GetMonthlyReviewEvidence(review.Month, "", false, 2, now)
	if err != nil {
		t.Fatal(err)
	}
	if first.Total != 51 || len(first.Items) != 50 || len(second.Items) != 1 {
		t.Fatalf("pagination: %+v / %+v", first, second)
	}
	for _, item := range first.Items {
		if item.ID == second.Items[0].ID {
			t.Fatal("pagination repeated a transaction")
		}
	}
}
