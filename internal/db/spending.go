package db

import "fmt"

// These filters share one spending policy across totals, charts and reports.
// The alias is an internal SQL table alias, never user input.
func expenseCategoryFilter(alias string) string {
	if alias != "" {
		alias += "."
	}
	return fmt.Sprintf("(%stx_type != 'DEBIT' OR COALESCE(%scategory_id, '') != 'cat_transfers')", alias, alias)
}

func analyticsFilter(alias string) string {
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	return prefix + "is_transfer = 0 AND " + prefix + "is_excluded = 0 AND " + expenseCategoryFilter(alias)
}

func spendingFilter(alias string) string {
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	return prefix + "tx_type = 'DEBIT' AND " + analyticsFilter(alias)
}

func automaticTransferCategory(transfer, manual bool, category *string) *string {
	if transfer && !manual {
		id := "cat_transfers"
		return &id
	}
	return category
}

func (d *DB) categorizeDetectedTransfers() error {
	_, err := d.conn.Exec(`UPDATE transactions SET category_id = 'cat_transfers'
		WHERE is_transfer = 1 AND COALESCE(is_manual_category, 0) = 0
		AND COALESCE(category_id, '') != 'cat_transfers'`)
	return err
}
