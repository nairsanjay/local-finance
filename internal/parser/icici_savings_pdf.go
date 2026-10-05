package parser

import (
	"io"
	"regexp"
	"slices"
	"strings"

	"local-finance/internal/models"
	"local-finance/internal/parser/extractor"
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
	content := strings.ToUpper(strings.Join(strings.Fields(string(sample)), " "))
	isPDF := strings.HasSuffix(name, ".PDF") || strings.HasPrefix(string(sample), "%PDF")
	hasTable := hasBankHistoryTableText(content, iciciSavingsProfile.description)
	if !isPDF || strings.Contains(name, "_CC") || strings.Contains(content, "TOTAL AMOUNT DUE") || strings.Contains(content, "CREDIT CARD") && !hasTable {
		return 0, "", models.AccountTypeSavings
	}
	confidence := 0.0
	if strings.Contains(name, "ICICI") || strings.Contains(content, "ICICI BANK") {
		confidence += 0.5
	}
	if hasTable || strings.Contains(content, "STATEMENT OF TRANSACTIONS IN SAVING ACCOUNT") ||
		(strings.Contains(content, "TRANSACTION REMARKS") || strings.Contains(content, "PARTICULARS")) && strings.Contains(content, "WITHDRAWAL") && strings.Contains(content, "DEPOSIT") {
		confidence += 0.45
	}
	return confidence, "ICICI Bank", models.AccountTypeSavings
}

var iciciSavingsProfile = bankHistoryProfile{
	bankName:       "ICICI Bank",
	accountType:    models.AccountTypeSavings,
	formatName:     "PDF",
	identityTerm:   "ICICI Bank",
	recognition:    []string{"DATE", "BALANCE"},
	dateHeaders:    []string{"TRANSACTION DATE", "DATE"},
	description:    []string{"TRANSACTION REMARKS", "REMARKS", "PARTICULARS"},
	debitHeaders:   []string{"WITHDRAWAL", "DEBIT"},
	creditHeaders:  []string{"DEPOSIT", "CREDIT"},
	balanceHeaders: []string{"BALANCE"},
	reference:      []string{"CHEQUE NUMBER", "CHEQUE NO", "CHQ NO", "REFERENCE"},
	accountPattern: regexp.MustCompile(`(?i)(?:account\s*(?:number|no\.?)|a/c\s*(?:number|no\.?))\s*:?\s*([0-9X*]{8,20})\b`),
	periodPattern:  regexp.MustCompile(`(?i)(?:statement\s+period|period\s+from|statement\s+from)\s*:?\s*(\d{2}[./-]\d{2}[./-]\d{2,4})\s*(?:to|-)\s*(\d{2}[./-]\d{2}[./-]\d{2,4})`),
	openingPattern: regexp.MustCompile(`(?i)opening\s+balance\s*:?\s*([0-9][0-9,]*(?:\.\d{2})?)\s*(CR|DR)?`),
	closingPattern: regexp.MustCompile(`(?i)closing\s+balance\s*:?\s*([0-9][0-9,]*(?:\.\d{2})?)\s*(CR|DR)?`),
	debitTotal:     regexp.MustCompile(`(?i)total\s+(?:debits|withdrawals)\s*:?\s*([0-9][0-9,]*(?:\.\d{2})?)`),
	creditTotal:    regexp.MustCompile(`(?i)total\s+(?:credits|deposits)\s*:?\s*([0-9][0-9,]*(?:\.\d{2})?)`),
	stopPattern:    regexp.MustCompile(`(?i)^\s*(?:total\s+(?:debits|withdrawals)|statement\s+summary|end\s+of\s+statement)\b`),
}

func (p *ICICISavingsPDFParser) Parse(r io.Reader, opts ParseOptions) ([]ParsedTransaction, StatementMeta, error) {
	rows, err := extractor.ExtractPDFPositionalRows(r, opts.Password)
	if err != nil {
		return nil, StatementMeta{}, err
	}
	return parseBankHistoryRows(coalesceICICIHistoryHeaders(rows), iciciSavingsProfile)
}

// Transaction-history exports split their table title over three baselines.
// Assemble only literal heading tokens on one page within the printed 10pt
// span. Transaction data never supplies a missing heading.
func coalesceICICIHistoryHeaders(rows []extractor.PositionalRow) []extractor.PositionalRow {
	var result []extractor.PositionalRow
	for index := 0; index < len(rows); index++ {
		row := rows[index]
		if _, complete := discoverBankHistoryColumns(row, iciciSavingsProfile); complete || !iciciHistoryHeadingRow(row) {
			result = append(result, row)
			continue
		}
		merged := extractor.PositionalRow{Page: row.Page, Y: row.Y, Elements: slices.Clone(row.Elements)}
		for end := index + 1; end < len(rows) && end <= index+2 && rows[end].Page == row.Page && absFloat(row.Y-rows[end].Y) <= 10.1; end++ {
			if !iciciHistoryHeadingRow(rows[end]) {
				break
			}
			merged.Elements = append(merged.Elements, rows[end].Elements...)
			candidate, ok := joinICICIHistoryHeaderWords(merged)
			if _, complete := discoverBankHistoryColumns(candidate, iciciSavingsProfile); ok && complete {
				row, index = candidate, end
				break
			}
		}
		result = append(result, row)
	}
	return result
}

func iciciHistoryHeadingRow(row extractor.PositionalRow) bool {
	if len(row.Elements) == 0 {
		return false
	}
	for _, element := range row.Elements {
		switch strings.ToUpper(strings.Join(strings.Fields(element.S), " ")) {
		case "S", "NO.", "S NO.", "S.NO", "TRANSACTION", "DATE", "TRANSACTION DATE", "CHEQUE", "NUMBER", "CHEQUE NUMBER", "REMARKS", "TRANSACTION REMARKS", "WITHDRAWAL", "DEPOSIT", "BALANCE", "AMOUNT (INR)", "(INR)":
		default:
			return false
		}
	}
	return true
}

func joinICICIHistoryHeaderWords(row extractor.PositionalRow) (extractor.PositionalRow, bool) {
	row.Elements = slices.Clone(row.Elements)
	row.Elements = slices.DeleteFunc(row.Elements, func(element extractor.PositionalElement) bool {
		label := strings.ToUpper(strings.Join(strings.Fields(element.S), " "))
		return label == "AMOUNT (INR)" || label == "(INR)"
	})
	for index := range row.Elements {
		if strings.ToUpper(strings.Join(strings.Fields(row.Elements[index].S), " ")) == "S NO." {
			row.Elements[index].S = "S.NO"
		}
	}
	for _, pair := range []struct {
		first, second, joined string
		gap                   float64
	}{
		{"TRANSACTION", "DATE", "TRANSACTION DATE", 24},
		{"TRANSACTION", "REMARKS", "TRANSACTION REMARKS", 65},
		{"CHEQUE", "NUMBER", "CHEQUE NUMBER", 40},
		{"S", "NO.", "S.NO", 15},
	} {
		for first := 0; first < len(row.Elements); first++ {
			if strings.ToUpper(strings.TrimSpace(row.Elements[first].S)) != pair.first {
				continue
			}
			second := -1
			for index, element := range row.Elements {
				if strings.ToUpper(strings.TrimSpace(element.S)) != pair.second || element.X < row.Elements[first].X || element.X-row.Elements[first].X > pair.gap {
					continue
				}
				if second >= 0 {
					return row, false
				}
				second = index
			}
			if second >= 0 {
				row.Elements[first].S = pair.joined
				if row.Elements[second].EndX > row.Elements[first].EndX {
					row.Elements[first].EndX = row.Elements[second].EndX
				}
				row.Elements = slices.Delete(row.Elements, second, second+1)
				if second < first {
					first--
				}
			}
		}
	}
	for _, element := range row.Elements {
		switch strings.ToUpper(strings.TrimSpace(element.S)) {
		case "TRANSACTION", "CHEQUE", "NUMBER", "S", "NO.":
			return row, false
		}
	}
	return row, true
}
