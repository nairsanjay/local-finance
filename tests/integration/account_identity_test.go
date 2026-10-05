package integration_test

import (
	"testing"

	"local-finance/internal/models"
)

func TestAccountIdentityKeepsFullNumbersDistinctAndUpgradesMasks(t *testing.T) {
	database := testDatabase(t)
	get := func(number, mask string) *models.Account {
		t.Helper()
		account, err := database.GetOrCreateAccount("Example Bank", models.AccountTypeSavings, number, mask, "", "", "", "", "", "", nil)
		if err != nil {
			t.Fatal(err)
		}
		return account
	}

	masked := get("", "XX1234")
	upgraded := get("123456781234", "XX1234")
	if upgraded.ID != masked.ID || upgraded.AccountNumber == nil || *upgraded.AccountNumber != "123456781234" {
		t.Fatalf("mask-only account was not upgraded: %+v", upgraded)
	}
	if repeated := get("", "XX1234"); repeated.ID != upgraded.ID {
		t.Fatal("masked-only reimport no longer matches the known account")
	}
	other := get("987654321234", "XX1234")
	if other.ID == upgraded.ID {
		t.Fatal("different full account numbers with the same last four merged")
	}
	if repeated := get("123456781234", "XX1234"); repeated.ID != upgraded.ID {
		t.Fatal("full number reimport did not choose its exact account")
	}
	if repeated := get("987654321234", "XX1234"); repeated.ID != other.ID {
		t.Fatal("second full number reimport did not choose its exact account")
	}

	anonymous := get("", "")
	if repeated := get("", ""); repeated.ID != anonymous.ID {
		t.Fatal("unidentified imports no longer reuse their unidentified account")
	}
	identified := get("112233445566", "")
	if identified.ID == anonymous.ID {
		t.Fatal("an empty mask matched an unidentified account despite a supplied full number")
	}
	accounts, err := database.ListAccounts()
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 4 {
		t.Fatalf("got %d accounts; want 4", len(accounts))
	}
}
