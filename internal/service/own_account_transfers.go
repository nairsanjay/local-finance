package service

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"local-finance/internal/models"
	"local-finance/internal/parser"
)

// UPIAR statements carry the network RRN in this field even when their
// separate reference column is empty. Extract evidence without rewriting the
// stored reference (which participates in transaction identity).
var upiARReference = regexp.MustCompile(`(?i)(?:^|\s)UPIAR/(\d{12})/(?:DR|CR)/`)
var paymentReference = regexp.MustCompile(`^[A-Z0-9]{10,32}$`)
var referenceDigit = regexp.MustCompile(`[1-9]`)

func transferReference(t models.Transaction) string {
	ref := t.ReferenceNumber
	if strings.TrimSpace(ref) == "" {
		ref = parser.CleanNarration(t.RawNarration).ReferenceNumber
	}
	if strings.TrimSpace(ref) == "" {
		if match := upiARReference.FindStringSubmatch(t.RawNarration); len(match) == 2 {
			ref = match[1]
		}
	}
	ref = strings.ToUpper(strings.TrimSpace(ref))
	if !paymentReference.MatchString(ref) || !referenceDigit.MatchString(ref) {
		return ""
	}
	return ref
}

// ReconcileOwnAccountTransfers repairs historical imports as well as newly
// imported statements. Only an unambiguous shared payment reference can link
// bank accounts automatically; payee names and categories are not evidence.
func (s *ReconciliationService) ReconcileOwnAccountTransfers() (int, error) {
	pairs, err := s.getOwnAccountCandidates()
	if err != nil {
		return 0, err
	}
	count := 0
	for _, pair := range pairs {
		if pair.MatchConfidence < 1 {
			continue
		}
		if err := s.LinkPair(pair.DebitTx.ID, pair.CreditTx.ID, pair.MatchReason); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *ReconciliationService) getOwnAccountCandidates() ([]models.TransferPair, error) {
	// Filter by amount/date in SQLite instead of truncating imported history.
	rows, err := s.db.Query(`
		SELECT d.id, d.account_id, da.bank_name || ' (' || da.account_type || ')',
		 d.tx_date, COALESCE(d.raw_narration,''), COALESCE(d.cleaned_payee,''), d.amount, COALESCE(d.reference_number,''),
		 c.id, c.account_id, ca.bank_name || ' (' || ca.account_type || ')',
		 c.tx_date, COALESCE(c.raw_narration,''), COALESCE(c.cleaned_payee,''), c.amount, COALESCE(c.reference_number,''),
		 COALESCE(d.transfer_match_reason,'') = 'MANUALLY_UNLINKED' OR COALESCE(c.transfer_match_reason,'') = 'MANUALLY_UNLINKED'
		FROM transactions d
		JOIN transactions c ON c.account_id != d.account_id AND c.tx_type = 'CREDIT'
		 AND ROUND(c.amount * 100) = ROUND(d.amount * 100)
		 AND ABS(julianday(c.tx_date) - julianday(d.tx_date)) <= 3
		JOIN accounts da ON da.id = d.account_id
		JOIN accounts ca ON ca.id = c.account_id
		WHERE d.tx_type = 'DEBIT' AND d.amount > 0
		 AND da.account_type IN ('SAVINGS','CURRENT') AND ca.account_type IN ('SAVINGS','CURRENT')
		 AND da.currency = ca.currency
		 AND d.is_transfer = 0 AND c.is_transfer = 0 AND d.is_excluded = 0 AND c.is_excluded = 0
		 AND COALESCE(d.transfer_peer_id,'') = '' AND COALESCE(c.transfer_peer_id,'') = ''
		ORDER BY d.tx_date DESC, d.id, c.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pairs []models.TransferPair
	debitMatches, creditMatches := map[string]int{}, map[string]int{}
	for rows.Next() {
		var p models.TransferPair
		var manuallyUnlinked bool
		if err := rows.Scan(&p.DebitTx.ID, &p.DebitTx.AccountID, &p.DebitTx.AccountName,
			&p.DebitTx.TxDate, &p.DebitTx.RawNarration, &p.DebitTx.CleanedPayee, &p.DebitTx.Amount, &p.DebitTx.ReferenceNumber,
			&p.CreditTx.ID, &p.CreditTx.AccountID, &p.CreditTx.AccountName,
			&p.CreditTx.TxDate, &p.CreditTx.RawNarration, &p.CreditTx.CleanedPayee, &p.CreditTx.Amount, &p.CreditTx.ReferenceNumber, &manuallyUnlinked); err != nil {
			return nil, err
		}
		dRef, cRef := transferReference(p.DebitTx), transferReference(p.CreditTx)
		if dRef != "" && cRef != "" && dRef != cRef {
			continue
		}
		p.ID = fmt.Sprintf("cand_%s_%s", p.DebitTx.ID, p.CreditTx.ID)
		p.MatchConfidence = 0.65
		p.MatchReason = "OWN_ACCOUNT_REVIEW_REQUIRED"
		if dRef != "" && dRef == cRef {
			p.MatchConfidence = 1
			p.MatchReason = "OWN_ACCOUNT_REFERENCE_MATCH"
			debitMatches[p.DebitTx.ID]++
			creditMatches[p.CreditTx.ID]++
		}
		if manuallyUnlinked {
			p.MatchConfidence = 0.65
			p.MatchReason = "OWN_ACCOUNT_MANUAL_REVIEW"
		}
		dDate, dErr := parseDate(p.DebitTx.TxDate)
		cDate, cErr := parseDate(p.CreditTx.TxDate)
		if dErr != nil || cErr != nil {
			continue
		}
		p.DateDifferenceDays = int(math.Abs(cDate.Sub(dDate).Hours() / 24))
		p.AmountDifference = math.Abs(p.DebitTx.Amount - p.CreditTx.Amount)
		pairs = append(pairs, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range pairs {
		p := &pairs[i]
		if p.MatchConfidence == 1 && (debitMatches[p.DebitTx.ID] != 1 || creditMatches[p.CreditTx.ID] != 1) {
			p.MatchConfidence = 0.65
			p.MatchReason = "OWN_ACCOUNT_AMBIGUOUS_REFERENCE"
		}
	}
	return pairs, nil
}
