package parser

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"local-finance/internal/models"
	"local-finance/internal/parser/extractor"
)

type HDFCCCPDFParser struct{}

func init() {
	DefaultRegistry.Register(&HDFCCCPDFParser{})
}

func (p *HDFCCCPDFParser) ID() string {
	return "hdfc_cc_pdf_v1"
}

func (p *HDFCCCPDFParser) Name() string {
	return "HDFC Bank Credit Card (PDF)"
}

func (p *HDFCCCPDFParser) SupportedTypes() []StatementType {
	return []StatementType{TypeCreditCardPDF}
}

func (p *HDFCCCPDFParser) CanParse(filename string, sample []byte) (float64, string, models.AccountType) {
	nameUpper := strings.ToUpper(filename)
	content := strings.ToUpper(string(sample))

	isPDF := strings.HasSuffix(nameUpper, ".PDF") || (len(sample) > 4 && string(sample[:4]) == "%PDF")
	if !isPDF {
		return 0.0, "", models.AccountTypeCreditCard
	}

	confidence := 0.0
	if strings.Contains(nameUpper, "HDFC") || strings.Contains(content, "HDFC BANK") {
		confidence += 0.3
	}
	if strings.Contains(nameUpper, "REGALIA") || strings.Contains(nameUpper, "RUPAY") ||
		strings.Contains(nameUpper, "SWIGGY") || strings.Contains(nameUpper, "INFINIA") ||
		strings.Contains(nameUpper, "MILLENNIA") || strings.Contains(nameUpper, "TATA NEU") ||
		strings.Contains(nameUpper, "_CC") || strings.Contains(content, "CREDIT CARD") {
		confidence += 0.45
	}
	if strings.Contains(content, "TOTAL AMOUNT DUE") || strings.Contains(content, "DOMESTIC TRANSACTIONS") {
		confidence += 0.2
	}

	return confidence, "HDFC Bank", models.AccountTypeCreditCard
}

var (
	hdfcCCCardNoRegex   = regexp.MustCompile(`(?i)(?:Credit Card No\.?|Credit Card Number|Card Number|Card No\.?)\s*[:\s]*([0-9X*]{4}[\s0-9X*]{8,14}\d{4})`)
	hdfcCCStmtDateRegex = regexp.MustCompile(`(?i)Statement Date\s*[:\s]*(\d{1,2}\s+[A-Za-z]+,?\s+\d{4})`)
	hdfcCCBillingPeriod = regexp.MustCompile(`(?i)Billing Period\s*[:\s]*(\d{1,2}\s+[A-Za-z]+,?\s+\d{4})\s*-\s*(\d{1,2}\s+[A-Za-z]+,?\s+\d{4})`)
	hdfcCCDueDateRegex  = regexp.MustCompile(`(?i)DUE DATE\s*[:\s]*(\d{1,2}\s+[A-Za-z]+,?\s+\d{4})`)
	hdfcCCPointsRegex   = regexp.MustCompile(`(?i)Reward Points\s*[:\s]*([0-9,]+)`)
	hdfcCCNameRegex     = regexp.MustCompile(`(?i)(?:Customer\s+Name|Name)\s*[:\s]*([A-Za-z\s]{3,50})`)
	hdfcCCDateCellRegex = regexp.MustCompile(`(\d{2}/\d{2}/\d{4})`)
	hdfcCCAmtCellRegex    = regexp.MustCompile(`(?i)([+])?\s*[C₹]?\s*([0-9,]+\.\d{2})`)
	hdfcCCTimeCellRegex   = regexp.MustCompile(`^\s*\|?\s*\d{1,2}:\d{2}(?::\d{2})?\s*$`)
	hdfcCCTimePrefixRegex = regexp.MustCompile(`^\s*\|?\s*\d{1,2}:\d{2}(?::\d{2})?\s*`)
)

func (p *HDFCCCPDFParser) Parse(r io.Reader, opts ParseOptions) ([]ParsedTransaction, StatementMeta, error) {
	rows, err := extractor.ExtractPDFPositionalRows(r, opts.Password)
	if err != nil {
		return nil, StatementMeta{}, fmt.Errorf("failed to extract PDF rows: %w", err)
	}

	meta := StatementMeta{
		BankName:        "HDFC Bank",
		AccountType:     models.AccountTypeCreditCard,
		StatementFormat: "PDF",
	}

	// 1. First pass: extract bill metadata and summary amounts
	for _, row := range rows {
		var lineParts []string
		for _, el := range row.Elements {
			lineParts = append(lineParts, el.S)
		}
		lineStr := strings.Join(lineParts, " ")

		// Detect card variant and network
		upperLine := strings.ToUpper(lineStr)
		if strings.Contains(upperLine, "REGALIA") {
			meta.CardVariant = "Regalia"
			meta.CardNetwork = "VISA"
		} else if strings.Contains(upperLine, "RUPAY") {
			meta.CardVariant = "RuPay UPI"
			meta.CardNetwork = "RUPAY"
		} else if strings.Contains(upperLine, "SWIGGY") {
			meta.CardVariant = "Swiggy HDFC"
			meta.CardNetwork = "MASTERCARD"
		} else if strings.Contains(upperLine, "INFINIA") {
			meta.CardVariant = "Infinia"
			meta.CardNetwork = "VISA"
		} else if strings.Contains(upperLine, "MILLENNIA") {
			meta.CardVariant = "Millennia"
		}

		if matches := hdfcCCNameRegex.FindStringSubmatch(lineStr); len(matches) > 1 && meta.AccountHolderName == "" {
			meta.AccountHolderName = extractor.CleanCustomerName(matches[1])
		}
		if matches := hdfcCCCardNoRegex.FindStringSubmatch(lineStr); len(matches) > 1 && meta.AccountNumberMask == "" {
			rawCard := matches[1]
			meta.AccountNumberMask = "XX" + rawCard[max(0, len(rawCard)-4):]
		}
		if matches := hdfcCCStmtDateRegex.FindStringSubmatch(lineStr); len(matches) > 1 && meta.StatementDate == "" {
			meta.StatementDate = extractor.NormalizeIndianDate(matches[1])
		}
		if matches := hdfcCCBillingPeriod.FindStringSubmatch(lineStr); len(matches) > 2 {
			if meta.StartDate == "" {
				meta.StartDate = extractor.NormalizeIndianDate(matches[1])
			}
			if meta.EndDate == "" {
				meta.EndDate = extractor.NormalizeIndianDate(matches[2])
			}
		}
		if matches := hdfcCCDueDateRegex.FindStringSubmatch(lineStr); len(matches) > 1 && meta.PaymentDueDate == "" {
			meta.PaymentDueDate = extractor.NormalizeIndianDate(matches[1])
		}
		if matches := hdfcCCPointsRegex.FindStringSubmatch(lineStr); len(matches) > 1 && meta.RewardPointsBalance == 0 {
			if pts, err := extractor.ParseIndianAmount(matches[1]); err == nil {
				meta.RewardPointsBalance = pts
			}
		}

		// Spatial extraction for Name and Bill Summary on Page 1
		if row.Page == 1 {
			// Cardholder Name on Page 1 at Y ~ 735 - 750, X < 300
			if meta.AccountHolderName == "" && row.Y >= 735.0 && row.Y <= 755.0 {
				for _, el := range row.Elements {
					if el.X < 300.0 {
						sClean := strings.TrimSpace(el.S)
						sUp := strings.ToUpper(sClean)
						if sClean != "" && !strings.Contains(sUp, "CREDIT CARD") && !strings.Contains(sUp, "HSN") &&
							!strings.Contains(sUp, "HDFC") && !strings.Contains(sUp, "STATEMENT") &&
							!strings.Contains(sUp, "ACCOUNT") {
							meta.AccountHolderName = extractor.CleanCustomerName(sClean)
							break
						}
					}
				}
			}

			// Total Credit Limit at Y ~ 540
			if row.Y >= 530.0 && row.Y <= 545.0 {
				for _, el := range row.Elements {
					if el.X >= 60.0 && el.X <= 120.0 && meta.CreditLimit == 0 {
						if a, err := extractor.ParseIndianAmount(el.S); err == nil && a > 0 {
							meta.CreditLimit = a
						}
					}
					if el.X >= 180.0 && el.X <= 240.0 && meta.AvailableCreditLimit == 0 {
						if a, err := extractor.ParseIndianAmount(el.S); err == nil && a > 0 {
							meta.AvailableCreditLimit = a
						}
					}
				}
			}
			// Total Amount Due at Y ~ 590, X ~ 445
			if row.Y >= 580.0 && row.Y <= 605.0 {
				for _, el := range row.Elements {
					if el.X >= 430.0 && el.X <= 480.0 {
						if a, err := extractor.ParseIndianAmount(el.S); err == nil && a > 0 {
							meta.TotalDueAmount = a
						}
					}
				}
			}
			// Minimum Due at Y ~ 544, X ~ 445
			if row.Y >= 535.0 && row.Y <= 555.0 {
				for _, el := range row.Elements {
					if el.X >= 430.0 && el.X <= 480.0 {
						if a, err := extractor.ParseIndianAmount(el.S); err == nil && a > 0 {
							meta.MinimumDueAmount = a
						}
					} else if el.X >= 490.0 && meta.PaymentDueDate == "" {
						if dNorm := extractor.NormalizeIndianDate(el.S); dNorm != "" {
							meta.PaymentDueDate = dNorm
						}
					}
				}
			}
		}
	}

	// 2. Second pass: Parse Transactions page by page with spatial midpoint isolation
	pageRows := make(map[int][]extractor.PositionalRow)
	var maxPage int
	for _, r := range rows {
		if r.Page > maxPage {
			maxPage = r.Page
		}
		pageRows[r.Page] = append(pageRows[r.Page], r)
	}

	type ccDateRow struct {
		Y       float64
		Date    string
		Amount  float64
		IsCr    bool
		RemNarr string
	}

	var transactions []ParsedTransaction
	inTxSection := false

	for p := 1; p <= maxPage; p++ {
		pRows := pageRows[p]
		if len(pRows) == 0 {
			continue
		}

		// Check section boundaries on this page
		var pageTxRows []extractor.PositionalRow

		for _, r := range pRows {
			var lineParts []string
			for _, el := range r.Elements {
				lineParts = append(lineParts, el.S)
			}
			lineStr := strings.Join(lineParts, " ")
			lineUpper := strings.ToUpper(lineStr)

			if strings.Contains(lineUpper, "DOMESTIC TRANSACTIONS") || strings.Contains(lineUpper, "INTERNATIONAL TRANSACTIONS") {
				inTxSection = true
				continue
			}

			if strings.Contains(lineUpper, "SMART EMI LOAN SUMMARY") || strings.Contains(lineUpper, "REWARDS PROGRAM POINTS") ||
				strings.Contains(lineUpper, "GST SUMMARY") || strings.Contains(lineUpper, "IMPORTANT INFORMATION") ||
				strings.Contains(lineUpper, "ELIGIBLE FOR") || strings.Contains(lineUpper, "OFFERS ON YOUR CARD") ||
				strings.Contains(lineUpper, "TRANSACTIONS TOTAL AMOUNT") || strings.Contains(lineUpper, "CASH BACK SUMMARY") ||
				strings.Contains(lineUpper, "CASHBACK SUMMARY") || strings.Contains(lineUpper, "CONVERT TO EMI") ||
				strings.Contains(lineUpper, "PAST DUES") || strings.Contains(lineUpper, "PURCHASE INDICATOR") ||
				strings.Contains(lineUpper, "YOUR CARD CONTROL SETTING") || strings.Contains(lineUpper, "AUTOPAY DEBIT") {
				inTxSection = false
				continue
			}

			if !inTxSection {
				continue
			}

			// Skip table column header rows & pagination
			if strings.Contains(lineUpper, "DATE & TIME") || strings.Contains(lineUpper, "TRANSACTION DESCRIPTION") ||
				strings.Contains(lineUpper, "PAGE ") || strings.Contains(lineUpper, "STATEMENT") {
				continue
			}

			pageTxRows = append(pageTxRows, r)
		}

		if len(pageTxRows) == 0 {
			continue
		}

		// Find DateRow anchors in pageTxRows
		var dateRows []ccDateRow
		for _, r := range pageTxRows {
			var dStr string
			var amt float64
			var isCr bool
			var rem string

			for _, el := range r.Elements {
				sTrim := strings.TrimSpace(el.S)
				if sTrim == "" {
					continue
				}

				// Check Date at X < 200
				if dStr == "" && el.X < 200.0 {
					if m := hdfcCCDateCellRegex.FindStringSubmatch(sTrim); len(m) > 1 {
						dStr = extractor.NormalizeIndianDate(m[1])
						rPart := strings.TrimSpace(strings.TrimPrefix(sTrim, m[1]))
						rPart = hdfcCCTimePrefixRegex.ReplaceAllString(rPart, "")
						rPart = strings.TrimSpace(strings.TrimPrefix(rPart, "|"))
						if rPart != "" && !hdfcCCTimeCellRegex.MatchString(rPart) {
							rem = rPart
						}
					}
				}

				// Check Amount at X >= 495
				if el.X >= 495.0 {
					if m := hdfcCCAmtCellRegex.FindStringSubmatch(sTrim); len(m) > 2 {
						if a, err := extractor.ParseIndianAmount(m[2]); err == nil && a > 0 {
							amt = a
							if m[1] == "+" || strings.Contains(sTrim, "+") || strings.Contains(strings.ToUpper(sTrim), "CR") {
								isCr = true
							}
						}
					}
				}
			}

			if dStr != "" {
				dateRows = append(dateRows, ccDateRow{
					Y:       r.Y,
					Date:    dStr,
					Amount:  amt,
					IsCr:    isCr,
					RemNarr: rem,
				})
			}
		}

		if len(dateRows) == 0 {
			continue
		}

		// Sort DateRows by Y descending (top-to-bottom)
		sort.Slice(dateRows, func(i, j int) bool {
			return dateRows[i].Y > dateRows[j].Y
		})

		holderUpper := strings.ToUpper(meta.AccountHolderName)

		// For each transaction anchor, determine bounding box [lowerY, upperY]
		for i, dr := range dateRows {
			var upperY, lowerY float64
			if i == 0 {
				gap := 15.0
				if len(dateRows) > 1 {
					g := (dr.Y - dateRows[i+1].Y) / 2.0
					if g > 0 && g < gap {
						gap = g
					}
				}
				upperY = dr.Y + gap
			} else {
				upperY = (dateRows[i-1].Y + dr.Y) / 2.0
			}

			if i == len(dateRows)-1 {
				gap := 15.0
				if len(dateRows) > 1 {
					g := (dateRows[i-1].Y - dr.Y) / 2.0
					if g > 0 && g < gap {
						gap = g
					}
				}
				lowerY = dr.Y - gap
			} else {
				lowerY = (dr.Y + dateRows[i+1].Y) / 2.0
			}

			// Collect narration tokens and amount within [lowerY, upperY]
			type narrToken struct {
				Y float64
				X float64
				S string
			}
			var tokens []narrToken
			txAmt := dr.Amount
			txIsCr := dr.IsCr

			if dr.RemNarr != "" {
				tokens = append(tokens, narrToken{Y: dr.Y, X: 150.0, S: dr.RemNarr})
			}

			for _, r := range pageTxRows {
				if r.Y <= upperY && r.Y > lowerY {
					for _, el := range r.Elements {
						sTrim := strings.TrimSpace(el.S)
						if sTrim == "" {
							continue
						}

						// Check amount if not already found
						if el.X >= 495.0 {
							if m := hdfcCCAmtCellRegex.FindStringSubmatch(sTrim); len(m) > 2 {
								if a, err := extractor.ParseIndianAmount(m[2]); err == nil && a > 0 {
									if txAmt == 0 {
										txAmt = a
									}
									if m[1] == "+" || strings.Contains(sTrim, "+") || strings.Contains(strings.ToUpper(sTrim), "CR") {
										txIsCr = true
									}
								}
							}
							continue
						}

						// Narration candidate: X between 100 and 495
						if el.X >= 100.0 && el.X < 495.0 {
							if hdfcCCTimeCellRegex.MatchString(sTrim) || hdfcCCDateCellRegex.MatchString(sTrim) {
								continue
							}
							up := strings.ToUpper(sTrim)
							if up == "EMI" || up == "PI" || up == "L" || up == "CR" ||
								strings.Contains(up, "DOMESTIC TRANSACTIONS") ||
								strings.Contains(up, "INTERNATIONAL TRANSACTIONS") ||
								strings.Contains(up, "DATE & TIME") ||
								strings.Contains(up, "TRANSACTION DESCRIPTION") ||
								strings.Contains(up, "REWARDS PROGRAM") ||
								strings.Contains(up, "CARD NUMBER") {
								continue
							}
							if holderUpper != "" && strings.Contains(up, holderUpper) {
								continue
							}
							// Avoid adding redundant RemNarr
							if dr.RemNarr != "" && strings.Contains(sTrim, dr.RemNarr) && r.Y == dr.Y {
								continue
							}
							tokens = append(tokens, narrToken{Y: r.Y, X: el.X, S: sTrim})
						}
					}
				}
			}

			// Sort tokens: top-to-bottom (descending Y), then left-to-right (ascending X)
			sort.Slice(tokens, func(a, b int) bool {
				if tokens[a].Y != tokens[b].Y {
					return tokens[a].Y > tokens[b].Y
				}
				return tokens[a].X < tokens[b].X
			})

			var parts []string
			for _, tok := range tokens {
				parts = append(parts, tok.S)
			}
			fullNarr := strings.Join(parts, " ")
			fullNarr = strings.Join(strings.Fields(fullNarr), " ")

			cleaned := CleanNarration(fullNarr)
			txType := models.TxTypeDebit
			if txIsCr {
				txType = models.TxTypeCredit
			}

			tx := ParsedTransaction{
				Date:            dr.Date,
				Amount:          txAmt,
				TxType:          txType,
				RawNarration:    fullNarr,
				CleanedPayee:    cleaned.CleanedPayee,
				PaymentMode:     cleaned.PaymentMode,
				ReferenceNumber: cleaned.ReferenceNumber,
				UPIVPA:          cleaned.UPIVPA,
				CardLast4:       cleaned.CardLast4,
			}
			if tx.PaymentMode == models.PaymentModeOther {
				tx.PaymentMode = models.PaymentModeCardPOS
			}
			if txType == models.TxTypeCredit {
				tx.IsTransfer = true
			} else {
				tx.IsTransfer = cleaned.IsTransfer
			}

			if tx.Amount > 0 && tx.CleanedPayee != "" {
				transactions = append(transactions, tx)
			}
		}
	}

	if len(transactions) > 0 {
		if meta.StartDate == "" {
			meta.StartDate = transactions[len(transactions)-1].Date
		}
		if meta.EndDate == "" {
			meta.EndDate = transactions[0].Date
		}
		if meta.StartDate > meta.EndDate {
			meta.StartDate, meta.EndDate = meta.EndDate, meta.StartDate
		}
	}

	return transactions, meta, nil
}
