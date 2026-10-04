export interface ReviewSpend {
  amount: number
  count: number
}

export interface ReviewCategory {
  id: string
  name: string
  current: ReviewSpend
  previous: ReviewSpend
  delta: number
  merchants: { name: string; current: ReviewSpend; previous: ReviewSpend; delta: number }[]
}

export interface MonthlyReviewData {
  month: string
  next_month: string
  available_months: string[]
  current_period: { start: string; end: string }
  previous_period: { start: string; end: string }
  is_partial_month: boolean
  coverage_complete: boolean
  coverage: {
    account_id: string
    name: string
    latest_end: string
    current_days: number
    previous_days: number
    current_total: number
    previous_total: number
  }[]
  current: ReviewSpend
  previous: ReviewSpend
  delta: number
  categories: ReviewCategory[]
}

export interface ReviewEvidence {
  items: { id: string; date: string; payee: string; account: string; amount: number }[]
  total: number
  page: number
  page_size: number
  period: { start: string; end: string }
}
