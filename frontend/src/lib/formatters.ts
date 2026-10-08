// Investment currencies are displayed as reported; no conversion is applied.
export function formatMoney(value: number | null, currency: string, isPrivacyMode = false, maximumFractionDigits = 2): string {
  if (value === null) return 'Not provided'
  if (isPrivacyMode) return '••••••'
  return new Intl.NumberFormat('en-IN', { style: 'currency', currency, maximumFractionDigits }).format(value)
}
