package parser

import (
	"io"
	"regexp"
	"strings"

	"local-finance/internal/models"
)

type UnionSavingsPDFParser struct{}

func init() { DefaultRegistry.Register(&UnionSavingsPDFParser{}) }

func (p *UnionSavingsPDFParser) ID() string   { return "union_savings_pdf_v1" }
func (p *UnionSavingsPDFParser) Name() string { return "Union Bank of India Savings Account (PDF)" }
func (p *UnionSavingsPDFParser) SupportedTypes() []StatementType {
	return []StatementType{TypeSavingsPDF}
}

func (p *UnionSavingsPDFParser) CanParse(filename string, sample []byte) (float64, string, models.AccountType) {
	name := strings.ToUpper(filename)
	content := strings.ToUpper(strings.Join(strings.Fields(string(sample)), " "))
	isPDF := strings.HasSuffix(name, ".PDF") || strings.HasPrefix(string(sample), "%PDF")
	hasTable := hasBankHistoryTableText(content, unionSavingsProfile.description)
	if !isPDF || isInvestmentStatement(string(sample)) || strings.Contains(name, "_CC") || strings.Contains(content, "TOTAL AMOUNT DUE") || strings.Contains(content, "CREDIT CARD") && !hasTable {
		return 0, "", models.AccountTypeSavings
	}
	confidence := 0.0
	if strings.Contains(name, "UNION") || strings.Contains(content, "UNION BANK OF INDIA") || strings.Contains(content, "UBIN") {
		confidence += 0.55
	}
	if strings.Contains(content, "DETAILS OF STATEMENT") && strings.Contains(content, "WITHDRAWAL") && strings.Contains(content, "DEPOSIT") {
		confidence += 0.4
	}
	return confidence, "Union Bank of India", models.AccountTypeSavings
}

var unionSavingsProfile = bankHistoryProfile{
	bankName:       "Union Bank of India",
	accountType:    models.AccountTypeSavings,
	formatName:     "PDF",
	identityTerm:   "UBIN",
	identityTerms:  []string{"UBIN", "UNION BANK OF INDIA", "UNION BANK"},
	statementTerm:  "DETAILS OF STATEMENT",
	statementTerms: []string{"DETAILS OF STATEMENT", "STATEMENT OF ACCOUNT", "ACCOUNT STATEMENT"},
	recognition:    []string{"DATE", "WITHDRAWAL", "DEPOSIT", "BALANCE"},
	dateHeaders:    []string{"TRANSACTION DATE", "DATE"},
	description:    []string{"TRANSACTION REMARKS", "PARTICULARS", "NARRATION"},
	debitHeaders:   []string{"WITHDRAWAL", "DEBIT"},
	creditHeaders:  []string{"DEPOSIT", "CREDIT"},
	balanceHeaders: []string{"BALANCE"},
	reference:      []string{"CHQ NUM", "CHEQUE NUMBER", "CHEQUE NO", "REFERENCE"},
	accountPattern: regexp.MustCompile(`(?i)(?:account\s*(?:number|no\.?)|a/c\s*(?:number|no\.?))\s*:?\s*([0-9X*]{8,20})\b`),
	periodPattern:  regexp.MustCompile(`(?i)(?:period\s+from|statement\s+period)\s*:?\s*(\d{2}[./-]\d{2}[./-]\d{2,4})\s*(?:to|-)\s*(\d{2}[./-]\d{2}[./-]\d{2,4})`),
	openingPattern: regexp.MustCompile(`(?i)opening\s+balance\s*:?\s*([0-9][0-9,]*(?:\.\d{2})?)\s*(CR|DR)?`),
	closingPattern: regexp.MustCompile(`(?i)closing\s+balance\s*:?\s*([0-9][0-9,]*(?:\.\d{2})?)\s*(CR|DR)?`),
	debitTotal:     regexp.MustCompile(`(?i)total\s+debits\s*:?\s*([0-9][0-9,]*(?:\.\d{2})?)`),
	creditTotal:    regexp.MustCompile(`(?i)total\s+credits\s*:?\s*([0-9][0-9,]*(?:\.\d{2})?)`),
	stopPattern:    regexp.MustCompile(`(?i)^\s*(?:total\s+debits|closing\s+balance|linked\s+casa|end\s+of\s+statement)\b`),
}

func (p *UnionSavingsPDFParser) Parse(r io.Reader, opts ParseOptions) ([]ParsedTransaction, StatementMeta, error) {
	return parseBankHistoryPDF(r, opts, unionSavingsProfile)
}
