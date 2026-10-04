package extractor

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	amountCleanRegex = regexp.MustCompile(`(?i)[₹ÁC,\sINR\(\)]`)
	monthsMap        = map[string]string{
		"JAN": "01", "FEB": "02", "MAR": "03", "APR": "04",
		"MAY": "05", "JUN": "06", "JUL": "07", "AUG": "08",
		"SEP": "09", "OCT": "10", "NOV": "11", "DEC": "12",
		"JANUARY": "01", "FEBRUARY": "02", "MARCH": "03", "APRIL": "04",
		"JUNE": "06", "JULY": "07", "AUGUST": "08",
		"SEPTEMBER": "09", "OCTOBER": "10", "NOVEMBER": "11", "DECEMBER": "12",
	}
)

// ParseIndianAmount parses an amount string formatted in Indian currency notation (e.g., "1,50,000.00", "2000.00 Cr", "100.00 Dr", "(500.00)", "C5,92,000", "Á44,110.56")
func ParseIndianAmount(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" || s == "--" {
		return 0, nil
	}

	isNegative := strings.Contains(s, "(") && strings.Contains(s, ")") || strings.HasPrefix(s, "-")
	isCredit := strings.HasSuffix(strings.ToUpper(s), "CR") || strings.HasPrefix(strings.ToUpper(s), "CR")
	isDebit := strings.HasSuffix(strings.ToUpper(s), "DR") || strings.HasPrefix(strings.ToUpper(s), "DR")

	cleaned := strings.ToUpper(s)
	cleaned = strings.TrimPrefix(cleaned, "CR")
	cleaned = strings.TrimSuffix(cleaned, "CR")
	cleaned = strings.TrimPrefix(cleaned, "DR")
	cleaned = strings.TrimSuffix(cleaned, "DR")
	cleaned = strings.TrimPrefix(cleaned, "RS.")
	cleaned = strings.TrimPrefix(cleaned, "RS")
	cleaned = strings.TrimPrefix(cleaned, "C")
	cleaned = strings.TrimPrefix(cleaned, "Á")
	cleaned = strings.TrimSuffix(cleaned, "/-")
	cleaned = amountCleanRegex.ReplaceAllString(cleaned, "")
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" {
		return 0, nil
	}

	val, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q: %w", s, err)
	}

	if isNegative {
		val = -math.Abs(val)
	} else if isDebit {
		val = math.Abs(val)
	} else if isCredit {
		val = math.Abs(val)
	}

	return val, nil
}

// NormalizeIndianDate standardizes varied date representations (DD/MM/YYYY, DD-MM-YYYY, DD/MM/YY, DD-Mon-YYYY, Month DD, YYYY, etc.) into ISO 8601 YYYY-MM-DD
func NormalizeIndianDate(d string) string {
	d = strings.TrimSpace(d)
	if d == "" {
		return ""
	}

	// First try standard Go time parser layouts
	layouts := []string{
		"02/01/2006", "02-01-2006", "02.01.2006",
		"02/01/06", "02-01-06", "02.01.06",
		"2006-01-02", "2006/01/02",
		"02-Jan-2006", "02/Jan/2006", "02 Jan 2006", "02 Jan, 2006", "02 Jan,2006",
		"02-Jan-06", "02/Jan/06", "02 Jan 06",
		"02-January-2006", "02 January 2006", "02 January, 2006",
		"January 02, 2006", "January 2, 2006", "Jan 02, 2006", "Jan 2, 2006",
		"January 02 2006", "January 2 2006", "Jan 02 2006", "Jan 2 2006",
		"August 12, 2026", "August 30, 2026",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, d); err == nil {
			return t.Format("2006-01-02")
		}
	}

	// Clean separators and tokenize
	dClean := strings.ReplaceAll(d, ",", " ")
	parts := strings.FieldsFunc(dClean, func(r rune) bool {
		return r == '/' || r == '-' || r == '.' || r == ' '
	})

	if len(parts) == 3 {
		p0 := parts[0]
		p1 := parts[1]
		p2 := parts[2]

		var day, month, year string

		// Case 1: "YYYY-MM-DD"
		if len(p0) == 4 && isDigitsOnly(p0) {
			year = p0
			month = p1
			day = p2
		} else if numM, exists := monthsMap[strings.ToUpper(p0)]; exists {
			// Case 2: "Month DD YYYY" (e.g. "August 12 2026")
			month = numM
			day = p1
			year = p2
		} else {
			// Case 3: "DD MM YYYY" or "DD Mon YYYY"
			day = p0
			if numM, exists := monthsMap[strings.ToUpper(p1)]; exists {
				month = numM
			} else {
				month = p1
			}
			year = p2
		}

		if len(year) == 2 {
			year = "20" + year
		}

		dayInt, _ := strconv.Atoi(day)
		monthInt, _ := strconv.Atoi(month)
		yearInt, _ := strconv.Atoi(year)

		if dayInt > 0 && dayInt <= 31 && monthInt > 0 && monthInt <= 12 && yearInt >= 1990 && yearInt <= 2100 {
			return fmt.Sprintf("%04d-%02d-%02d", yearInt, monthInt, dayInt)
		}
	}

	return d
}

func isDigitsOnly(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// ColumnSpec defines matching patterns for a table column
type ColumnSpec struct {
	Name     string   // Internal canonical name (e.g. "date", "narration", "debit", "credit", "balance", "ref_no")
	Keywords []string // Case-insensitive keywords to match against header cells
}

// FindHeaderRow searches a 2D matrix of string rows and returns the header row index and mapped column indices
func FindHeaderRow(rows [][]string, requiredColumns []ColumnSpec) (int, map[string]int, bool) {
	for rowIdx, row := range rows {
		colMap := make(map[string]int)
		matchedCount := 0

		for cIdx, cell := range row {
			cellUpper := strings.ToUpper(strings.TrimSpace(cell))
			if cellUpper == "" {
				continue
			}

			for _, spec := range requiredColumns {
				if _, alreadyFound := colMap[spec.Name]; alreadyFound {
					continue
				}
				for _, kw := range spec.Keywords {
					if strings.Contains(cellUpper, strings.ToUpper(kw)) {
						colMap[spec.Name] = cIdx
						matchedCount++
						break
					}
				}
			}
		}

		// Check if all required columns or at least essential columns (date + narration + at least one amount col) matched
		hasDate := false
		hasNarration := false
		hasAmount := false

		if _, ok := colMap["date"]; ok {
			hasDate = true
		}
		if _, ok := colMap["narration"]; ok {
			hasNarration = true
		}
		if _, ok := colMap["debit"]; ok {
			hasAmount = true
		}
		if _, ok := colMap["credit"]; ok {
			hasAmount = true
		}
		if _, ok := colMap["withdrawal"]; ok {
			hasAmount = true
		}
		if _, ok := colMap["deposit"]; ok {
			hasAmount = true
		}
		if _, ok := colMap["amount"]; ok {
			hasAmount = true
		}

		if hasDate && (hasNarration || matchedCount >= 3) && hasAmount {
			return rowIdx, colMap, true
		}
	}

	return -1, nil, false
}

var titlePrefixRegex = regexp.MustCompile(`(?i)^(?:MR|MS|MRS|DR|PROF|SHRI|SMT)\.?\s+`)
var multipleSpaceRegex = regexp.MustCompile(`\s+`)

// CleanCustomerName strips honorifics and standardizes customer name
func CleanCustomerName(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = titlePrefixRegex.ReplaceAllString(raw, "")
	raw = strings.TrimSpace(raw)
	raw = multipleSpaceRegex.ReplaceAllString(raw, " ")
	return strings.ToUpper(raw)
}
