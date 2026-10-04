package parser

import (
	"fmt"
	"io"
	"strings"

	"local-finance/internal/models"
	"local-finance/internal/parser/extractor"
)

type HDFCSavingsCSVParser struct{}

func init() {
	DefaultRegistry.Register(&HDFCSavingsCSVParser{})
}

func (p *HDFCSavingsCSVParser) ID() string {
	return "hdfc_savings_csv_v1"
}

func (p *HDFCSavingsCSVParser) Name() string {
	return "HDFC Bank Savings/Current Account CSV"
}

func (p *HDFCSavingsCSVParser) SupportedTypes() []StatementType {
	return []StatementType{TypeSavingsCSV}
}

func (p *HDFCSavingsCSVParser) CanParse(filename string, sample []byte) (float64, string, models.AccountType) {
	content := strings.ToUpper(string(sample))
	confidence := 0.0

	// Check signatures
	if strings.Contains(content, "HDFC BANK") || strings.Contains(strings.ToLower(filename), "hdfc") {
		confidence += 0.4
	}

	// Typical HDFC Savings CSV Column Keywords
	if strings.Contains(content, "WITHDRAWAL") && strings.Contains(content, "DEPOSIT") && strings.Contains(content, "CLOSING BALANCE") {
		confidence += 0.5
	} else if strings.Contains(content, "NARRATION") && strings.Contains(content, "CHQ/REF") {
		confidence += 0.4
	}

	return confidence, "HDFC Bank", models.AccountTypeSavings
}

func (p *HDFCSavingsCSVParser) Parse(r io.Reader, opts ParseOptions) ([]ParsedTransaction, StatementMeta, error) {
	rows, err := extractor.ExtractCSV(r)
	if err != nil {
		return nil, StatementMeta{}, fmt.Errorf("failed to extract CSV rows: %w", err)
	}

	meta := StatementMeta{
		BankName:        "HDFC Bank",
		AccountType:     models.AccountTypeSavings,
		StatementFormat: "CSV",
	}

	// 1. Scan preamble metadata
	for i := 0; i < len(rows) && i < 25; i++ {
		rowStr := strings.Join(rows[i], " ")
		if matches := hdfcPDFAccMaskRegex.FindStringSubmatch(rowStr); len(matches) > 1 && meta.AccountNumberMask == "" {
			rawAcc := matches[1]
			if !strings.ContainsAny(rawAcc, "X*") && len(rawAcc) >= 8 {
				meta.AccountNumber = rawAcc
			}
			meta.AccountNumberMask = "XX" + rawAcc[max(0, len(rawAcc)-4):]
		}
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
		return nil, meta, fmt.Errorf("could not find HDFC transaction table header row in CSV file")
	}

	dateCol := colMap["date"]
	narrationCol := colMap["narration"]
	refNoCol, hasRefNo := colMap["ref_no"]
	valDateCol, hasValDate := colMap["value_date"]
	withdrawalCol, hasWithdrawal := colMap["withdrawal"]
	depositCol, hasDeposit := colMap["deposit"]
	balanceCol, hasBalance := colMap["balance"]

	var transactions []ParsedTransaction

	for r := headerIdx + 1; r < len(rows); r++ {
		row := rows[r]
		if len(row) <= dateCol {
			continue
		}

		rawDate := row[dateCol]
		normDate := extractor.NormalizeIndianDate(rawDate)
		if normDate == "" || strings.Contains(rawDate, "*") {
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

		var runBalPtr *float64
		if hasBalance && len(row) > balanceCol {
			if bAmt, err := extractor.ParseIndianAmount(row[balanceCol]); err == nil && bAmt > 0 {
				runBalPtr = &bAmt
			}
		}

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

func containsDigits(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
