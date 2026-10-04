package extractor

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func TestParseIndianAmount(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"1,50,000.00", 150000.00},
		{"2000.50 Cr", 2000.50},
		{"500.00 Dr", 500.00},
		{"-1,200.00", -1200.00},
		{"(4,500.00)", -4500.00},
		{"₹ 87,965.05", 87965.05},
		{"Rs. 250/-", 250.00},
		{"", 0.0},
		{"-", 0.0},
	}

	for _, tc := range tests {
		got, err := ParseIndianAmount(tc.input)
		if err != nil {
			t.Errorf("ParseIndianAmount(%q) returned error: %v", tc.input, err)
		}
		if got != tc.expected {
			t.Errorf("ParseIndianAmount(%q) = %f, expected %f", tc.input, got, tc.expected)
		}
	}
}

func TestNormalizeIndianDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"01/08/2026", "2026-08-01"},
		{"01-08-2026", "2026-08-01"},
		{"01/08/26", "2026-08-01"},
		{"01-Aug-2026", "2026-08-01"},
		{"01-August-2026", "2026-08-01"},
		{"2026-08-01", "2026-08-01"},
	}

	for _, tc := range tests {
		got := NormalizeIndianDate(tc.input)
		if got != tc.expected {
			t.Errorf("NormalizeIndianDate(%q) = %q, expected %q", tc.input, got, tc.expected)
		}
	}
}

func TestIsPDFEncrypted(t *testing.T) {
	unencPath := filepath.Join("..", "..", "..", "samples", "credit_cards", "HDFC_Swiggy_Credit_Card.pdf")
	data, err := os.ReadFile(unencPath)
	if err != nil {
		t.Fatalf("sample not found: %v", err)
	}

	if IsPDFEncrypted(bytes.NewReader(data)) {
		t.Errorf("Expected sample PDF to NOT be encrypted")
	}

	// Encrypt sample in memory
	conf := model.NewAESConfiguration("secret123", "secret123", 256)
	var encBuf bytes.Buffer
	err = api.Encrypt(bytes.NewReader(data), &encBuf, conf)
	if err != nil {
		t.Fatalf("Failed to encrypt PDF in memory: %v", err)
	}

	if !IsPDFEncrypted(bytes.NewReader(encBuf.Bytes())) {
		t.Errorf("Expected in-memory encrypted PDF to be detected as encrypted")
	}
}

func TestExtractCSV(t *testing.T) {
	csvData := `Date,Narration,Chq/Ref No,Withdrawal,Deposit,Balance
01/08/2026,UPI-SWIGGY-123456,123456,630.00,,87335.05
02/08/2026,SALARY CREDIT,REF999,,150000.00,237335.05`

	rows, err := ExtractCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("ExtractCSV error: %v", err)
	}

	if len(rows) != 3 {
		t.Fatalf("Expected 3 rows, got %d", len(rows))
	}

	if rows[0][0] != "Date" || rows[1][1] != "UPI-SWIGGY-123456" {
		t.Errorf("Unexpected cell contents: %v", rows)
	}
}
