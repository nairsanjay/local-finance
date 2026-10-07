package investment

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Shared workbook helpers keep adapters independent of other providers.
func columns(row []string) map[string]int {
	result := map[string]int{}
	for i, v := range row {
		if v = strings.TrimSpace(v); v != "" {
			result[v] = i
		}
	}
	return result
}

func cell(row []string, i int) string {
	if i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func number(v string) (float64, error) {
	n, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(v), ",", ""), 64)
	if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, fmt.Errorf("invalid numeric value %q", v)
	}
	return n, nil
}
