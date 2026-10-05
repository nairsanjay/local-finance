package parser

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"local-finance/internal/models"
	"local-finance/internal/parser/extractor"
)

// Shared bank-history parsing utilities. Institution-specific adapters keep
// their own detection and statement profiles; these helpers handle the common
// positional-table parsing and reconciliation mechanics.
//
// bankHistoryProfile describes the adapter-provided headings and controls.
type bankHistoryProfile struct {
	bankName       string
	accountType    models.AccountType
	formatName     string
	recognition    []string
	identityTerm   string
	statementTerm  string
	dateHeaders    []string
	description    []string
	debitHeaders   []string
	creditHeaders  []string
	balanceHeaders []string
	reference      []string
	periodPattern  *regexp.Regexp
	openingPattern *regexp.Regexp
	closingPattern *regexp.Regexp
	debitTotal     *regexp.Regexp
	creditTotal    *regexp.Regexp
	stopPattern    *regexp.Regexp
}

type bankHistoryColumns struct {
	date, description, debit, credit, balance, reference float64
	hasReference                                         bool
}

type bankHistoryRow struct {
	tx      ParsedTransaction
	balance float64
}

var (
	bankHistoryDatePrefix = regexp.MustCompile(`^(?:\d{2}[./-]\d{2}[./-]\d{2,4}|\d{4}[./-]\d{2}[./-]\d{2})\b`)
	bankHistoryAmount     = regexp.MustCompile(`(?i)^\s*(?:₹|INR\s*)?([0-9][0-9,]*(?:\.\d{2})?)(?:\s*(CR|DR))?\s*$`)
	bankHistoryMoneyToken = regexp.MustCompile(`[0-9][0-9,]*\.[0-9]{2}`)
)

func parseBankHistoryPDF(r io.Reader, opts ParseOptions, profile bankHistoryProfile) ([]ParsedTransaction, StatementMeta, error) {
	rows, err := extractor.ExtractPDFPositionalRows(r, opts.Password)
	if err != nil {
		return nil, StatementMeta{}, fmt.Errorf("failed to extract %s statement: %w", profile.bankName, err)
	}
	return parseBankHistoryRows(rows, opts.Filename, profile)
}

func parseBankHistoryRows(rows []extractor.PositionalRow, filename string, profile bankHistoryProfile) ([]ParsedTransaction, StatementMeta, error) {
	meta := StatementMeta{BankName: profile.bankName, AccountType: profile.accountType, StatementFormat: profile.formatName}
	var headerIndexes []int
	var columns bankHistoryColumns
	documentText := strings.ToUpper(bankHistoryDocumentText(rows))
	if !strings.Contains(documentText, strings.ToUpper(profile.identityTerm)) || !strings.Contains(documentText, strings.ToUpper(profile.statementTerm)) {
		return nil, meta, fmt.Errorf("document does not match the %s account statement format", profile.bankName)
	}
	for index, row := range rows {
		if candidate, ok := discoverBankHistoryColumns(row, profile); ok {
			headerIndexes = append(headerIndexes, index)
			columns = candidate
		}
	}
	if len(headerIndexes) == 0 {
		return nil, meta, fmt.Errorf("%s statement must contain one unambiguous transaction table", profile.bankName)
	}
	headerIndex := headerIndexes[0]

	// Extract only explicitly labeled statement controls. They are optional for
	// bank exports that provide no summary; per-row balances still reconcile.
	openingFound, closingFound, debitTotalFound, creditTotalFound := false, false, false, false
	for _, row := range rows {
		line := bankHistoryRowText(row)
		upper := strings.ToUpper(line)
		if profile.periodPattern != nil && meta.StartDate == "" {
			if match := profile.periodPattern.FindStringSubmatch(line); len(match) == 3 {
				meta.StartDate = strictIndianDate(match[1])
				meta.EndDate = strictIndianDate(match[2])
			}
		}
		if profile.openingPattern != nil && strings.Contains(upper, "OPENING BALANCE") {
			amount, ok := parseLabeledHistoryAmount(profile.openingPattern, line)
			if !ok || openingFound && !historyMoneyEqual(meta.OpeningBalance, amount) {
				return nil, meta, fmt.Errorf("%s statement has a malformed or conflicting opening balance", profile.bankName)
			}
			meta.OpeningBalance, openingFound = amount, true
		}
		if profile.closingPattern != nil && strings.Contains(upper, "CLOSING BALANCE") {
			amount, ok := parseLabeledHistoryAmount(profile.closingPattern, line)
			if !ok || closingFound && !historyMoneyEqual(meta.ClosingBalance, amount) {
				return nil, meta, fmt.Errorf("%s statement has a malformed or conflicting closing balance", profile.bankName)
			}
			meta.ClosingBalance, closingFound = amount, true
		}
		if profile.debitTotal != nil && (strings.Contains(upper, "TOTAL DEBIT") || strings.Contains(upper, "TOTAL WITHDRAWAL")) {
			amount, ok := parseLabeledHistoryAmount(profile.debitTotal, line)
			if !ok || debitTotalFound && !historyMoneyEqual(meta.TotalDebits, amount) {
				return nil, meta, fmt.Errorf("%s statement has a malformed or conflicting debit total", profile.bankName)
			}
			meta.TotalDebits, debitTotalFound = amount, true
		}
		if profile.creditTotal != nil && (strings.Contains(upper, "TOTAL CREDIT") || strings.Contains(upper, "TOTAL DEPOSIT")) {
			amount, ok := parseLabeledHistoryAmount(profile.creditTotal, line)
			if !ok || creditTotalFound && !historyMoneyEqual(meta.TotalCredits, amount) {
				return nil, meta, fmt.Errorf("%s statement has a malformed or conflicting credit total", profile.bankName)
			}
			meta.TotalCredits, creditTotalFound = amount, true
		}
	}

	parsed := make([]bankHistoryRow, 0, len(rows)-headerIndex)
	active := true
	for index := headerIndex + 1; index < len(rows); index++ {
		row := rows[index]
		line := bankHistoryRowText(row)
		if profile.stopPattern != nil && profile.stopPattern.MatchString(line) {
			active = false
		}
		if !active || len(row.Elements) == 0 {
			continue
		}
		dateIndex, dateValue := bankHistoryDateCell(row, columns.date, columns.description)
		if dateIndex < 0 {
			continue
		}
		date := strictIndianDate(dateValue)
		if date == "" {
			return nil, meta, fmt.Errorf("%s statement contains an invalid transaction date", profile.bankName)
		}
		debit, debitOK, err := amountInBankColumn(row, columns.debit, columns.credit)
		if err != nil {
			return nil, meta, fmt.Errorf("%s statement contains an ambiguous withdrawal amount", profile.bankName)
		}
		credit, creditOK, err := amountInBankColumn(row, columns.credit, columns.balance)
		if err != nil {
			return nil, meta, fmt.Errorf("%s statement contains an ambiguous deposit amount", profile.bankName)
		}
		balance, balanceOK, err := amountInBankColumn(row, columns.balance, 0)
		if err != nil || !balanceOK {
			return nil, meta, fmt.Errorf("%s statement contains a missing or ambiguous running balance", profile.bankName)
		}
		if !debitOK && !creditOK {
			return nil, meta, fmt.Errorf("%s statement contains a dated row without a withdrawal or deposit", profile.bankName)
		}
		if debitOK && creditOK && debit > 0 && credit > 0 {
			return nil, meta, fmt.Errorf("%s statement row contains both a withdrawal and deposit", profile.bankName)
		}
		amount, direction := debit, models.TxTypeDebit
		if credit > 0 {
			amount, direction = credit, models.TxTypeCredit
		}
		if amount <= 0 {
			return nil, meta, fmt.Errorf("%s statement contains a zero or negative transaction amount", profile.bankName)
		}
		balanceCell := bankHistoryCellInColumn(row, columns.balance, 0)
		if strings.HasSuffix(strings.ToUpper(strings.TrimSpace(balanceCell)), "DR") {
			balance = -absFloat(balance)
		} else if strings.HasSuffix(strings.ToUpper(strings.TrimSpace(balanceCell)), "CR") {
			balance = absFloat(balance)
		}
		narration := bankHistoryNarration(row, columns)
		if narration == "" {
			return nil, meta, fmt.Errorf("%s statement contains a dated transaction without narration", profile.bankName)
		}
		cleaned := CleanNarration(narration)
		ref := bankHistoryReference(row, columns)
		if ref == "" {
			ref = cleaned.ReferenceNumber
		}
		parsed = append(parsed, bankHistoryRow{tx: ParsedTransaction{
			Date: date, RawNarration: narration, CleanedPayee: cleaned.CleanedPayee,
			PaymentMode: cleaned.PaymentMode, ReferenceNumber: ref, TxType: direction,
			Amount: amount, IsTransfer: cleaned.IsTransfer,
		}, balance: balance})
	}
	if len(parsed) == 0 {
		return nil, meta, fmt.Errorf("%s transaction table contains no complete rows", profile.bankName)
	}

	// Normalize source ordering before checking every adjacent running balance.
	sort.SliceStable(parsed, func(i, j int) bool { return parsed[i].tx.Date < parsed[j].tx.Date })
	debits, credits := 0.0, 0.0
	for index, row := range parsed {
		if row.tx.TxType == models.TxTypeDebit {
			debits += row.tx.Amount
		} else {
			credits += row.tx.Amount
		}
		previous := row.balance
		if row.tx.TxType == models.TxTypeCredit {
			previous -= row.tx.Amount
		} else {
			previous += row.tx.Amount
		}
		if index == 0 {
			if !openingFound {
				meta.OpeningBalance = previous
			} else if !historyMoneyEqual(meta.OpeningBalance, previous) {
				return nil, meta, fmt.Errorf("%s opening balance does not reconcile with its first running balance", profile.bankName)
			}
			continue
		}
		if !historyMoneyEqual(parsed[index-1].balance, previous) {
			return nil, meta, fmt.Errorf("%s running balances do not reconcile between dated transactions", profile.bankName)
		}
	}
	if !closingFound {
		meta.ClosingBalance = parsed[len(parsed)-1].balance
	} else if !historyMoneyEqual(meta.ClosingBalance, parsed[len(parsed)-1].balance) {
		return nil, meta, fmt.Errorf("%s closing balance does not match the last running balance", profile.bankName)
	}
	if debitTotalFound && !historyMoneyEqual(meta.TotalDebits, debits) {
		return nil, meta, fmt.Errorf("%s withdrawals do not match the printed debit total", profile.bankName)
	}
	if creditTotalFound && !historyMoneyEqual(meta.TotalCredits, credits) {
		return nil, meta, fmt.Errorf("%s deposits do not match the printed credit total", profile.bankName)
	}
	if meta.TotalDebits == 0 {
		meta.TotalDebits = debits
	}
	if meta.TotalCredits == 0 {
		meta.TotalCredits = credits
	}
	if meta.OpeningBalance+credits-debits-meta.ClosingBalance > 0.01 || meta.ClosingBalance-(meta.OpeningBalance+credits-debits) > 0.01 {
		return nil, meta, fmt.Errorf("%s opening balance, transactions, and closing balance do not reconcile", profile.bankName)
	}
	if meta.StartDate == "" || meta.EndDate == "" {
		meta.StartDate, meta.EndDate = parsed[0].tx.Date, parsed[len(parsed)-1].tx.Date
	}
	transactions := make([]ParsedTransaction, len(parsed))
	for index := range parsed {
		balance := parsed[index].balance
		transactions[index] = parsed[index].tx
		transactions[index].RunningBalance = &balance
	}
	return transactions, meta, nil
}

func discoverBankHistoryColumns(row extractor.PositionalRow, profile bankHistoryProfile) (bankHistoryColumns, bool) {
	text := bankHistoryRowText(row)
	upper := strings.ToUpper(text)
	if !containsAllBankTerms(upper, profile.recognition) {
		return bankHistoryColumns{}, false
	}
	columns := bankHistoryColumns{date: -1, description: -1, debit: -1, credit: -1, balance: -1, reference: -1}
	for _, element := range row.Elements {
		label := strings.ToUpper(strings.Join(strings.Fields(element.S), " "))
		switch {
		case columns.date < 0 && matchesBankTerm(label, profile.dateHeaders):
			columns.date = element.X
		case columns.description < 0 && matchesBankTerm(label, profile.description):
			columns.description = element.X
		case columns.debit < 0 && matchesBankTerm(label, profile.debitHeaders):
			columns.debit = element.X
		case columns.credit < 0 && matchesBankTerm(label, profile.creditHeaders):
			columns.credit = element.X
		case columns.balance < 0 && matchesBankTerm(label, profile.balanceHeaders):
			columns.balance = element.X
		case !columns.hasReference && matchesBankTerm(label, profile.reference):
			columns.reference, columns.hasReference = element.X, true
		}
	}
	if columns.date < 0 || columns.description < 0 || columns.debit < 0 || columns.credit < 0 || columns.balance < 0 {
		return bankHistoryColumns{}, false
	}
	if !(columns.date < columns.description && columns.description < columns.debit && columns.debit < columns.credit && columns.credit < columns.balance) {
		return bankHistoryColumns{}, false
	}
	return columns, true
}

func containsAllBankTerms(text string, terms []string) bool {
	for _, term := range terms {
		if !strings.Contains(text, strings.ToUpper(term)) {
			return false
		}
	}
	return true
}

func matchesBankTerm(text string, terms []string) bool {
	for _, term := range terms {
		if strings.Contains(text, strings.ToUpper(term)) {
			return true
		}
	}
	return false
}

func bankHistoryRowText(row extractor.PositionalRow) string {
	parts := make([]string, 0, len(row.Elements))
	for _, element := range row.Elements {
		parts = append(parts, strings.TrimSpace(element.S))
	}
	return strings.Join(parts, " ")
}

func bankHistoryDocumentText(rows []extractor.PositionalRow) string {
	var lines []string
	for _, row := range rows {
		lines = append(lines, bankHistoryRowText(row))
	}
	return strings.Join(lines, "\n")
}

func bankHistoryDateCell(row extractor.PositionalRow, dateX, descriptionX float64) (int, string) {
	for index, element := range row.Elements {
		if element.X < dateX-10 || element.X >= descriptionX {
			continue
		}
		value := strings.TrimSpace(element.S)
		if bankHistoryDatePrefix.MatchString(value) {
			return index, bankHistoryDatePrefix.FindString(value)
		}
	}
	return -1, ""
}

func strictIndianDate(raw string) string {
	normalized := extractor.NormalizeIndianDate(raw)
	if _, err := time.Parse("2006-01-02", normalized); err != nil {
		return ""
	}
	return normalized
}

func amountInBankColumn(row extractor.PositionalRow, start, end float64) (float64, bool, error) {
	cell := bankHistoryCellInColumn(row, start, end)
	if cell == "" || cell == "-" || cell == "--" {
		return 0, false, nil
	}
	match := bankHistoryAmount.FindStringSubmatch(cell)
	if len(match) != 3 {
		return 0, false, fmt.Errorf("invalid amount cell")
	}
	amount, err := extractor.ParseIndianAmount(match[1])
	if err != nil || amount < 0 {
		return 0, false, fmt.Errorf("invalid amount value")
	}
	return amount, amount > 0, nil
}

func bankHistoryCellInColumn(row extractor.PositionalRow, start, end float64) string {
	var values []string
	for _, element := range row.Elements {
		if element.X < start || (end > start && element.X >= end) {
			continue
		}
		value := strings.TrimSpace(element.S)
		if value != "" {
			values = append(values, value)
		}
	}
	return strings.Join(values, " ")
}

func bankHistoryNarration(row extractor.PositionalRow, columns bankHistoryColumns) string {
	var parts []string
	for _, element := range row.Elements {
		if element.X < columns.description || element.X >= columns.debit || (columns.hasReference && element.X >= columns.reference && element.X < columns.debit) {
			continue
		}
		value := strings.TrimSpace(element.S)
		if bankHistoryDatePrefix.MatchString(value) {
			continue
		}
		if _, err := extractor.ParseIndianAmount(value); err == nil && bankHistoryMoneyToken.MatchString(value) {
			continue
		}
		parts = append(parts, value)
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func bankHistoryReference(row extractor.PositionalRow, columns bankHistoryColumns) string {
	if !columns.hasReference {
		return ""
	}
	for _, element := range row.Elements {
		if element.X < columns.reference || element.X >= columns.debit {
			continue
		}
		value := strings.TrimSpace(element.S)
		if value != "" && !bankHistoryDatePrefix.MatchString(value) {
			return value
		}
	}
	return ""
}

func parseLabeledHistoryAmount(pattern *regexp.Regexp, line string) (float64, bool) {
	if pattern == nil {
		return 0, false
	}
	match := pattern.FindStringSubmatch(line)
	if len(match) < 2 {
		return 0, false
	}
	amount, err := extractor.ParseIndianAmount(match[1])
	if err != nil || amount < 0 {
		return 0, false
	}
	if len(match) > 2 && strings.EqualFold(strings.TrimSpace(match[2]), "DR") {
		return -absFloat(amount), true
	}
	return absFloat(amount), true
}

func historyMoneyEqual(a, b float64) bool { return absFloat(a-b) <= 0.01 }

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
