package parser

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"local-finance/internal/models"
	"local-finance/internal/parser/extractor"
)

type AxisCCPDFParser struct{}

func init() {
	DefaultRegistry.Register(&AxisCCPDFParser{})
}

func (p *AxisCCPDFParser) ID() string {
	return "axis_cc_pdf_v1"
}

func (p *AxisCCPDFParser) Name() string {
	return "Axis Bank Credit Card (PDF)"
}

func (p *AxisCCPDFParser) SupportedTypes() []StatementType {
	return []StatementType{TypeCreditCardPDF}
}

func (p *AxisCCPDFParser) CanParse(filename string, sample []byte) (float64, string, models.AccountType) {
	nameUpper := strings.ToUpper(filename)
	content := strings.ToUpper(string(sample))

	isPDF := strings.HasSuffix(nameUpper, ".PDF") || (len(sample) > 4 && string(sample[:4]) == "%PDF")
	if !isPDF {
		return 0.0, "", models.AccountTypeCreditCard
	}

	confidence := 0.0
	if strings.Contains(nameUpper, "AXIS") || strings.Contains(content, "AXIS BANK") {
		confidence += 0.4
	}
	if strings.Contains(nameUpper, "FLIPKART") || strings.Contains(nameUpper, "ACE") ||
		strings.Contains(nameUpper, "MAGNUS") || strings.Contains(nameUpper, "ATLAS") ||
		strings.Contains(content, "TOTAL PAYMENT DUE") || strings.Contains(content, "MERCHANT CATEGORY") {
		confidence += 0.45
	}

	return confidence, "Axis Bank", models.AccountTypeCreditCard
}

var (
	axisCCCardNoRegex = regexp.MustCompile(`(?i)(?:Card No\.?|Credit Card Number)\s*[:\s]*([0-9X*]{4}[\s0-9X*]{8,14}\d{4})`)
	axisCCStmtPeriod  = regexp.MustCompile(`(?i)Statement Period\s*[:\s]*(\d{2}/\d{2}/\d{4})\s*-\s*(\d{2}/\d{2}/\d{4})`)
	axisCCDueDateReg  = regexp.MustCompile(`(?i)Payment Due Date\s*[:\s]*(\d{2}/\d{2}/\d{4})`)
	axisCCStmtDateReg = regexp.MustCompile(`(?i)Statement Generation Date\s*[:\s]*(\d{2}/\d{2}/\d{4})`)
	axisCCTotDueReg   = regexp.MustCompile(`(?i)Total Payment Due\s*[:\s]*([0-9,]+\.\d{2})`)
	axisCCMinDueReg   = regexp.MustCompile(`(?i)Minimum Payment Due\s*[:\s]*([0-9,]+\.\d{2})`)
	axisCCNameRegex   = regexp.MustCompile(`(?i)(?:Customer\s+Name|Name)\s*[:\s]*([A-Za-z\s]{3,50})`)
	axisCCLineDateReg = regexp.MustCompile(`^(\d{2}/\d{2}/\d{4})`)
	axisCCAmountDrCr  = regexp.MustCompile(`(?i)([0-9,]+\.\d{2})\s*(Dr|Cr)`)
)

func (p *AxisCCPDFParser) Parse(r io.Reader, opts ParseOptions) ([]ParsedTransaction, StatementMeta, error) {
	rows, err := extractor.ExtractPDFPositionalRows(r, opts.Password)
	if err != nil {
		return nil, StatementMeta{}, fmt.Errorf("failed to extract PDF rows: %w", err)
	}

	meta := StatementMeta{
		BankName:        "Axis Bank",
		AccountType:     models.AccountTypeCreditCard,
		CardNetwork:     "MASTERCARD",
		CardVariant:     "Flipkart Axis",
		StatementFormat: "PDF",
	}

	// 1. Scan metadata
	for _, row := range rows {
		var lineParts []string
		for _, el := range row.Elements {
			lineParts = append(lineParts, el.S)
		}
		lineStr := strings.Join(lineParts, " ")

		if matches := axisCCNameRegex.FindStringSubmatch(lineStr); len(matches) > 1 && meta.AccountHolderName == "" {
			meta.AccountHolderName = extractor.CleanCustomerName(matches[1])
		}
		if matches := axisCCCardNoRegex.FindStringSubmatch(lineStr); len(matches) > 1 && meta.AccountNumberMask == "" {
			rawCard := matches[1]
			meta.AccountNumberMask = "XX" + rawCard[max(0, len(rawCard)-4):]
		}
		if matches := axisCCStmtPeriod.FindStringSubmatch(lineStr); len(matches) > 2 {
			if meta.StartDate == "" {
				meta.StartDate = extractor.NormalizeIndianDate(matches[1])
			}
			if meta.EndDate == "" {
				meta.EndDate = extractor.NormalizeIndianDate(matches[2])
			}
		}
		if matches := axisCCDueDateReg.FindStringSubmatch(lineStr); len(matches) > 1 && meta.PaymentDueDate == "" {
			meta.PaymentDueDate = extractor.NormalizeIndianDate(matches[1])
		}
		if matches := axisCCStmtDateReg.FindStringSubmatch(lineStr); len(matches) > 1 && meta.StatementDate == "" {
			meta.StatementDate = extractor.NormalizeIndianDate(matches[1])
		}
		if matches := axisCCTotDueReg.FindStringSubmatch(lineStr); len(matches) > 1 && meta.TotalDueAmount == 0 {
			if amt, err := extractor.ParseIndianAmount(matches[1]); err == nil {
				meta.TotalDueAmount = amt
			}
		}
		if matches := axisCCMinDueReg.FindStringSubmatch(lineStr); len(matches) > 1 && meta.MinimumDueAmount == 0 {
			if amt, err := extractor.ParseIndianAmount(matches[1]); err == nil {
				meta.MinimumDueAmount = amt
			}
		}

		// Spatial check on Page 1 summary row (Y ~ 842)
		if row.Page == 1 {
			if row.Y >= 835.0 && row.Y <= 850.0 {
				for _, el := range row.Elements {
					if el.X >= 50.0 && el.X <= 100.0 && meta.TotalDueAmount == 0 {
						if m := axisCCAmountDrCr.FindStringSubmatch(el.S); len(m) > 1 {
							if a, err := extractor.ParseIndianAmount(m[1]); err == nil {
								meta.TotalDueAmount = a
							}
						}
					}
					if el.X >= 150.0 && el.X <= 200.0 && meta.MinimumDueAmount == 0 {
						if m := axisCCAmountDrCr.FindStringSubmatch(el.S); len(m) > 1 {
							if a, err := extractor.ParseIndianAmount(m[1]); err == nil {
								meta.MinimumDueAmount = a
							}
						}
					}
					if el.X >= 240.0 && el.X <= 300.0 && strings.Contains(el.S, " - ") {
						pParts := strings.Split(el.S, " - ")
						if len(pParts) == 2 {
							meta.StartDate = extractor.NormalizeIndianDate(pParts[0])
							meta.EndDate = extractor.NormalizeIndianDate(pParts[1])
						}
					}
					if el.X >= 380.0 && el.X <= 430.0 && meta.PaymentDueDate == "" {
						meta.PaymentDueDate = extractor.NormalizeIndianDate(el.S)
					}
					if el.X >= 480.0 && el.X <= 530.0 && meta.StatementDate == "" {
						meta.StatementDate = extractor.NormalizeIndianDate(el.S)
					}
				}
			}
			// Credit Limit at Y ~ 813
			if row.Y >= 805.0 && row.Y <= 820.0 {
				for _, el := range row.Elements {
					if el.X >= 150.0 && el.X <= 220.0 && meta.CreditLimit == 0 {
						if a, err := extractor.ParseIndianAmount(el.S); err == nil && a > 0 {
							meta.CreditLimit = a
						}
					}
					if el.X >= 260.0 && el.X <= 330.0 && meta.AvailableCreditLimit == 0 {
						if a, err := extractor.ParseIndianAmount(el.S); err == nil && a > 0 {
							meta.AvailableCreditLimit = a
						}
					}
				}
			}
			// Cashback Summary at Y ~ 411
			if row.Y >= 405.0 && row.Y <= 420.0 {
				for _, el := range row.Elements {
					if el.X >= 130.0 && el.X <= 200.0 && meta.CashbackEarned == 0 {
						if a, err := extractor.ParseIndianAmount(el.S); err == nil {
							meta.CashbackEarned = a
						}
					}
					if el.X >= 400.0 && el.X <= 470.0 && meta.CashbackCredited == 0 {
						if a, err := extractor.ParseIndianAmount(el.S); err == nil {
							meta.CashbackCredited = a
						}
					}
				}
			}
		}
	}

	// 2. Parse transactions
	var transactions []ParsedTransaction
	inTxSection := false

	for _, row := range rows {
		var lineParts []string
		for _, el := range row.Elements {
			lineParts = append(lineParts, el.S)
		}
		lineStr := strings.Join(lineParts, " ")
		lineUpper := strings.ToUpper(lineStr)

		if strings.Contains(lineUpper, "TRANSACTION DETAILS") {
			inTxSection = true
			continue
		}

		if strings.Contains(lineUpper, "END OF STATEMENT") || strings.Contains(lineUpper, "CASHBACK DETAILS") {
			inTxSection = false
			continue
		}

		if !inTxSection {
			continue
		}

		if matches := axisCCLineDateReg.FindStringSubmatch(lineStr); len(matches) > 1 {
			rawDate := matches[1]
			normDate := extractor.NormalizeIndianDate(rawDate)
			if normDate == "" {
				continue
			}

			rem := strings.TrimSpace(strings.TrimPrefix(lineStr, rawDate))

			amtMatches := axisCCAmountDrCr.FindAllStringSubmatch(rem, -1)
			if len(amtMatches) == 0 {
				continue
			}

			txAmtMatch := amtMatches[0]
			amount, err := extractor.ParseIndianAmount(txAmtMatch[1])
			if err != nil || amount == 0 {
				continue
			}

			var cashbackAmt float64
			if len(amtMatches) > 1 {
				if cb, err := extractor.ParseIndianAmount(amtMatches[1][1]); err == nil {
					cashbackAmt = cb
				}
			}

			txType := models.TxTypeDebit
			isTransfer := false
			if strings.EqualFold(txAmtMatch[2], "Cr") {
				txType = models.TxTypeCredit
				isTransfer = true
			}

			amtIndex := strings.Index(rem, txAmtMatch[0])
			narration := rem
			if amtIndex > 0 {
				narration = strings.TrimSpace(rem[:amtIndex])
			}

			// Extract Merchant Category from narration (e.g. "FLIPKART PAYMENTS,GURGAON MISC STORE")
			merchantCategory := ""
			for _, catKeyword := range []string{"MISC STORE", "UTILITIES", "DINING", "GROCERY", "APPAREL", "FUEL", "TRAVEL", "HEALTH"} {
				if strings.Contains(strings.ToUpper(narration), catKeyword) {
					merchantCategory = catKeyword
					narration = strings.TrimSpace(strings.ReplaceAll(narration, catKeyword, ""))
					break
				}
			}

			if strings.Contains(strings.ToUpper(narration), "PAYMENT RECEIVED") {
				txType = models.TxTypeCredit
				isTransfer = true
			}

			cleaned := CleanNarration(narration)
			if txType == models.TxTypeCredit {
				cleaned.IsTransfer = true
			}

			transactions = append(transactions, ParsedTransaction{
				Date:             normDate,
				RawNarration:     narration,
				CleanedPayee:     cleaned.CleanedPayee,
				PaymentMode:      models.PaymentModeCardPOS,
				ReferenceNumber:  cleaned.ReferenceNumber,
				TxType:           txType,
				Amount:           amount,
				MerchantCategory: merchantCategory,
				CashbackAmount:   cashbackAmt,
				UPIVPA:           cleaned.UPIVPA,
				CardLast4:        cleaned.CardLast4,
				IsTransfer:       cleaned.IsTransfer || isTransfer,
			})
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
