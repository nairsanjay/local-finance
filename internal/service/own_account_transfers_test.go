package service

import (
	"fmt"
	"path/filepath"
	"testing"

	"local-finance/internal/db"
	"local-finance/internal/models"
)

func TestOwnAccountReferenceMatching(t *testing.T) {
	for _, tc := range []struct {
		name, debitRef, creditRef, narration, creditDate    string
		ambiguous, excluded, differentCurrency, sameAccount bool
		want                                                int
	}{
		{name: "shared reference", debitRef: "123456789012", creditRef: "123456789012", want: 1},
		{name: "structured UPIAR", narration: "UPIAR/123456789012/DR/ANYONE/BANK/person", creditRef: "123456789012", want: 1},
		{name: "settlement delay", debitRef: "123456789012", creditRef: "123456789012", creditDate: "2026-04-04", want: 1},
		{name: "too late", debitRef: "123456789012", creditRef: "123456789012", creditDate: "2026-04-05"},
		{name: "name alone", narration: "SELF PAYMENT TO SAME NAME"},
		{name: "missing one reference", debitRef: "123456789012"},
		{name: "conflicting references", debitRef: "123456789012", creditRef: "123456789013"},
		{name: "placeholder", debitRef: "000000000000", creditRef: "000000000000"},
		{name: "ambiguous", debitRef: "123456789012", creditRef: "123456789012", ambiguous: true},
		{name: "excluded", debitRef: "123456789012", creditRef: "123456789012", excluded: true},
		{name: "different currencies", debitRef: "123456789012", creditRef: "123456789012", differentCurrency: true},
		{name: "same account", debitRef: "123456789012", creditRef: "123456789012", sameAccount: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			database, err := db.NewDB(filepath.Join(t.TempDir(), "test.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer database.Close()
			a, err := database.GetOrCreateAccount("Bank A", models.AccountTypeSavings, "", "XX0001", "", "", "", "", "", "", nil)
			if err != nil {
				t.Fatal(err)
			}
			b, err := database.GetOrCreateAccount("Bank B", models.AccountTypeSavings, "", "XX0002", "", "", "", "", "", "", nil)
			if err != nil {
				t.Fatal(err)
			}
			if tc.differentCurrency {
				if _, err := database.Exec("UPDATE accounts SET currency='USD' WHERE id=?", b.ID); err != nil {
					t.Fatal(err)
				}
			}
			if tc.sameAccount {
				b = a
			}
			creditDate := tc.creditDate
			if creditDate == "" {
				creditDate = "2026-04-01"
			}
			debit := &models.Transaction{AccountID: a.ID, TxHash: "debit-hash", TxDate: "2026-04-01", TxType: models.TxTypeDebit, Amount: 500, ReferenceNumber: tc.debitRef, RawNarration: tc.narration, Notes: "keep note", Tags: "keep tag", IsExcluded: tc.excluded}
			credit := &models.Transaction{AccountID: b.ID, TxHash: "credit-hash", TxDate: creditDate, TxType: models.TxTypeCredit, Amount: 500, ReferenceNumber: tc.creditRef, CleanedPayee: "SALARY"}
			for _, tx := range []*models.Transaction{credit, debit} {
				if _, err := database.UpsertTransaction(tx); err != nil {
					t.Fatal(err)
				}
			}
			if tc.ambiguous {
				duplicate := *credit
				duplicate.ID = ""
				duplicate.TxHash = "ambiguous-credit"
				if _, err := database.UpsertTransaction(&duplicate); err != nil {
					t.Fatal(err)
				}
			}
			svc := NewReconciliationService(database)
			count, err := svc.ReconcileOwnAccountTransfers()
			if err != nil || count != tc.want {
				t.Fatalf("count=%d want=%d err=%v", count, tc.want, err)
			}
			count, err = svc.ReconcileOwnAccountTransfers()
			if err != nil || count != 0 {
				t.Fatalf("not idempotent: %d %v", count, err)
			}
			if tc.want == 1 {
				overview, err := database.GetAnalyticsOverview()
				if err != nil {
					t.Fatal(err)
				}
				if overview.TotalIncome != 0 || overview.TotalExpense != 0 {
					t.Fatalf("transfer in totals: %+v", overview)
				}
				salary, err := database.GetSalaryInsights()
				if err != nil {
					t.Fatal(err)
				}
				if salary.TotalPaychecks != 0 {
					t.Fatal("transfer counted as salary")
				}
				if _, err := database.UpsertTransaction(debit); err != nil {
					t.Fatal(err)
				}
				var flag bool
				var hash, note, tags string
				rows, err := database.Query("SELECT is_transfer,tx_hash,notes,tags FROM transactions WHERE tx_hash=?", debit.TxHash)
				if err != nil {
					t.Fatal(err)
				}
				if !rows.Next() {
					rows.Close()
					t.Fatal("missing transaction")
				}
				err = rows.Scan(&flag, &hash, &note, &tags)
				rows.Close()
				if err != nil {
					t.Fatal(err)
				}
				if !flag || hash != debit.TxHash || note != "keep note" || tags != "keep tag" {
					t.Fatal("reimport lost transfer or user data")
				}
			}
			if tc.name == "name alone" || tc.ambiguous || tc.name == "missing one reference" {
				pairs, err := svc.getOwnAccountCandidates()
				if err != nil || len(pairs) == 0 {
					t.Fatalf("missing review candidates: %v", err)
				}
				for _, p := range pairs {
					if p.MatchConfidence >= .85 {
						t.Fatal("unsafe automatic candidate")
					}
				}
			}
		})
	}
}

func TestOwnAccountMatchingBeyondRecentHistory(t *testing.T) {
	database, err := db.NewDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	a, _ := database.GetOrCreateAccount("A", models.AccountTypeSavings, "", "XX0001", "", "", "", "", "", "", nil)
	b, _ := database.GetOrCreateAccount("B", models.AccountTypeCurrent, "", "XX0002", "", "", "", "", "", "", nil)
	for i := 0; i < 220; i++ {
		_, err := database.UpsertTransaction(&models.Transaction{AccountID: a.ID, TxHash: fmt.Sprint("new", i), TxDate: "2026-09-01", TxType: models.TxTypeDebit, Amount: 1})
		if err != nil {
			t.Fatal(err)
		}
	}
	for i, typ := range []models.TxType{models.TxTypeDebit, models.TxTypeCredit} {
		acc := a
		if i == 1 {
			acc = b
		}
		_, err := database.UpsertTransaction(&models.Transaction{AccountID: acc.ID, TxHash: fmt.Sprint("old", i), TxDate: "2025-04-01", TxType: typ, Amount: 700, ReferenceNumber: "123456789012"})
		if err != nil {
			t.Fatal(err)
		}
	}
	count, err := NewReconciliationService(database).ReconcileOwnAccountTransfers()
	if count != 1 || err != nil {
		t.Fatalf("historical pair: %d %v", count, err)
	}
}
