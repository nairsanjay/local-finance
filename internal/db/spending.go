package db

import "fmt"

// These filters share one spending policy across totals, charts and reports.
// The alias is an internal SQL table alias, never user input.
func expenseCategoryFilter(alias string) string {
	if alias != "" {
		alias += "."
	}
	return fmt.Sprintf("COALESCE(%scategory_id, '') != 'cat_transfers'", alias)
}

func spendingFilter(alias string) string {
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	return prefix + "tx_type = 'DEBIT' AND " + prefix + "is_transfer = 0 AND " + prefix + "is_excluded = 0 AND " + expenseCategoryFilter(alias)
}
