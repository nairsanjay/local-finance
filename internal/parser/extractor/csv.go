package extractor

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// ExtractCSV reads any CSV, TSV, or delimited file and returns a 2D matrix of strings
func ExtractCSV(r io.Reader) ([][]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV stream: %w", err)
	}

	// Strip UTF-8 BOM if present
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}

	// Sniff delimiter
	delimiter := detectDelimiter(data)

	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = delimiter
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // Allow variable number of fields
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		// Fallback: Line-by-line manual split if strict CSV parser failed due to unescaped quotes
		lines := strings.Split(string(data), "\n")
		var fallbackRecords [][]string
		for _, line := range lines {
			line = strings.TrimRight(line, "\r")
			if strings.TrimSpace(line) == "" {
				continue
			}
			parts := strings.Split(line, string(delimiter))
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			fallbackRecords = append(fallbackRecords, parts)
		}
		if len(fallbackRecords) > 0 {
			return fallbackRecords, nil
		}
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	// Trim whitespace from all cells
	for rIdx := range records {
		for cIdx := range records[rIdx] {
			records[rIdx][cIdx] = strings.TrimSpace(records[rIdx][cIdx])
		}
	}

	return records, nil
}

func detectDelimiter(sample []byte) rune {
	lines := strings.Split(string(sample), "\n")
	candidateDelimiters := []rune{',', '\t', ';', '|'}

	maxLinesToCheck := 10
	if len(lines) < maxLinesToCheck {
		maxLinesToCheck = len(lines)
	}

	counts := make(map[rune]int)
	for i := 0; i < maxLinesToCheck; i++ {
		line := lines[i]
		for _, d := range candidateDelimiters {
			counts[d] += strings.Count(line, string(d))
		}
	}

	bestDelim := ','
	bestCount := -1
	for _, d := range candidateDelimiters {
		if counts[d] > bestCount {
			bestCount = counts[d]
			bestDelim = d
		}
	}

	if bestCount > 0 {
		return bestDelim
	}
	return ','
}

// IsValidUTF8 checks if sample bytes are valid UTF-8
func IsValidUTF8(sample []byte) bool {
	return utf8.Valid(sample)
}
