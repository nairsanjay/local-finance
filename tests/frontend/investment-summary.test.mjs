import { test } from 'node:test'
import assert from 'node:assert/strict'
import { investmentTotals } from '../../frontend/src/lib/investment-summary.ts'

const snapshot = (patch = {}) => ({
  id: 'demo', provider: 'Example', account_ref: 'DEMO-1', currency: 'INR',
  as_of: '2026-04-01', imported_at: '2026-04-02T00:00:00Z',
  invested_value: 80, current_value: 100, holdings: [{ symbol: 'DEMO' }], ...patch,
})

test('latest statement per account replaces history; currencies never combine', () => {
  const data = [snapshot(), snapshot({ id: 'new', as_of: '2026-05-01', current_value: 200, invested_value: 150 }),
    snapshot({ account_ref: 'DEMO-2', current_value: 40 }),
    snapshot({ currency: 'USD', current_value: 30, invested_value: 20 })]
  const totals = investmentTotals(data)
  assert.deepEqual(totals.map(t => [t.currency, t.currentValue, t.investedValue, t.accounts, t.holdings]), [['INR', 240, 230, 2, 2], ['USD', 30, 20, 1, 1]])
  assert.equal(totals[0].oldestDate, '2026-04-01')
  assert.equal(totals[0].newestDate, '2026-05-01')
  assert.deepEqual(investmentTotals([...data].reverse()), totals)
})

test('same-day revisions use latest import; providers distinguish accounts', () => {
  const totals = investmentTotals([snapshot({ imported_at: '2026-04-03T00:00:00Z', current_value: 250 }),
    snapshot(), snapshot({ provider: 'Other Example', current_value: 60 })])
  assert.equal(totals[0].currentValue, 310)
  assert.equal(totals[0].accounts, 2)
  const revisions = [snapshot({ imported_at: '2026-04-02T00:00:00.123Z', current_value: 20 }),
    snapshot({ imported_at: '2026-04-02T00:00:00.123001Z', current_value: 30 })]
  assert.equal(investmentTotals(revisions)[0].currentValue, 30)
  assert.equal(investmentTotals(revisions.reverse())[0].currentValue, 30)
})

test('missing valuations never become zero or a complete partial total', () => {
  const totals = investmentTotals([snapshot({ currency: 'USD', current_value: null }),
    snapshot({ currency: 'USD', account_ref: 'DEMO-2', current_value: 0 })])
  assert.equal(totals[0].currentValue, null)
  assert.equal(totals[0].knownCurrentValue, 0)
  assert.equal(totals[0].missingValuations, 1)
  assert.equal(totals[0].investedValue, 160)
  assert.equal(investmentTotals([snapshot({ current_value: 0 })])[0].currentValue, 0)
  assert.deepEqual(investmentTotals([]), [])
})

test('current-value-only reports keep acquisition costs unknown', () => {
  const total = investmentTotals([snapshot({ currency: 'USD', invested_value: null, current_value: 60 }),
    snapshot({ currency: 'USD', account_ref: 'DEMO-2', invested_value: 20, current_value: 25 })])[0]
  assert.equal(total.currentValue, 85)
  assert.equal(total.investedValue, null)
  assert.equal(total.knownInvestedValue, 20)
  assert.equal(total.missingCosts, 1)
  assert.equal(total.missingValuations, 0)
})
