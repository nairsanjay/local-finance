"""Regenerate the fictional BIFF8 fixture with xlwt 1.3.0 (test-only tooling)."""
from pathlib import Path
import xlwt

book = xlwt.Workbook()
sheet = book.add_sheet('HOLDINGS_BOOK')
rows = {
    0: ['Account Details'],
    1: ['Broker Name', 'Example US Broker'],
    2: ['Broker Account', 'DEMO-US-01'],
    3: ['Holdings as on', '2026-04-01'],
    7: ['Stock Symbol', 'Holding Since', 'Quantity', 'Avg. Price ($)', 'Total Value ($)'],
    8: ['DEMO-A', '01 Mar 2026, 10:30 AM', 0.123456789, 100, 12.345679],
    9: ['DEMO-B', '02 Mar 2026, 10:30 AM', 2.5, 20, 50],
    19: ['Disclaimer:-'],
    20: ['INDmoney report information'],
}
for r, values in rows.items():
    for c, value in enumerate(values):
        sheet.write(r, c, value)
book.add_sheet('Report Notes').write(0, 0, 'Fictional test data only')
book.save(str(Path(__file__).with_name('indmoney-fictional.xls')))
