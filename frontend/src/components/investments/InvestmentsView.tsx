import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { UploadCloud, TrendingUp } from 'lucide-react'
import { fetchInvestments, importInvestment, deleteInvestment } from '@/lib/api'
import { usePrivacy } from '@/components/privacy-provider'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'

export function InvestmentsView() {
  const client = useQueryClient()
  const { isPrivacyMode, maskValue } = usePrivacy()
  const [selected, setSelected] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const [fileInputVersion, setFileInputVersion] = useState(0)
  const [search, setSearch] = useState('')
  const [asset, setAsset] = useState('ALL')
  const [sourceSheet, setSourceSheet] = useState('')
  const [showSource, setShowSource] = useState(false)
  const [message, setMessage] = useState('')
  const query = useQuery({ queryKey: ['investments'], queryFn: fetchInvestments })
  const snapshot = query.data?.find(s => s.id === selected) ?? query.data?.[0]
  const upload = useMutation({
    mutationFn: importInvestment,
    onSuccess: async result => {
      await client.invalidateQueries({ queryKey: ['investments'] })
      setSelected(result.snapshot.id)
      setMessage(result.duplicate ? 'This statement is already imported.' : 'Investment statement imported.')
      setFile(null)
      setFileInputVersion(version => version + 1)
    },
  })
  const remove = useMutation({
    mutationFn: deleteInvestment,
    onSuccess: async () => { await client.invalidateQueries({ queryKey: ['investments'] }); setSelected(''); setMessage('Snapshot deleted.') },
  })
  const holdings = useMemo(() => snapshot?.holdings.filter(h =>
    (asset === 'ALL' || h.asset_class === asset) && `${h.symbol} ${h.isin}`.toLowerCase().includes(search.toLowerCase())
  ) ?? [], [snapshot, asset, search])
  const money = (value: number) => isPrivacyMode ? '••••••' : new Intl.NumberFormat('en-IN', { style: 'currency', currency: snapshot?.currency ?? 'INR', maximumFractionDigits: 2 }).format(value)
  const price = (value: number) => isPrivacyMode ? '••••••' : new Intl.NumberFormat('en-IN', { style: 'currency', currency: snapshot?.currency ?? 'INR', maximumFractionDigits: 6 }).format(value)
  const percent = (value: number | null) => value === null ? 'Not available' : maskValue(`${value.toFixed(2)}%`)
  const sheet = snapshot?.sheets.find(s => s.name === sourceSheet) ?? snapshot?.sheets[0]
  const sourceRows = sheet?.rows.filter(row => row?.some(cell => cell.trim() !== '')) ?? []
  const error = upload.error ?? remove.error ?? query.error

  return <div className="space-y-6">
    <div><h1 className="text-2xl font-semibold flex items-center gap-2"><TrendingUp className="h-6 w-6" />Investments</h1>
      <p className="text-sm text-muted-foreground mt-1">Dated portfolio holdings and unrealized returns. Kept separate from income and expense totals.</p></div>
    <Card><CardHeader><CardTitle>Import investment statement</CardTitle></CardHeader><CardContent className="space-y-3">
      <p className="text-sm text-muted-foreground">Upload a Zerodha holdings Excel export. Equity, mutual funds, and statement details are stored locally.</p>
      <div className="flex flex-wrap items-end gap-3"><div className="space-y-2"><Label htmlFor="investment-file">Holdings statement (.xlsx, up to 10 MB)</Label>
        <Input key={fileInputVersion} id="investment-file" type="file" accept=".xlsx" onChange={e => { setFile(e.target.files?.[0] ?? null); setMessage(''); upload.reset() }} /></div>
        <Button disabled={!file || upload.isPending} onClick={() => file && upload.mutate(file)}><UploadCloud className="h-4 w-4 mr-2" />{upload.isPending ? 'Importing…' : 'Import'}</Button></div>
      {error && <p role="alert" className="text-sm text-destructive">{error.message}</p>}
      {message && <p role="status" className="text-sm">{message}</p>}
    </CardContent></Card>
    {query.isPending && <p>Loading investments…</p>}
    {!query.isPending && !query.error && !snapshot && <Card><CardContent className="py-12 text-center text-muted-foreground">Import your first holdings statement to see your portfolio.</CardContent></Card>}
    {snapshot && <>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="space-y-2"><Label>Portfolio snapshot</Label><Select value={snapshot.id} onValueChange={value => { setSelected(value ?? ''); setSourceSheet(''); setAsset('ALL') }}>
          <SelectTrigger className="w-full sm:w-[420px]" aria-label="Portfolio snapshot"><SelectValue>{snapshot.provider} · {maskValue(snapshot.account_ref)} · {snapshot.as_of}</SelectValue></SelectTrigger>
          <SelectContent>{query.data?.map(s => <SelectItem key={s.id} value={s.id}>{s.provider} · {maskValue(s.account_ref)} · {s.as_of} · {new Date(s.imported_at).toLocaleString()}</SelectItem>)}</SelectContent>
        </Select><p className="text-xs text-muted-foreground">Values as of {snapshot.as_of}, using statement closing prices. This is a full snapshot; earlier imports are not added to it.</p><p className="text-xs text-muted-foreground">{maskValue(snapshot.filename)} · Imported {new Date(snapshot.imported_at).toLocaleString()}</p></div>
        <Dialog><DialogTrigger render={<Button variant="outline" disabled={remove.isPending} />}>Delete snapshot</DialogTrigger>
          <DialogContent><DialogHeader><DialogTitle>Delete this investment snapshot?</DialogTitle><DialogDescription>This removes the imported holdings and statement details from LocalFinance. You can import the file again.</DialogDescription></DialogHeader>
            <DialogFooter><DialogClose render={<Button variant="outline" />}>Cancel</DialogClose><DialogClose render={<Button variant="destructive" />} onClick={() => remove.mutate(snapshot.id)}>Delete</DialogClose></DialogFooter></DialogContent>
        </Dialog>
      </div>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">{[
        ['Invested amount', money(snapshot.invested_value)], ['Current value', money(snapshot.current_value)],
        ['Unrealized return', money(snapshot.unrealized_return)], ['Return on cost', percent(snapshot.return_percent)],
      ].map(([label, value]) => <Card key={label}><CardHeader className="pb-2"><CardTitle className="text-sm text-muted-foreground">{label}</CardTitle></CardHeader><CardContent className="text-xl font-semibold">{value}</CardContent></Card>)}</div>
      <p className="text-xs text-muted-foreground">Returns exclude realized trades, dividends, charges, and cash flows not included in the statement. Annualized return and XIRR are unavailable from a holdings snapshot.</p>
      {snapshot.warnings.map(w => <p key={w} role="status" className="text-sm">{w}</p>)}
      <Card><CardHeader><CardTitle>Holdings ({holdings.length})</CardTitle><div className="flex flex-wrap gap-3 pt-2">
        <Input className="sm:max-w-xs" placeholder="Search symbol or ISIN" aria-label="Search holdings" value={search} onChange={e => setSearch(e.target.value)} />
        <Select value={asset} onValueChange={value => setAsset(value ?? 'ALL')}><SelectTrigger className="w-[180px]" aria-label="Asset class"><SelectValue>{asset === 'ALL' ? 'All asset classes' : asset}</SelectValue></SelectTrigger><SelectContent><SelectItem value="ALL">All asset classes</SelectItem>{[...new Set(snapshot.holdings.map(h => h.asset_class))].map(a => <SelectItem key={a} value={a}>{a}</SelectItem>)}</SelectContent></Select>
      </div></CardHeader><CardContent><Table><TableHeader><TableRow>{['Holding', 'Quantity', 'Average price', 'Closing price', 'Invested', 'Current value', 'Unrealized return', 'Return %'].map(h => <TableHead key={h} className="whitespace-nowrap">{h}</TableHead>)}</TableRow></TableHeader>
        <TableBody>{holdings.map(h => <TableRow key={h.isin}><TableCell className="min-w-[220px]"><div className="font-medium">{h.symbol}</div><div className="text-xs text-muted-foreground">{h.isin} · {h.asset_class}</div></TableCell>
          <TableCell>{maskValue(h.quantity.toLocaleString('en-IN', { maximumFractionDigits: 6 }))}</TableCell><TableCell>{price(h.average_price)}</TableCell><TableCell>{price(h.closing_price)}</TableCell><TableCell>{money(h.invested_value)}</TableCell><TableCell>{money(h.current_value)}</TableCell><TableCell className={h.unrealized_return < 0 ? 'text-destructive' : ''}>{money(h.unrealized_return)}</TableCell><TableCell>{percent(h.return_percent)}</TableCell></TableRow>)}
          {!holdings.length && <TableRow><TableCell colSpan={8} className="text-center">No holdings match your filters.</TableCell></TableRow>}
        </TableBody></Table></CardContent></Card>
      <Card><CardHeader className="flex flex-row items-center justify-between"><CardTitle>Statement details</CardTitle><Button variant="outline" onClick={() => setShowSource(!showSource)}>{showSource ? 'Hide details' : 'Show all fields'}</Button></CardHeader>
        {showSource && <CardContent className="space-y-3"><p className="text-xs text-muted-foreground">Original worksheet fields include sector, instrument type, available, long-term, discrepant, and pledged quantities. Repeated sheets are shown here without adding them to portfolio totals.</p>
          <Select value={sheet?.name ?? ''} onValueChange={value => setSourceSheet(value ?? '')}><SelectTrigger className="w-[200px]" aria-label="Statement worksheet"><SelectValue>{sheet?.name}</SelectValue></SelectTrigger><SelectContent>{snapshot.sheets.map(s => <SelectItem key={s.name} value={s.name}>{s.name}</SelectItem>)}</SelectContent></Select>
          <Table><TableBody>{sourceRows.map((row, i) => <TableRow key={i}>{row.map((value, j) => <TableCell key={j} className="whitespace-nowrap">{maskValue(value)}</TableCell>)}</TableRow>)}</TableBody></Table>
        </CardContent>}
      </Card>
    </>}
  </div>
}
