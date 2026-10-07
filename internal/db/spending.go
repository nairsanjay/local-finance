package db

import "fmt"

// Transfers belong to neither income nor expense totals, charts or reports.
// The alias is an internal SQL table alias, never user input.
func nonTransferCategoryFilter(alias string) string {
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
	return prefix + "tx_type = 'DEBIT' AND " + prefix + "is_transfer = 0 AND " + prefix + "is_excluded = 0 AND " + nonTransferCategoryFilter(alias)
}

func incomeFilter(alias string) string {
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	return prefix + "tx_type = 'CREDIT' AND " + prefix + "is_transfer = 0 AND " + prefix + "is_excluded = 0 AND " + nonTransferCategoryFilter(alias)
}
