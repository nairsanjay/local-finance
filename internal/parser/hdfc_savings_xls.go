package parser

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"local-finance/internal/models"
	"local-finance/internal/parser/extractor"
)

type HDFCSavingsXLSParser struct{}

func init() {
	DefaultRegistry.Register(&HDFCSavingsXLSParser{})
}

func (p *HDFCSavingsXLSParser) ID() string {
	return "hdfc_savings_xls_v1"
}

func (p *HDFCSavingsXLSParser) Name() string {
	return "HDFC Bank Savings/Current Account (Excel)"
}

func (p *HDFCSavingsXLSParser) SupportedTypes() []StatementType {
	return []StatementType{TypeSavingsCSV}
}

func (p *HDFCSavingsXLSParser) CanParse(filename string, sample []byte) (float64, string, models.AccountType) {
	nameUpper := strings.ToUpper(filename)
	isExcel := strings.HasSuffix(nameUpper, ".XLS") || strings.HasSuffix(nameUpper, ".XLSX")

	if !isExcel {
		return 0.0, "", models.AccountTypeSavings
	}

	confidence := 0.3
	if strings.Contains(nameUpper, "ACCT_STATEMENT") || strings.Contains(nameUpper, "HDFC") {
		confidence += 0.4
	}

	return confidence, "HDFC Bank", models.AccountTypeSavings
}

var (
	dateRangeRegex = regexp.MustCompile(`(?i)Statement\s*From\s*:\s*([0-9/.-]+)\s*To\s*:\s*([0-9/.-]+)`)
)

func (p *HDFCSavingsXLSParser) Parse(r io.Reader, opts ParseOptions) ([]ParsedTransaction, StatementMeta, error) {
	rows, err := extractor.ExtractExcel(r, opts.Password)
	if err != nil {
		return nil, StatementMeta{}, fmt.Errorf("failed to extract excel rows: %w", err)
	}

	meta := StatementMeta{
		BankName:        "HDFC Bank",
		AccountType:     models.AccountTypeSavings,
		StatementFormat: "XLS",
	}

	// 1. Scan preamble metadata
	for i := 0; i < len(rows) && i < 25; i++ {
		rowStr := strings.Join(rows[i], " ")
		if matches := dateRangeRegex.FindStringSubmatch(rowStr); len(matches) > 2 {
			meta.StartDate = extractor.NormalizeIndianDate(matches[1])
			meta.EndDate = extractor.NormalizeIndianDate(matches[2])
		}
	}

	// 2. Find table header row
	requiredCols := []extractor.ColumnSpec{
		{Name: "date", Keywords: []string{"Date", "Txn Date", "Transaction Date"}},
		{Name: "narration", Keywords: []string{"Narration", "Description", "Particulars"}},
		{Name: "ref_no", Keywords: []string{"Chq", "Ref", "Cheque"}},
		{Name: "value_date", Keywords: []string{"Value Dt", "Value Date"}},
		{Name: "withdrawal", Keywords: []string{"Withdrawal", "Debit", "Dr"}},
		{Name: "deposit", Keywords: []string{"Deposit", "Credit", "Cr"}},
		{Name: "balance", Keywords: []string{"Closing Balance", "Balance", "Running"}},
	}

	headerIdx, colMap, found := extractor.FindHeaderRow(rows, requiredCols)
	if !found {
		return nil, meta, fmt.Errorf("could not find HDFC transaction table header row in Excel file")
	}

	dateCol := colMap["date"]
	narrationCol := colMap["narration"]
	refNoCol, hasRefNo := colMap["ref_no"]
	valDateCol, hasValDate := colMap["value_date"]
	withdrawalCol, hasWithdrawal := colMap["withdrawal"]
	depositCol, hasDeposit := colMap["deposit"]
	balanceCol, hasBalance := colMap["balance"]

	// Scan summary for opening balance if present
	for i := len(rows) - 1; i >= 0 && i >= headerIdx; i-- {
		rowStr := strings.ToUpper(strings.Join(rows[i], " "))
		if strings.Contains(rowStr, "OPENING BALANCE") && i+1 < len(rows) {
			for _, cell := range rows[i+1] {
				if amt, err := extractor.ParseIndianAmount(cell); err == nil && amt > 0 {
					meta.OpeningBalance = amt
					break
				}
			}
		}
	}

	var transactions []ParsedTransaction
	runningBal := meta.OpeningBalance

	for r := headerIdx + 1; r < len(rows); r++ {
		row := rows[r]
		if len(row) <= dateCol {
			continue
		}

		rawDate := row[dateCol]
		normDate := extractor.NormalizeIndianDate(rawDate)
		if normDate == "" || strings.Contains(rawDate, "*") {
			// Not a transaction row (e.g. divider row or summary row)
			continue
		}

		rawNarration := ""
		if len(row) > narrationCol {
			rawNarration = row[narrationCol]
		}
		if rawNarration == "" || strings.Contains(rawNarration, "***") {
			continue
		}

		refNo := ""
		if hasRefNo && len(row) > refNoCol {
			refNo = strings.TrimLeft(row[refNoCol], "0")
			if refNo == "" && strings.TrimSpace(row[refNoCol]) != "" {
				refNo = "0"
			}
		}

		var valDatePtr *string
		if hasValDate && len(row) > valDateCol {
			vDate := extractor.NormalizeIndianDate(row[valDateCol])
			if vDate != "" {
				valDatePtr = &vDate
			}
		}

		var withdrawalAmt, depositAmt float64
		if hasWithdrawal && len(row) > withdrawalCol {
			withdrawalAmt, _ = extractor.ParseIndianAmount(row[withdrawalCol])
		}
		if hasDeposit && len(row) > depositCol {
			depositAmt, _ = extractor.ParseIndianAmount(row[depositCol])
		}

		txType := models.TxTypeDebit
		amount := withdrawalAmt
		if depositAmt > 0 {
			txType = models.TxTypeCredit
			amount = depositAmt
		}

		if amount == 0 {
			continue
		}

		// Update running balance if present in row or calculated
		var runBalPtr *float64
		if hasBalance && len(row) > balanceCol {
			if bAmt, err := extractor.ParseIndianAmount(row[balanceCol]); err == nil && bAmt > 0 {
				runBalPtr = &bAmt
				runningBal = bAmt
			}
		}
		if runBalPtr == nil {
			if txType == models.TxTypeCredit {
				runningBal += amount
			} else {
				runningBal -= amount
			}
			balCopy := runningBal
			runBalPtr = &balCopy
		}

		// Clean narration with Indian regex engine
		cleaned := CleanNarration(rawNarration)
		if cleaned.ReferenceNumber != "" && refNo == "" {
			refNo = cleaned.ReferenceNumber
		}

		transactions = append(transactions, ParsedTransaction{
			Date:            normDate,
			ValueDate:       valDatePtr,
			RawNarration:    rawNarration,
			CleanedPayee:    cleaned.CleanedPayee,
			PaymentMode:     cleaned.PaymentMode,
			ReferenceNumber: refNo,
			TxType:          txType,
			Amount:          amount,
			RunningBalance:  runBalPtr,
			UPIVPA:          cleaned.UPIVPA,
			CardLast4:       cleaned.CardLast4,
			IsTransfer:      cleaned.IsTransfer,
		})
	}

	if len(transactions) > 0 {
		if meta.StartDate == "" {
			meta.StartDate = transactions[len(transactions)-1].Date
		}
		if meta.EndDate == "" {
			meta.EndDate = transactions[0].Date
		}
	}

	return transactions, meta, nil
}
