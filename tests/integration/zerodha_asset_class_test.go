package integration_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/xuri/excelize/v2"
	"local-finance/internal/investment"
)

func TestZerodhaAssetClasses(t *testing.T) {
	data, err := os.ReadFile("../../samples/investments/zerodha-fictional.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ instrument, sheet, expected string }{
		{"EQ", "Combined", "Equity"},
		{"BE", "Combined", "Equity"},
		{"-", "Combined", "Equity"},
		{"", "Combined", "Equity"},
		{"OTHER", "Combined", "Equity"},
		{"mf", "Combined", "Mutual Fund"},
		{"Mutual Fund", "Combined", "Mutual Fund"},
		{"-", "Mutual Funds", "Mutual Fund"},
	} {
		t.Run(tc.sheet+"/"+tc.instrument, func(t *testing.T) {
			workbook, err := excelize.OpenReader(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			defer workbook.Close()
			rows, err := workbook.GetRows("Combined")
			if err != nil {
				t.Fatal(err)
			}
			instrumentColumn, header := -1, -1
			for i, row := range rows {
				for j, value := range row {
					if value == "Instrument Type" {
						instrumentColumn, header = j, i
					}
				}
			}
			if header < 0 {
				t.Fatal("fixture missing instrument header")
			}
			for i := header + 1; i < len(rows); i++ {
				axis, _ := excelize.CoordinatesToCellName(instrumentColumn+1, i+1)
				if err := workbook.SetCellValue("Combined", axis, tc.instrument); err != nil {
					t.Fatal(err)
				}
			}
			if tc.sheet != "Combined" {
				if err := workbook.DeleteSheet("Equity"); err != nil {
					t.Fatal(err)
				}
				if err := workbook.SetSheetName("Combined", tc.sheet); err != nil {
					t.Fatal(err)
				}
			}
			modified, err := workbook.WriteToBuffer()
			if err != nil {
				t.Fatal(err)
			}
			snapshot, err := (investment.ZerodhaHoldingsParser{}).Parse(modified.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			if len(snapshot.Holdings) != 2 {
				t.Fatalf("holdings missing: %+v", snapshot.Holdings)
			}
			for _, holding := range snapshot.Holdings {
				if holding.AssetClass != tc.expected {
					t.Fatalf("instrument %q on %s classified as %q, want %q", tc.instrument, tc.sheet, holding.AssetClass, tc.expected)
				}
			}
		})
	}
}
