package parser

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"local-finance/internal/models"
	"local-finance/internal/parser/extractor"
)

type HDFCCreditCardCSVParser struct{}

func init() {
	DefaultRegistry.Register(&HDFCCreditCardCSVParser{})
}

func (p *HDFCCreditCardCSVParser) ID() string {
	return "hdfc_credit_card_csv_v1"
}

func (p *HDFCCreditCardCSVParser) Name() string {
	return "HDFC Credit Card Statement CSV"
}

func (p *HDFCCreditCardCSVParser) SupportedTypes() []StatementType {
	return []StatementType{TypeCreditCardCSV}
}

func (p *HDFCCreditCardCSVParser) CanParse(filename string, sample []byte) (float64, string, models.AccountType) {
	content := strings.ToUpper(string(sample))
	confidence := 0.0

	if strings.Contains(content, "HDFC") || strings.Contains(strings.ToLower(filename), "hdfc") {
		confidence += 0.3
	}

	if strings.Contains(content, "CARD") || strings.Contains(content, "CREDIT CARD") || strings.Contains(content, "REWARD POINTS") {
		confidence += 0.3
	}

	if (strings.Contains(content, "CR/DR") || strings.Contains(content, "BILLING PERIOD")) && strings.Contains(content, "TRANSACTION DESCRIPTION") {
		confidence += 0.4
	}

	return confidence, "HDFC Bank", models.AccountTypeCreditCard
}

func (p *HDFCCreditCardCSVParser) Parse(r io.Reader, opts ParseOptions) ([]ParsedTransaction, StatementMeta, error) {
	rows, err := extractor.ExtractCSV(r)
	if err != nil {
		return nil, StatementMeta{}, fmt.Errorf("failed to extract credit card CSV rows: %w", err)
	}

	meta := StatementMeta{
		BankName:        "HDFC Bank",
		AccountType:     models.AccountTypeCreditCard,
		StatementFormat: "CSV",
	}

	// 1. Scan preamble metadata
	nameRegex := regexp.MustCompile(`(?i)(?:Customer\s+Name|Name)\s*[:\s]*([A-Za-z\s]{3,50})`)
	for i := 0; i < len(rows) && i < 20; i++ {
		line := strings.Join(rows[i], " ")
		if matches := nameRegex.FindStringSubmatch(line); len(matches) > 1 && meta.AccountHolderName == "" {
			meta.AccountHolderName = extractor.CleanCustomerName(matches[1])
		}
		if matches := hdfcCCCardNoRegex.FindStringSubmatch(line); len(matches) > 1 && meta.AccountNumberMask == "" {
			rawCard := matches[1]
			meta.AccountNumberMask = "XX" + rawCard[max(0, len(rawCard)-4):]
		}
	}

	// 2. Find table header row
	requiredCols := []extractor.ColumnSpec{
		{Name: "date", Keywords: []string{"Date", "Txn Date"}},
		{Name: "narration", Keywords: []string{"Description", "Transaction", "Particulars"}},
		{Name: "amount", Keywords: []string{"Amount", "Amt"}},
		{Name: "type", Keywords: []string{"CR/DR", "Type"}},
	}

	headerIdx, colMap, found := extractor.FindHeaderRow(rows, requiredCols)
	if !found {
		return nil, meta, fmt.Errorf("could not find credit card table header row in CSV file")
	}

	dateCol := colMap["date"]
	descCol := colMap["narration"]
	amtCol := colMap["amount"]
	typeCol, hasTypeCol := colMap["type"]

	var transactions []ParsedTransaction

	for r := headerIdx + 1; r < len(rows); r++ {
		row := rows[r]
		if len(row) <= dateCol || len(row) <= descCol || len(row) <= amtCol {
			continue
		}

		rawDate := row[dateCol]
		normDate := extractor.NormalizeIndianDate(rawDate)
		if normDate == "" || !containsDigits(rawDate) {
			continue
		}

		rawDesc := row[descCol]
		if rawDesc == "" {
			continue
		}

		cleaned := CleanNarration(rawDesc)
		amt, err := extractor.ParseIndianAmount(row[amtCol])
		if err != nil || amt <= 0 {
			continue
		}

		txType := models.TxTypeDebit // CC default is expense/debit
		if hasTypeCol && len(row) > typeCol {
			t := strings.ToUpper(strings.TrimSpace(row[typeCol]))
			if strings.Contains(t, "CR") || strings.Contains(t, "CREDIT") || strings.Contains(strings.ToUpper(rawDesc), "PAYMENT RECEIVED") {
				txType = models.TxTypeCredit
			}
		} else if strings.Contains(strings.ToUpper(rawDesc), "PAYMENT RECEIVED") || strings.Contains(strings.ToUpper(rawDesc), "AUTO DEBIT") {
			txType = models.TxTypeCredit
		}

		transactions = append(transactions, ParsedTransaction{
			Date:            normDate,
			RawNarration:    rawDesc,
			CleanedPayee:    cleaned.CleanedPayee,
			PaymentMode:     models.PaymentModeCardOnline,
			ReferenceNumber: cleaned.ReferenceNumber,
			TxType:          txType,
			Amount:          amt,
			CardLast4:       cleaned.CardLast4,
			IsTransfer:      cleaned.IsTransfer,
		})
	}

	if len(transactions) > 0 {
		meta.StartDate = transactions[len(transactions)-1].Date
		meta.EndDate = transactions[0].Date
		if meta.StartDate > meta.EndDate {
			meta.StartDate, meta.EndDate = meta.EndDate, meta.StartDate
		}
	}

	return transactions, meta, nil
}
