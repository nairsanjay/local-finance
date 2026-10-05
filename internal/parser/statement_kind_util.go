package parser

import "strings"

// Investment providers can share a bank's name and table headings. Only clear
// document titles identify this unsupported format; merchant names in bank
// narrations (NSE, CDSL, etc.) must never exclude a bank statement.
func isInvestmentStatement(text string) bool {
	cas, nse, transactionDetails := false, false, false
	for _, line := range strings.Split(text, "\n") {
		line = strings.ToUpper(strings.Join(strings.Fields(line), " "))
		line = strings.TrimSpace(strings.TrimPrefix(line, "%PDF"))
		for _, title := range []string{
			"CONSOLIDATED ACCOUNT STATEMENT", "CONSOLIDATED MUTUAL FUND STATEMENT",
			"DEMAT ACCOUNT STATEMENT", "CDSL CONSOLIDATED ACCOUNT STATEMENT",
			"NSDL CONSOLIDATED ACCOUNT STATEMENT",
		} {
			if line == title || strings.HasPrefix(line, title+" ") || strings.HasPrefix(line, title+":") {
				return true
			}
		}
		cas = cas || line == "CAS"
		nse = nse || line == "NSE" || line == "NATIONAL STOCK EXCHANGE"
		transactionDetails = transactionDetails || line == "TRANSACTION DETAILS"
	}
	return cas && nse && transactionDetails
}
