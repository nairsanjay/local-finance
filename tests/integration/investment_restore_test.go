package integration_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"local-finance/internal/service"
)

func TestInvestmentReadsDuringRestore(t *testing.T) {
	database := testDatabase(t)
	data, err := os.ReadFile("../../samples/investments/zerodha-fictional.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	imported, _, err := service.NewInvestmentService(database).Import("example.xlsx", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	backupPath := filepath.Join(t.TempDir(), "backup.db")
	if err := database.BackupTo(backupPath); err != nil {
		t.Fatal(err)
	}
	backup, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	errors := make(chan error, 2)
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 10; i++ {
			if err := database.RestoreFrom(bytes.NewReader(backup)); err != nil {
				errors <- err
				return
			}
		}
	}()
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 50; i++ {
			snapshots, err := database.ListInvestmentSnapshots()
			if err != nil {
				errors <- err
				return
			}
			if len(snapshots) != 1 || snapshots[0].ID != imported.ID {
				errors <- fmt.Errorf("restore exposed inconsistent snapshots: %+v", snapshots)
				return
			}
			if _, err := database.ExportAllDataJSON(); err != nil {
				errors <- err
				return
			}
		}
	}()
	close(start)
	workers.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}
