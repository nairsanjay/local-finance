import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { test } from 'node:test'
import ts from 'typescript'

const source = await readFile(new URL('../src/lib/spending.ts', import.meta.url), 'utf8')
const { outputText } = ts.transpileModule(source, {
  compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ES2022 },
})
const { validSpendingMonth, spendingMonthRange, spendingMonthLabel } = await import(
  `data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`
)

test('calendar spending ranges include the whole month, including leap days', () => {
  for (const [month, end] of [['2024-02', 29], ['2026-02', 28], ['1900-02', 28], ['2000-02', 29], ['2026-04', 30], ['2026-12', 31], ['0001-01', 31]]) {
    assert.deepEqual(spendingMonthRange(month), { start_date: `${month}-01`, end_date: `${month}-${end}` })
  }
  assert.equal(spendingMonthLabel('2026-08'), 'August 2026')
})

test('all-time and malformed URL months never produce invalid date requests', () => {
  for (const value of [undefined, null, '', 202608, '2026-00', '2026-13', '0000-01', '2026-8', '2026-08-01', '2026-08&category=food']) {
    assert.equal(validSpendingMonth(value), false)
    assert.deepEqual(spendingMonthRange(value), {})
    assert.equal(spendingMonthLabel(value), 'All time')
  }
})
