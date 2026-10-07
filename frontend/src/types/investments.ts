export interface InvestmentHolding {
  symbol: string
  isin: string
  asset_class: string
  quantity: number
  average_price: number
  closing_price: number
  invested_value: number
  current_value: number
  unrealized_return: number
  return_percent: number | null
  fields: Record<string, string>
}

export interface InvestmentFormats {
  parsers: { id: string; provider: string; name: string; extensions: string[] }[]
  max_file_size: number
}

export interface InvestmentSnapshot {
  id: string
  provider: string
  parser_id: string
  account_ref: string
  as_of: string
  currency: string
  filename: string
  imported_at: string
  invested_value: number
  current_value: number
  unrealized_return: number
  return_percent: number | null
  holdings: InvestmentHolding[]
  sheets: { name: string; rows: string[][] }[]
  warnings: string[]
}
