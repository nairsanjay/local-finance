package parser

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"local-finance/internal/models"
	"local-finance/internal/parser/extractor"
)

type HDFCSavingsPDFParser struct{}

func init() {
	DefaultRegistry.Register(&HDFCSavingsPDFParser{})
}

func (p *HDFCSavingsPDFParser) ID() string {
	return "hdfc_savings_pdf_v1"
}

func (p *HDFCSavingsPDFParser) Name() string {
	return "HDFC Bank Savings/Current Account (PDF)"
}

func (p *HDFCSavingsPDFParser) SupportedTypes() []StatementType {
	return []StatementType{TypeSavingsPDF}
}

var (
	hdfcBulkFilenameRegex = regexp.MustCompile(`(?i)^\d{4}[0-9X*]{6,10}\d{4}.*_TO_.*\.PDF$`)
)

func (p *HDFCSavingsPDFParser) CanParse(filename string, sample []byte) (float64, string, models.AccountType) {
	nameUpper := strings.ToUpper(filename)
	content := strings.ToUpper(string(sample))
	isPDF := strings.HasSuffix(nameUpper, ".PDF") || (len(sample) > 4 && string(sample[:4]) == "%PDF")

	if !isPDF {
		return 0.0, "", models.AccountTypeSavings
	}

	// If it contains credit card keywords, do not parse as savings/current
	if strings.Contains(nameUpper, "REGALIA") || strings.Contains(nameUpper, "RUPAY") ||
		strings.Contains(nameUpper, "SWIGGY") || strings.Contains(nameUpper, "INFINIA") ||
		strings.Contains(nameUpper, "MILLENNIA") || strings.Contains(nameUpper, "_CC") ||
		strings.Contains(content, "TOTAL AMOUNT DUE") || strings.Contains(content, "CREDIT CARD") {
		return 0.0, "", models.AccountTypeSavings
	}

	confidence := 0.0
	if strings.Contains(nameUpper, "HDFC") || strings.Contains(content, "HDFC BANK") || strings.Contains(content, "HDFC00") {
		confidence += 0.5
	}
	if strings.Contains(nameUpper, "ACCT_STATEMENT") || strings.Contains(nameUpper, "ACCOUNT_STATEMENT") {
		confidence += 0.5
	}

	baseName := filepath.Base(filename)
	if hdfcBulkFilenameRegex.MatchString(baseName) {
		confidence += 0.85
	}
	if strings.Contains(nameUpper, "_TO_") && (strings.Contains(nameUpper, "WITHOUT") || strings.Contains(nameUpper, "0356") || strings.Contains(nameUpper, "HDFC")) {
		confidence += 0.7
	}

	accountType := models.AccountTypeSavings
	if strings.Contains(nameUpper, "CURRENT") || strings.Contains(content, "CURRENT - RESIDENTS") || strings.Contains(content, "CURRENT ACCOUNT") {
		accountType = models.AccountTypeCurrent
	}
	if strings.Contains(nameUpper, "5020") || strings.HasPrefix(filepath.Base(nameUpper), "5020") {
		accountType = models.AccountTypeCurrent
	}

	return confidence, "HDFC Bank", accountType
}

var (
	hdfcPDFDatePrefixRegex = regexp.MustCompile(`^(\d{2}[/.-]\d{2}[/.-]\d{2,4})`)
	hdfcPDFDateRangeRegex  = regexp.MustCompile(`(?i)(?:From|Statement\s*From)\s*:\s*(\d{1,2}/\d{2}/\d{2,4})\s*(?:To|TO)\s*:\s*(\d{1,2}/\d{2}/\d{2,4})`)
	hdfcPDFAccMaskRegex    = regexp.MustCompile(`(?i)(?:Account\s*No\.?|Account\s*number)\s*[:\s]*([0-9X*]{8,18})`)
	hdfc14DigitAccRegex    = regexp.MustCompile(`(?i)\b([0-9]{14}|\d{4}[X*]{6}\d{4})\b`)
	hdfcProductCodeRegex   = regexp.MustCompile(`(?i)Pr\.?\s*Code\s*:\s*(\d+)`)
	hdfcPDFAccTypeRegex    = regexp.MustCompile(`(?i)Account\s*Type\s*[:\s]*(CURRENT|SAVINGS)`)
	hdfcPDFCustIDRegex     = regexp.MustCompile(`(?i)Cust\s*ID\s*:\s*(\d{6,14})`)
	hdfcPDFIFSCRegex       = regexp.MustCompile(`(?i)(?:RTGS/NEFT\s*IFSC|IFSC)\s*:\s*([A-Z]{4}0[A-Z0-9]{6})`)
	hdfcPDFBranchRegex     = regexp.MustCompile(`(?i)Account\s*Branch\s*:\s*([A-Z0-9\s]+)`)
	hdfcCustNameRegex      = regexp.MustCompile(`(?i)^(?:M\s*r\s*\.?|M\s*r\s*s\s*\.?|M\s*s\s*\.?|D\s*r\s*\.?|MR|MRS|MS|DR|PROF)\.?\s*([A-Za-z\s]{3,50})$`)
)

func (p *HDFCSavingsPDFParser) Parse(r io.Reader, opts ParseOptions) ([]ParsedTransaction, StatementMeta, error) {
	rows, err := extractor.ExtractPDFPositionalRows(r, opts.Password)
	if err != nil {
		return nil, StatementMeta{}, fmt.Errorf("failed to extract PDF rows: %w", err)
	}

	meta := StatementMeta{
		BankName:        "HDFC Bank",
		AccountType:     models.AccountTypeSavings,
		StatementFormat: "PDF",
	}

	// 1. Determine horizontal scale
	maxX := 0.0
	for _, row := range rows {
		for _, el := range row.Elements {
			if el.X > maxX {
				maxX = el.X
			}
		}
	}
	scale := 1.0
	if maxX > 0 {
		scale = maxX / 550.0
	}

	var xDate, xNarr, xRef, xValDate, xWithdrawal, xDeposit, xBalance float64

	// 2. Extract metadata from header text
	for _, row := range rows {
		var rowParts []string
		for _, el := range row.Elements {
			rowParts = append(rowParts, el.S)
		}
		rowStr := strings.Join(rowParts, " ")
		rowUpper := strings.ToUpper(rowStr)

		if matches := hdfcPDFDateRangeRegex.FindStringSubmatch(rowStr); len(matches) > 2 && meta.StartDate == "" {
			meta.StartDate = extractor.NormalizeIndianDate(matches[1])
			meta.EndDate = extractor.NormalizeIndianDate(matches[2])
		}

		if matches := hdfcPDFAccTypeRegex.FindStringSubmatch(rowStr); len(matches) > 1 {
			if strings.EqualFold(matches[1], "CURRENT") {
				meta.AccountType = models.AccountTypeCurrent
			}
		}

		if matches := hdfcProductCodeRegex.FindStringSubmatch(rowStr); len(matches) > 1 {
			// Product Code 247 in HDFC is Regular Current Account
			if matches[1] == "247" {
				meta.AccountType = models.AccountTypeCurrent
			}
		}

		if matches := hdfcPDFAccMaskRegex.FindStringSubmatch(rowStr); len(matches) > 1 && meta.AccountNumberMask == "" {
			rawAcc := matches[1]
			if !strings.ContainsAny(rawAcc, "X*") && len(rawAcc) >= 10 {
				meta.AccountNumber = rawAcc
			}
			meta.AccountNumberMask = "XX" + rawAcc[max(0, len(rawAcc)-4):]
		}

		// Also scan individual elements for 14-digit account numbers (e.g. ": 12345678901234")
		if meta.AccountNumberMask == "" || meta.AccountNumber == "" {
			for _, el := range row.Elements {
				trimmed := strings.TrimSpace(el.S)
				trimmed = strings.TrimPrefix(trimmed, ":")
				trimmed = strings.TrimSpace(trimmed)
				if m := hdfc14DigitAccRegex.FindStringSubmatch(trimmed); len(m) > 1 {
					rawAcc := m[1]
					if !strings.ContainsAny(rawAcc, "X*") && len(rawAcc) >= 10 {
						meta.AccountNumber = rawAcc
					}
					if meta.AccountNumberMask == "" {
						meta.AccountNumberMask = "XX" + rawAcc[max(0, len(rawAcc)-4):]
					}
					break
				}
			}
		}

		// Scan Page 1 customer name block
		if row.Page == 1 && meta.AccountHolderName == "" {
			for _, el := range row.Elements {
				if el.X < 80.0*scale {
					trimmed := strings.TrimSpace(el.S)
					if m := hdfcCustNameRegex.FindStringSubmatch(trimmed); len(m) > 1 {
						meta.AccountHolderName = extractor.CleanCustomerName(m[1])
						break
					}
				}
			}
		}

		if matches := hdfcPDFCustIDRegex.FindStringSubmatch(rowStr); len(matches) > 1 && meta.CustomerID == "" {
			meta.CustomerID = matches[1]
		}
		if matches := hdfcPDFIFSCRegex.FindStringSubmatch(rowStr); len(matches) > 1 && meta.IFSCCode == "" {
			meta.IFSCCode = matches[1]
		}
		if matches := hdfcPDFBranchRegex.FindStringSubmatch(rowStr); len(matches) > 1 && meta.BranchName == "" {
			meta.BranchName = strings.TrimSpace(matches[1])
		}

		// Detect table header columns
		if (strings.Contains(rowUpper, "DATE") && strings.Contains(rowUpper, "NARRATION")) ||
			(strings.Contains(rowUpper, "WITHDRAWAL") && strings.Contains(rowUpper, "DEPOSIT")) {
			for _, el := range row.Elements {
				upper := strings.ToUpper(strings.TrimSpace(el.S))
				if strings.Contains(upper, "DATE") && !strings.Contains(upper, "VALUE") && xDate == 0 {
					xDate = el.X
				}
				if strings.Contains(upper, "NARRATION") && xNarr == 0 {
					xNarr = el.X
				}
				if (strings.Contains(upper, "CHQ") || strings.Contains(upper, "REF")) && xRef == 0 {
					xRef = el.X
				}
				if (strings.Contains(upper, "VALUE") || strings.Contains(upper, "VAL")) && xValDate == 0 {
					xValDate = el.X
				}
				if strings.Contains(upper, "WITHDRAWAL") && xWithdrawal == 0 {
					xWithdrawal = el.X
				}
				if strings.Contains(upper, "DEPOSIT") {
					if !strings.Contains(upper, "WITHDRAWAL") || xDeposit == 0 {
						xDeposit = el.X
					}
				}
				if strings.Contains(upper, "BALANCE") && xBalance == 0 {
					xBalance = el.X
				}
			}
		}
	}

	// Determine account type from account number prefix (HDFC 502/5020 is Current Account)
	if strings.HasPrefix(meta.AccountNumber, "5020") || strings.HasPrefix(meta.AccountNumber, "502") {
		meta.AccountType = models.AccountTypeCurrent
	}
	if opts.Filename != "" {
		fUpper := strings.ToUpper(opts.Filename)
		if strings.Contains(fUpper, "5020") || strings.Contains(fUpper, "CURRENT") {
			meta.AccountType = models.AccountTypeCurrent
		}
		if meta.AccountNumberMask == "" {
			baseName := filepath.Base(fUpper)
			if m := regexp.MustCompile(`^(\d{4}[0-9X*]{6,10}\d{4})`).FindStringSubmatch(baseName); len(m) > 1 {
				rawAcc := m[1]
				meta.AccountNumberMask = "XX" + rawAcc[len(rawAcc)-4:]
			}
		}
	}

	var thresholdAmount, thresholdDeposit, thresholdBalance, thresholdRefMin, thresholdRefMax float64

	if xWithdrawal > 0 && xBalance > xWithdrawal {
		if xDeposit > xWithdrawal {
			if xValDate > 0 && xValDate < xWithdrawal {
				thresholdAmount = (xValDate + xWithdrawal) / 2
			} else {
				thresholdAmount = xWithdrawal - (xDeposit-xWithdrawal)*0.5
			}
			thresholdDeposit = (xWithdrawal + xDeposit) / 2
			thresholdBalance = (xDeposit + xBalance) / 2
		} else {
			// Bundled withdrawal & deposit headers (e.g. compact format)
			mid := (xWithdrawal + xBalance) / 2
			if xValDate > 0 && xValDate < xWithdrawal {
				thresholdAmount = (xValDate + xWithdrawal) / 2
			} else {
				thresholdAmount = xWithdrawal * 0.85
			}
			thresholdDeposit = (xWithdrawal + mid) / 2
			thresholdBalance = (mid + xBalance) / 2
		}
	} else {
		// Fallback proportional to maxX
		thresholdAmount = 380.0 * scale
		thresholdDeposit = 460.0 * scale
		thresholdBalance = 505.0 * scale
	}

	if xRef > 0 && xNarr > 0 && xRef > xNarr {
		thresholdRefMin = (xNarr + xRef) / 2
	} else {
		thresholdRefMin = 230.0 * scale
	}

	if xRef > 0 && xValDate > 0 && xValDate > xRef {
		thresholdRefMax = (xRef + xValDate) / 2
	} else if xRef > 0 && xWithdrawal > xRef {
		thresholdRefMax = (xRef + xWithdrawal) / 2
	} else {
		thresholdRefMax = 360.0 * scale
	}

	dateXLimit := 80.0 * scale
	if xDate > 0 {
		dateXLimit = xDate + 50.0*scale
	}

	// 3. Parse transactions with page-by-page midpoint vertical isolation
	type pageData struct {
		page int
		rows []extractor.PositionalRow
	}
	var pages []pageData
	pageMap := make(map[int]int)
	for _, row := range rows {
		idx, exists := pageMap[row.Page]
		if !exists {
			idx = len(pages)
			pageMap[row.Page] = idx
			pages = append(pages, pageData{page: row.Page})
		}
		pages[idx].rows = append(pages[idx].rows, row)
	}

	var transactions []ParsedTransaction
	summaryReached := false

	for _, pData := range pages {
		if summaryReached {
			continue
		}

		// Check if summary starts on this page
		summaryY := -1.0
		for _, row := range pData.rows {
			rowText := ""
			for _, el := range row.Elements {
				rowText += el.S + " "
			}
			rowUpper := strings.ToUpper(rowText)

			if strings.Contains(rowUpper, "STATEMENTSUMMARY") || strings.Contains(rowUpper, "STATEMENT SUMMARY") {
				summaryReached = true
				summaryY = row.Y
				break
			}
		}

		if summaryReached {
			var summaryAmts []float64
			for _, row := range pData.rows {
				if row.Y <= summaryY {
					for _, el := range row.Elements {
						if amt, err := extractor.ParseIndianAmount(el.S); err == nil && amt >= 0 {
							summaryAmts = append(summaryAmts, amt)
						}
					}
				}
			}
			if len(summaryAmts) >= 6 {
				meta.OpeningBalance = summaryAmts[0]
				meta.TotalDebits = summaryAmts[3]
				meta.TotalCredits = summaryAmts[4]
				meta.ClosingBalance = summaryAmts[5]
			} else if len(summaryAmts) >= 4 {
				meta.OpeningBalance = summaryAmts[0]
				meta.TotalDebits = summaryAmts[len(summaryAmts)-3]
				meta.TotalCredits = summaryAmts[len(summaryAmts)-2]
				meta.ClosingBalance = summaryAmts[len(summaryAmts)-1]
			} else if len(summaryAmts) >= 2 {
				if meta.OpeningBalance == 0 {
					meta.OpeningBalance = summaryAmts[0]
				}
				meta.ClosingBalance = summaryAmts[len(summaryAmts)-1]
			}
		}

		// Find all date rows on this page (must be above summaryY if summary is on this page)
		type DateRowInfo struct {
			row extractor.PositionalRow
			y   float64
		}
		var dateRows []DateRowInfo

		for _, r := range pData.rows {
			if summaryY > 0 && r.Y <= summaryY {
				continue
			}
			if len(r.Elements) == 0 {
				continue
			}
			firstEl := r.Elements[0]
			firstTrim := strings.TrimSpace(firstEl.S)
			if firstEl.X < dateXLimit && hdfcPDFDatePrefixRegex.MatchString(firstTrim) {
				// Must also have at least one amount/balance element
				hasAmt := false
				for _, el := range r.Elements[1:] {
					if el.X >= thresholdAmount {
						if amt, err := extractor.ParseIndianAmount(el.S); err == nil && amt >= 0 {
							hasAmt = true
							break
						}
					}
				}
				if hasAmt {
					dateRows = append(dateRows, DateRowInfo{row: r, y: r.Y})
				}
			}
		}

		if len(dateRows) == 0 {
			continue
		}

		// Process each transaction on this page using midpoint boundaries
		for i, dr := range dateRows {
			var upperY, lowerY float64
			if i == 0 {
				if len(dateRows) > 1 {
					gap := (dr.y - dateRows[i+1].y) / 2.0
					if gap > 25.0*scale {
						gap = 25.0 * scale
					}
					upperY = dr.y + gap
				} else {
					upperY = dr.y + 25.0*scale
				}
			} else {
				upperY = (dateRows[i-1].y + dr.y) / 2.0
			}

			if i == len(dateRows)-1 {
				if len(dateRows) > 1 {
					gap := (dateRows[i-1].y - dr.y) / 2.0
					if gap > 25.0*scale {
						gap = 25.0 * scale
					}
					lowerY = dr.y - gap
				} else {
					lowerY = dr.y - 25.0*scale
				}
				if summaryY > 0 && lowerY < summaryY {
					lowerY = summaryY
				}
			} else {
				lowerY = (dr.y + dateRows[i+1].y) / 2.0
			}

			firstText := strings.TrimSpace(dr.row.Elements[0].S)
			matches := hdfcPDFDatePrefixRegex.FindStringSubmatch(firstText)
			normDate := extractor.NormalizeIndianDate(matches[1])

			var valDate *string
			refNo := ""
			amount := 0.0
			txType := models.TxTypeDebit
			var balance *float64

			remDateNarr := strings.TrimSpace(strings.TrimPrefix(firstText, matches[1]))

			for _, el := range dr.row.Elements[1:] {
				sTrim := strings.TrimSpace(el.S)
				if sTrim == "" {
					continue
				}

				// Combined RefNo and ValueDate (e.g. "0000127198376346 01/08/26")
				parts := strings.Fields(sTrim)
				if len(parts) == 2 && hdfcPDFDatePrefixRegex.MatchString(parts[1]) {
					refNo = strings.TrimLeft(parts[0], "0")
					if refNo == "" {
						refNo = "0"
					}
					vDate := extractor.NormalizeIndianDate(parts[1])
					valDate = &vDate
					continue
				}

				// Amount columns (Withdrawal, Deposit, Balance)
				if el.X >= thresholdAmount {
					if amt, err := extractor.ParseIndianAmount(sTrim); err == nil && amt >= 0 {
						if el.X >= thresholdBalance {
							balance = &amt
							continue
						}
						if amt > 0 {
							amount = amt
							if el.X >= thresholdDeposit {
								txType = models.TxTypeCredit
							} else {
								txType = models.TxTypeDebit
							}
							continue
						} else {
							continue
						}
					}
				}

				// Ref No column
				if el.X >= thresholdRefMin && el.X < thresholdRefMax {
					if len(sTrim) >= 3 && len(sTrim) <= 30 && refNo == "" && !hdfcPDFDatePrefixRegex.MatchString(sTrim) {
						refNo = strings.TrimLeft(sTrim, "0")
						if refNo == "" {
							refNo = "0"
						}
						continue
					}
				}

				// Value Date column
				if hdfcPDFDatePrefixRegex.MatchString(sTrim) && valDate == nil {
					vDate := extractor.NormalizeIndianDate(sTrim)
					valDate = &vDate
					continue
				}
			}

			// Gather narration tokens across all rows on this page in [lowerY, upperY]
			var narrTokens []string
			if remDateNarr != "" {
				narrTokens = append(narrTokens, remDateNarr)
			}

			for _, r := range pData.rows {
				if r.Y <= upperY && r.Y > lowerY {
					// Skip header/footer rows
					rowStr := ""
					for _, el := range r.Elements {
						rowStr += el.S + " "
					}
					rowUpper := strings.ToUpper(rowStr)
					if (strings.Contains(rowUpper, "DATE") && strings.Contains(rowUpper, "NARRATION")) ||
						strings.Contains(rowUpper, "STATEMENT FROM") || strings.Contains(rowUpper, "CLOSING BALANCE*") ||
						strings.Contains(rowUpper, "PAGENO") || strings.Contains(rowUpper, "HDFCBANKLIMITED") ||
						strings.Contains(rowUpper, "NOMINATION") || strings.Contains(rowUpper, "REGISTEREDOFFICE") ||
						strings.Contains(rowUpper, "GENERATED BY") || strings.Contains(rowUpper, "GENERATED ON") {
						continue
					}

					for _, el := range r.Elements {
						sTrim := strings.TrimSpace(el.S)
						if sTrim == "" {
							continue
						}
						// Skip Date column
						if el.X < dateXLimit && hdfcPDFDatePrefixRegex.MatchString(sTrim) {
							continue
						}
						// Skip Amounts & Balance
						if el.X >= thresholdAmount {
							continue
						}
						// Skip RefNo if already captured
						if el.X >= thresholdRefMin && el.X < thresholdRefMax {
							cleanRef := strings.TrimLeft(sTrim, "0")
							if cleanRef == refNo || (refNo == "0" && cleanRef == "") {
								continue
							}
						}
						// Skip ValueDate if already captured
						if valDate != nil && extractor.NormalizeIndianDate(sTrim) == *valDate {
							continue
						}

						narrTokens = append(narrTokens, sTrim)
					}
				}
			}

			rawNarr := strings.Join(narrTokens, " ")
			rawNarr = strings.Join(strings.Fields(rawNarr), " ")

			cleaned := CleanNarration(rawNarr)
			ref := refNo
			if cleaned.ReferenceNumber != "" && ref == "" {
				ref = cleaned.ReferenceNumber
			}

			if amount > 0 {
				pt := ParsedTransaction{
					Date:            normDate,
					ValueDate:       valDate,
					RawNarration:    rawNarr,
					CleanedPayee:    cleaned.CleanedPayee,
					PaymentMode:     cleaned.PaymentMode,
					ReferenceNumber: ref,
					TxType:          txType,
					Amount:          amount,
					RunningBalance:  balance,
					UPIVPA:          cleaned.UPIVPA,
					CardLast4:       cleaned.CardLast4,
					IsTransfer:      cleaned.IsTransfer,
				}
				transactions = append(transactions, pt)
			}
		}
	}

	// Reverse if transactions are in reverse chronological order
	if len(transactions) > 1 {
		firstDate := transactions[0].Date
		lastDate := transactions[len(transactions)-1].Date
		if firstDate > lastDate {
			for i, j := 0, len(transactions)-1; i < j; i, j = i+1, j-1 {
				transactions[i], transactions[j] = transactions[j], transactions[i]
			}
		}
	}

	if len(transactions) > 0 {
		if meta.StartDate == "" {
			meta.StartDate = transactions[0].Date
		}
		if meta.EndDate == "" {
			meta.EndDate = transactions[len(transactions)-1].Date
		}

		// If no summary block was found in the statement, calculate totals from transactions
		if meta.OpeningBalance == 0 && meta.ClosingBalance == 0 {
			var totDebits, totCredits float64
			for _, tx := range transactions {
				if tx.TxType == models.TxTypeDebit {
					totDebits += tx.Amount
				} else if tx.TxType == models.TxTypeCredit {
					totCredits += tx.Amount
				}
			}
			meta.TotalDebits = totDebits
			meta.TotalCredits = totCredits

			firstTx := transactions[0]
			if firstTx.RunningBalance != nil {
				if firstTx.TxType == models.TxTypeCredit {
					meta.OpeningBalance = *firstTx.RunningBalance - firstTx.Amount
				} else {
					meta.OpeningBalance = *firstTx.RunningBalance + firstTx.Amount
				}
			}
			lastTx := transactions[len(transactions)-1]
			if lastTx.RunningBalance != nil {
				meta.ClosingBalance = *lastTx.RunningBalance
			}
		}
	}

	return transactions, meta, nil
}
