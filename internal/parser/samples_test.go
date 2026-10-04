package parser

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"local-finance/internal/parser/extractor"
)

func TestParseMaskedSampleStatements(t *testing.T) {
	samplesDir := filepath.Join("..", "..", "samples")

	type SampleTestCase struct {
		RelPath             string
		ExpectedBank        string
		ExpectedAccountType string
		ExpectedMask        string
		ExpectedMinTxs      int
	}

	testCases := []SampleTestCase{
		{
			RelPath:             "savings/HDFC_Savings_Account_Statement.pdf",
			ExpectedBank:        "HDFC Bank",
			ExpectedAccountType: "SAVINGS",
			ExpectedMask:        "XX0099",
			ExpectedMinTxs:      10,
		},
		{
			RelPath:             "savings/HDFC_Current_Account_Statement.pdf",
			ExpectedBank:        "HDFC Bank",
			ExpectedAccountType: "CURRENT",
			ExpectedMask:        "XX7496",
			ExpectedMinTxs:      10,
		},
		{
			RelPath:             "savings/HDFC_Savings_Account_Statement.csv",
			ExpectedBank:        "HDFC Bank",
			ExpectedAccountType: "SAVINGS",
			ExpectedMask:        "XX0099",
			ExpectedMinTxs:      5,
		},
		{
			RelPath:             "credit_cards/HDFC_Regalia_Credit_Card.pdf",
			ExpectedBank:        "HDFC Bank",
			ExpectedAccountType: "CREDIT_CARD",
			ExpectedMask:        "XX3638",
			ExpectedMinTxs:      5,
		},
		{
			RelPath:             "credit_cards/HDFC_Swiggy_Credit_Card.pdf",
			ExpectedBank:        "HDFC Bank",
			ExpectedAccountType: "CREDIT_CARD",
			ExpectedMask:        "XX0826",
			ExpectedMinTxs:      5,
		},
		{
			RelPath:             "credit_cards/HDFC_RuPay_Credit_Card.pdf",
			ExpectedBank:        "HDFC Bank",
			ExpectedAccountType: "CREDIT_CARD",
			ExpectedMask:        "XX0161",
			ExpectedMinTxs:      5,
		},
		{
			RelPath:             "credit_cards/ICICI_Amazon_Pay_Credit_Card.pdf",
			ExpectedBank:        "ICICI Bank",
			ExpectedAccountType: "CREDIT_CARD",
			ExpectedMask:        "XX0001",
			ExpectedMinTxs:      5,
		},
		{
			RelPath:             "credit_cards/Axis_Flipkart_Credit_Card.pdf",
			ExpectedBank:        "Axis Bank",
			ExpectedAccountType: "CREDIT_CARD",
			ExpectedMask:        "XX0308",
			ExpectedMinTxs:      4,
		},
		{
			RelPath:             "credit_cards/HDFC_Credit_Card_Statement.csv",
			ExpectedBank:        "HDFC Bank",
			ExpectedAccountType: "CREDIT_CARD",
			ExpectedMask:        "XX3638",
			ExpectedMinTxs:      4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.RelPath, func(t *testing.T) {
			filePath := filepath.Join(samplesDir, tc.RelPath)
			data, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read sample file %s: %v", tc.RelPath, err)
			}

			// 1. Sniff & Detect
			p, confidence, detectedMeta := DefaultRegistry.Detect(tc.RelPath, data)
			if p == nil {
				t.Fatalf("No parser detected for sample %s", tc.RelPath)
			}

			t.Logf("[%s] Parser: %s (Confidence: %.2f, Bank: %s, Type: %s)", tc.RelPath, p.Name(), confidence, detectedMeta.BankName, detectedMeta.AccountType)

			if confidence < 0.5 {
				t.Errorf("Low confidence %.2f for %s", confidence, tc.RelPath)
			}
			if detectedMeta.BankName != tc.ExpectedBank {
				t.Errorf("Expected bank %s, got %s", tc.ExpectedBank, detectedMeta.BankName)
			}
			if string(detectedMeta.AccountType) != tc.ExpectedAccountType {
				t.Errorf("Expected account type %s, got %s", tc.ExpectedAccountType, detectedMeta.AccountType)
			}

			// 2. Parse
			txs, meta, err := p.Parse(bytes.NewReader(data), ParseOptions{Password: ""})
			if err != nil {
				t.Fatalf("Parse error for %s: %v", tc.RelPath, err)
			}

			t.Logf("  ==> Meta: Bank=%s, Type=%s, Mask=%s, Period=%s to %s", meta.BankName, meta.AccountType, meta.AccountNumberMask, meta.StartDate, meta.EndDate)
			if meta.TotalDueAmount > 0 || meta.PaymentDueDate != "" {
				t.Logf("  ==> Bill: DueDate=%s, TotalDue=₹%.2f, MinDue=₹%.2f, Limit=₹%.2f, AvailLimit=₹%.2f, Cashback=₹%.2f",
					meta.PaymentDueDate, meta.TotalDueAmount, meta.MinimumDueAmount,
					meta.CreditLimit, meta.AvailableCreditLimit, meta.CashbackEarned)
			}

			if len(txs) < tc.ExpectedMinTxs {
				t.Errorf("Expected at least %d transactions, parsed %d", tc.ExpectedMinTxs, len(txs))
			}

			if meta.AccountNumberMask != tc.ExpectedMask {
				t.Errorf("Expected mask %s, got %s", tc.ExpectedMask, meta.AccountNumberMask)
			}

			// 3. Verify Deterministic Fingerprints
			hashes := make(map[string]bool)
			for i, tx := range txs {
				rawSig := fmt.Sprintf("%s|%s|%.2f|%s|%s|%s",
					meta.AccountNumberMask, tx.Date, tx.Amount, tx.RawNarration, tx.ReferenceNumber, tx.TxType)
				h := sha256.Sum256([]byte(rawSig))
				hexHash := fmt.Sprintf("%x", h)

				if hashes[hexHash] {
					t.Errorf("Duplicate transaction fingerprint at row %d (%s): %s", i, tx.RawNarration, hexHash)
				}
				hashes[hexHash] = true

				t.Logf("    [%02d] %s | %-6s | ₹%9.2f | Payee: %-30s | Mode: %-12s | Ref: %-15s",
					i, tx.Date, tx.TxType, tx.Amount, tx.CleanedPayee, tx.PaymentMode, tx.ReferenceNumber)
			}

			t.Logf("  ==> Verified %d valid transactions (%d unique fingerprints)", len(txs), len(hashes))
		})
	}
}

func TestEncryptedPDFStatementParsing(t *testing.T) {
	samplesDir := filepath.Join("..", "..", "samples")
	pdfPath := filepath.Join(samplesDir, "savings", "HDFC_Savings_Account_Statement.pdf")

	rawData, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("Failed to read sample PDF: %v", err)
	}

	testPassword := "SamplePass123"
	conf := model.NewAESConfiguration(testPassword, testPassword, 256)
	var encBuf bytes.Buffer
	err = api.Encrypt(bytes.NewReader(rawData), &encBuf, conf)
	if err != nil {
		t.Fatalf("Failed to encrypt sample PDF in memory: %v", err)
	}
	encData := encBuf.Bytes()

	// Verify encryption detector
	if !extractor.IsPDFEncrypted(bytes.NewReader(encData)) {
		t.Errorf("Expected IsPDFEncrypted to return true for encrypted PDF")
	}

	p, _, _ := DefaultRegistry.Detect("HDFC_Savings_Account_Statement.pdf", encData)
	if p == nil {
		t.Fatalf("Failed to detect parser for encrypted statement")
	}

	// Parsing without password should fail
	_, _, err = p.Parse(bytes.NewReader(encData), ParseOptions{Password: ""})
	if err == nil {
		t.Errorf("Expected parse without password to fail, but it succeeded")
	}

	// Parsing with correct password should succeed
	txs, meta, err := p.Parse(bytes.NewReader(encData), ParseOptions{Password: testPassword})
	if err != nil {
		t.Fatalf("Failed to parse encrypted PDF with password: %v", err)
	}

	if len(txs) == 0 {
		t.Errorf("Expected parsed transactions from encrypted PDF, got 0")
	}
	if meta.BankName != "HDFC Bank" {
		t.Errorf("Expected HDFC Bank, got %s", meta.BankName)
	}
}
