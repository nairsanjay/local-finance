package integration_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/xuri/excelize/v2"
	"local-finance/internal/service"
)

func TestZerodhaAuxiliaryWorksheetsWithoutCombined(t *testing.T) {
	data, err := os.ReadFile("../../samples/investments/zerodha-fictional.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	for _, malformed := range []bool{false, true} {
		name := "summary and disclaimer retained"
		if malformed {
			name = "incomplete holdings table rejected"
		}
		t.Run(name, func(t *testing.T) {
			workbook, err := excelize.OpenReader(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			defer workbook.Close()
			if err := workbook.DeleteSheet("Equity"); err != nil {
				t.Fatal(err)
			}
			if err := workbook.SetSheetName("Combined", "Equity"); err != nil {
				t.Fatal(err)
			}
			for _, sheet := range []string{"Summary", "Disclaimer"} {
				if _, err := workbook.NewSheet(sheet); err != nil {
					t.Fatal(err)
				}
				if err := workbook.SetCellValue(sheet, "A1", "Fictional report information"); err != nil {
					t.Fatal(err)
				}
			}
			if malformed {
				if _, err := workbook.NewSheet("Incomplete holdings"); err != nil {
					t.Fatal(err)
				}
				headers := []interface{}{"Symbol", "ISIN", "Quantity Available"}
				if err := workbook.SetSheetRow("Incomplete holdings", "A1", &headers); err != nil {
					t.Fatal(err)
				}
			}
			modified, err := workbook.WriteToBuffer()
			if err != nil {
				t.Fatal(err)
			}
			svc := service.NewInvestmentService(testDatabase(t))
			preview, err := svc.Preview("generic.xlsx", bytes.NewReader(modified.Bytes()))
			if malformed {
				if err == nil {
					t.Fatal("incomplete holdings table silently ignored")
				}
				if _, _, err := svc.Import("generic.xlsx", bytes.NewReader(modified.Bytes())); err == nil {
					t.Fatal("incomplete holdings imported")
				}
				return
			}
			if err != nil || len(preview.Holdings) != 2 || len(preview.Sheets) != 3 || preview.CurrentValue == nil || *preview.CurrentValue != 900 {
				t.Fatalf("auxiliary sheets prevented a complete preview: %+v %v", preview, err)
			}
			if preview.Sheets[1].Name != "Summary" || preview.Sheets[2].Name != "Disclaimer" || preview.Sheets[2].Rows[0][0] != "Fictional report information" {
				t.Fatalf("original auxiliary worksheets lost: %+v", preview.Sheets)
			}
			imported, duplicate, err := svc.Import("generic.xlsx", bytes.NewReader(modified.Bytes()))
			if err != nil || duplicate || len(imported.Holdings) != 2 || len(imported.Sheets) != 3 {
				t.Fatalf("auxiliary worksheets prevented import: %+v %v %v", imported, duplicate, err)
			}
		})
	}
}

func TestZerodhaSeparateAssetWorksheets(t *testing.T) {
	data, err := os.ReadFile("../../samples/investments/zerodha-fictional.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []string{"valid", "missing fund header", "duplicate ISIN"} {
		t.Run(scenario, func(t *testing.T) {
			workbook, err := excelize.OpenReader(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			defer workbook.Close()
			rows, err := workbook.GetRows("Combined")
			if err != nil {
				t.Fatal(err)
			}
			header, isinColumn := -1, -1
			var fund []string
			for i, row := range rows {
				for j, value := range row {
					if value == "ISIN" {
						header, isinColumn = i, j
					}
				}
				if len(row) > 0 && row[0] == "DEMO FUND" {
					fund = append([]string{}, row...)
				}
			}
			if header < 0 || len(fund) <= isinColumn {
				t.Fatal("fixture missing fund or holdings header")
			}
			if err := workbook.DeleteSheet("Combined"); err != nil {
				t.Fatal(err)
			}
			if _, err := workbook.NewSheet("Mutual Funds"); err != nil {
				t.Fatal(err)
			}
			if scenario == "duplicate ISIN" {
				fund[isinColumn] = "INE000DEMO01"
			}
			for i, row := range append(rows[:header+1], fund) {
				axis, _ := excelize.CoordinatesToCellName(1, i+1)
				if err := workbook.SetSheetRow("Mutual Funds", axis, &row); err != nil {
					t.Fatal(err)
				}
			}
			if scenario == "missing fund header" {
				axis, _ := excelize.CoordinatesToCellName(isinColumn+1, header+1)
				if err := workbook.SetCellValue("Mutual Funds", axis, "Missing"); err != nil {
					t.Fatal(err)
				}
			}
			modified, err := workbook.WriteToBuffer()
			if err != nil {
				t.Fatal(err)
			}
			svc := service.NewInvestmentService(testDatabase(t))
			preview, previewErr := svc.Preview("generic.xlsx", bytes.NewReader(modified.Bytes()))
			imported, duplicate, importErr := svc.Import("generic.xlsx", bytes.NewReader(modified.Bytes()))
			if scenario != "valid" {
				if previewErr == nil || importErr == nil {
					t.Fatalf("%s allowed a partial or duplicated portfolio", scenario)
				}
				return
			}
			if previewErr != nil || importErr != nil || duplicate || len(imported.Holdings) != 2 || len(preview.Holdings) != 2 || imported.CurrentValue == nil || *imported.CurrentValue != 900 || imported.InvestedValue == nil || *imported.InvestedValue != 800 {
				t.Fatalf("separate asset worksheets lost holdings or values: %+v %v %v", imported, previewErr, importErr)
			}
			if imported.Holdings[0].AssetClass != "Equity" || imported.Holdings[1].AssetClass != "Mutual Fund" {
				t.Fatalf("separate worksheet asset classes incorrect: %+v", imported.Holdings)
			}
		})
	}
}
