package parser

import (
	"io"
	"regexp"
	"strings"

	"local-finance/internal/models"
)

type ICICISavingsPDFParser struct{}

func init() { DefaultRegistry.Register(&ICICISavingsPDFParser{}) }

func (p *ICICISavingsPDFParser) ID() string   { return "icici_savings_pdf_v1" }
func (p *ICICISavingsPDFParser) Name() string { return "ICICI Bank Savings Account (PDF)" }
func (p *ICICISavingsPDFParser) SupportedTypes() []StatementType {
	return []StatementType{TypeSavingsPDF}
}

func (p *ICICISavingsPDFParser) CanParse(filename string, sample []byte) (float64, string, models.AccountType) {
	name := strings.ToUpper(filename)
	content := strings.ToUpper(string(sample))
	isPDF := strings.HasSuffix(name, ".PDF") || strings.HasPrefix(string(sample), "%PDF")
	if !isPDF || strings.Contains(content, "CREDIT CARD") || strings.Contains(name, "_CC") {
		return 0, "", models.AccountTypeSavings
	}
	confidence := 0.0
	if strings.Contains(name, "ICICI") || strings.Contains(content, "ICICI BANK") {
		confidence += 0.5
	}
	if strings.Contains(content, "STATEMENT OF TRANSACTIONS IN SAVING ACCOUNT") ||
		strings.Contains(content, "TRANSACTION REMARKS") && strings.Contains(content, "WITHDRAWAL") {
		confidence += 0.45
	}
	return confidence, "ICICI Bank", models.AccountTypeSavings
}

var iciciSavingsProfile = bankHistoryProfile{
	bankName:       "ICICI Bank",
	accountType:    models.AccountTypeSavings,
	formatName:     "PDF",
	identityTerm:   "ICICI Bank",
	statementTerm:  "Statement of Transactions in Saving Account",
	recognition:    []string{"DATE", "TRANSACTION REMARKS", "WITHDRAWAL", "DEPOSIT", "BALANCE"},
	dateHeaders:    []string{"TRANSACTION DATE", "DATE"},
	description:    []string{"TRANSACTION REMARKS", "REMARKS", "PARTICULARS"},
	debitHeaders:   []string{"WITHDRAWAL", "DEBIT"},
	creditHeaders:  []string{"DEPOSIT", "CREDIT"},
	balanceHeaders: []string{"BALANCE"},
	reference:      []string{"CHEQUE NUMBER", "CHEQUE NO", "CHQ NO", "REFERENCE"},
	periodPattern:  regexp.MustCompile(`(?i)(?:statement\s+period|period\s+from|statement\s+from)\s*:?\s*(\d{2}[./-]\d{2}[./-]\d{2,4})\s*(?:to|-)\s*(\d{2}[./-]\d{2}[./-]\d{2,4})`),
	openingPattern: regexp.MustCompile(`(?i)opening\s+balance\s*:?\s*([0-9][0-9,]*(?:\.\d{2})?)\s*(CR|DR)?`),
	closingPattern: regexp.MustCompile(`(?i)closing\s+balance\s*:?\s*([0-9][0-9,]*(?:\.\d{2})?)\s*(CR|DR)?`),
	debitTotal:     regexp.MustCompile(`(?i)total\s+(?:debits|withdrawals)\s*:?\s*([0-9][0-9,]*(?:\.\d{2})?)`),
	creditTotal:    regexp.MustCompile(`(?i)total\s+(?:credits|deposits)\s*:?\s*([0-9][0-9,]*(?:\.\d{2})?)`),
	stopPattern:    regexp.MustCompile(`(?i)^\s*(?:total\s+(?:debits|withdrawals)|statement\s+summary|end\s+of\s+statement)\b`),
}

func (p *ICICISavingsPDFParser) Parse(r io.Reader, opts ParseOptions) ([]ParsedTransaction, StatementMeta, error) {
	return parseBankHistoryPDF(r, opts, iciciSavingsProfile)
}
