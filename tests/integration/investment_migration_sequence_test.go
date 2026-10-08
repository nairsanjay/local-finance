package integration_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"local-finance/internal/db"
	"local-finance/internal/service"
)

func TestMigrationVersionsAreSequential(t *testing.T) {
	files, err := os.ReadDir("../../internal/db/migrations")
	if err != nil {
		t.Fatal(err)
	}
	expected := 1
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}
		version, err := strconv.Atoi(strings.SplitN(file.Name(), "_", 2)[0])
		if err != nil || version != expected {
			t.Fatalf("want migration %05d, found %s", expected, file.Name())
		}
		expected++
	}
}

func TestInvestmentMigrationReservationBackfillsPreviews(t *testing.T) {
	for _, previewVersion := range []int{16, 17} {
		t.Run(fmt.Sprint(previewVersion), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "preview.db")
			database, err := db.NewDB(path)
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile("../../samples/investments/zerodha-fictional.xlsx")
			if err != nil {
				t.Fatal(err)
			}
			snapshot, _, err := service.NewInvestmentService(database).Import("fictional.xlsx", bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			if err := database.Close(); err != nil {
				t.Fatal(err)
			}
			conn, err := sql.Open("sqlite", path)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			var originalJSON, originalSchema string
			if err := conn.QueryRow(`SELECT data_json FROM investment_snapshots WHERE id=?`, snapshot.ID).Scan(&originalJSON); err != nil {
				t.Fatal(err)
			}
			if err := conn.QueryRow(`SELECT sql FROM sqlite_master WHERE name='investment_snapshots'`).Scan(&originalSchema); err != nil {
				t.Fatal(err)
			}
			if _, err := conn.Exec(`DELETE FROM goose_db_version WHERE version_id=15 OR version_id>?`, previewVersion); err != nil {
				t.Fatal(err)
			}
			for attempt := 0; attempt < 2; attempt++ {
				reopened, err := db.NewDB(path)
				if err != nil {
					t.Fatalf("preview startup failed: %v", err)
				}
				if err := reopened.Close(); err != nil {
					t.Fatal(err)
				}
				var actualJSON, actualSchema string
				var reservationCount, currentVersion int
				if err := conn.QueryRow(`SELECT data_json FROM investment_snapshots WHERE id=?`, snapshot.ID).Scan(&actualJSON); err != nil {
					t.Fatal(err)
				}
				if err := conn.QueryRow(`SELECT sql FROM sqlite_master WHERE name='investment_snapshots'`).Scan(&actualSchema); err != nil {
					t.Fatal(err)
				}
				if err := conn.QueryRow(`SELECT COUNT(*) FROM goose_db_version WHERE version_id=15 AND is_applied=1`).Scan(&reservationCount); err != nil {
					t.Fatal(err)
				}
				if err := conn.QueryRow(`SELECT MAX(version_id) FROM goose_db_version WHERE is_applied=1`).Scan(&currentVersion); err != nil {
					t.Fatal(err)
				}
				if actualJSON != originalJSON || actualSchema != originalSchema || reservationCount != 1 || currentVersion != 17 {
					t.Fatalf("reservation changed snapshot/schema or repeated: count=%d version=%d", reservationCount, currentVersion)
				}
			}
		})
	}
}

func TestInvestmentMigrationReservationKeepsOtherGapsStrict(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid-history.db")
	database, err := db.NewDB(path)
	if err != nil {
		t.Fatal(err)
	}
	database.Close()
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`DELETE FROM goose_db_version WHERE version_id IN (12,15)`); err != nil {
		t.Fatal(err)
	}
	conn.Close()
	reopened, err := db.NewDB(path)
	if reopened != nil {
		reopened.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "12") {
		t.Fatalf("unrelated missing migration was permitted: %v", err)
	}
}

func TestInvestmentMigrationsUpgradeMainVersion14(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.db")
	database, err := db.NewDB(path)
	if err != nil {
		t.Fatal(err)
	}
	database.Close()
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.Exec(`DROP TABLE investment_snapshots`); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`DELETE FROM goose_db_version WHERE version_id>=15`); err != nil {
		t.Fatal(err)
	}
	upgraded, err := db.NewDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer upgraded.Close()
	var applied int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM goose_db_version WHERE version_id BETWEEN 15 AND 17 AND is_applied=1`).Scan(&applied); err != nil {
		t.Fatal(err)
	}
	snapshots, err := upgraded.ListInvestmentSnapshots()
	if err != nil || len(snapshots) != 0 || applied != 3 {
		t.Fatalf("main upgrade did not apply 15/16/17: count=%d snapshots=%+v err=%v", applied, snapshots, err)
	}
}
