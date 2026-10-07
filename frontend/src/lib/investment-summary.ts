import type { InvestmentSnapshot } from '@/types/investments'

export interface CurrencyPortfolioTotal {
  currency: string
  investedValue: number | null
  knownInvestedValue: number
  missingCosts: number
  currentValue: number | null
  knownCurrentValue: number
  missingValuations: number
  accounts: number
  holdings: number
  oldestDate: string
  newestDate: string
}

// Full snapshots replace earlier statements for the same provider/account/currency.
export function investmentTotals(snapshots: InvestmentSnapshot[]): CurrencyPortfolioTotal[] {
  // Imports use UTC RFC3339Nano. Pad fractions so same-millisecond revisions
  // retain their ordering rather than losing precision through Date.parse.
  const importedAt = (snapshot: InvestmentSnapshot) => snapshot.imported_at.replace(/(?:\.(\d+))?Z$/, (_, fraction: string | undefined) => `.${(fraction ?? '').padEnd(9, '0')}Z`)
  const latest = new Map<string, InvestmentSnapshot>()
  for (const snapshot of snapshots) {
    const key = JSON.stringify([snapshot.provider, snapshot.account_ref, snapshot.currency])
    const previous = latest.get(key)
    if (!previous || snapshot.as_of > previous.as_of ||
      (snapshot.as_of === previous.as_of && (importedAt(snapshot) > importedAt(previous) ||
        (importedAt(snapshot) === importedAt(previous) && snapshot.id > previous.id)))) latest.set(key, snapshot)
  }
  const totals = new Map<string, CurrencyPortfolioTotal>()
  for (const snapshot of latest.values()) {
    const total = totals.get(snapshot.currency) ?? {
      currency: snapshot.currency, investedValue: 0, knownInvestedValue: 0, missingCosts: 0, currentValue: 0, knownCurrentValue: 0,
      missingValuations: 0, accounts: 0, holdings: 0, oldestDate: snapshot.as_of, newestDate: snapshot.as_of,
    }
    if (snapshot.invested_value === null) total.missingCosts++
    else total.knownInvestedValue += snapshot.invested_value
    total.investedValue = total.missingCosts ? null : total.knownInvestedValue
    if (snapshot.current_value === null) total.missingValuations++
    else total.knownCurrentValue += snapshot.current_value
    total.currentValue = total.missingValuations ? null : total.knownCurrentValue
    total.accounts++
    total.holdings += snapshot.holdings.length
    if (snapshot.as_of < total.oldestDate) total.oldestDate = snapshot.as_of
    if (snapshot.as_of > total.newestDate) total.newestDate = snapshot.as_of
    totals.set(snapshot.currency, total)
  }
  return [...totals.values()].sort((a, b) => a.currency.localeCompare(b.currency))
}
