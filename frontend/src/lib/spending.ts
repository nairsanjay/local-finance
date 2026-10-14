export interface SpendingSearch {
  month?: string
  category?: string
}

export function validSpendingMonth(value: unknown): value is string {
  return (
    typeof value === 'string' && /^(?!0000)\d{4}-(0[1-9]|1[0-2])$/.test(value)
  )
}

export function spendingMonthRange(month?: string): {
  start_date?: string
  end_date?: string
} {
  if (!validSpendingMonth(month)) return {}
  const [year, number] = month.split('-').map(Number)
  const leap = year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0)
  const days = [31, leap ? 29 : 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31][
    number - 1
  ]
  return { start_date: `${month}-01`, end_date: `${month}-${days}` }
}

export function spendingMonthLabel(month?: string): string {
  if (!validSpendingMonth(month)) return 'All time'
  const date = new Date(`${month}-01T12:00:00`)
  return date.toLocaleDateString('en-IN', { month: 'long', year: 'numeric' })
}
