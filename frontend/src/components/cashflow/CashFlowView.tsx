import React, { useState, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useSearch, useNavigate } from '@tanstack/react-router'
import { fetchCashFlowIntelligence, fetchTransactions } from '@/lib/api'
import { SankeyDiagram } from './SankeyDiagram'
import { AnomaliesSection } from './AnomaliesSection'
import { CategoryComparisonTable } from './CategoryComparisonTable'
import { MonthlyReview } from './MonthlyReview'
import { formatINR, formatDate } from '@/lib/utils'
import { SankeyNode, Transaction, CashFlowSearchParams } from '@/types'
import {
  TrendingUp,
  TrendingDown,
  PiggyBank,
  PieChart,
  Calendar,
  Layers,
  RefreshCw,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

export const CashFlowView: React.FC = () => {
  const searchParams = (useSearch({ strict: false }) as CashFlowSearchParams) || {}
  const navigate = useNavigate()
  const selectedPeriod = searchParams.month || searchParams.period || ''

  const handlePeriodChange = useCallback(
    (newPeriod: string) => {
      navigate({
        to: '/cashflow',
        search: newPeriod ? { month: newPeriod } : {},
        replace: false,
      })
    },
    [navigate]
  )

  const [activeNode, setActiveNode] = useState<SankeyNode | null>(null)
  const [isDrilldownOpen, setIsDrilldownOpen] = useState(false)

  // 1. Fetch Cash Flow Intelligence Data
  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['cashflow-intelligence', selectedPeriod],
    queryFn: () => fetchCashFlowIntelligence(selectedPeriod || undefined),
  })

  // 2. Drilldown Query for Selected Node
  const { data: nodeTxs, isLoading: isNodeTxsLoading } = useQuery({
    queryKey: ['node-transactions', activeNode?.id, data?.selected_month],
    queryFn: async () => {
      if (!activeNode) return []
      const month = data?.selected_month
      const startDate = month && month.length === 7 ? `${month}-01` : undefined
      const endDate = month && month.length === 7 ? `${month}-31` : undefined

      let typeFilter: string | undefined = undefined
      let catFilter: string | undefined = undefined

      if (activeNode.type === 'INCOME_SOURCE') {
        typeFilter = 'CREDIT'
      } else if (activeNode.type === 'CATEGORY') {
        typeFilter = 'DEBIT'
        catFilter = activeNode.id.replace('cat_', '')
      }

      const res = await fetchTransactions({
        start_date: startDate,
        end_date: endDate,
        tx_type: typeFilter,
        category_id: catFilter,
        page_size: 50,
      })
      return res.items || []
    },
    enabled: !!activeNode && isDrilldownOpen,
  })

  const handleNodeClick = (node: SankeyNode) => {
    setActiveNode(node)
    setIsDrilldownOpen(true)
  }

  const handleCategoryRowClick = (categoryId: string) => {
    const node = data?.sankey.nodes.find((n) => n.id === `cat_${categoryId}`)
    if (node) {
      setActiveNode(node)
    } else {
      setActiveNode({
        id: `cat_${categoryId}`,
        name: categoryId,
        type: 'CATEGORY',
        color_hex: '#3B82F6',
        total_value: 0,
        layer: 3,
      })
    }
    setIsDrilldownOpen(true)
  }

  if (isLoading) {
    return (
      <div className="flex h-96 flex-col items-center justify-center gap-3 text-muted-foreground text-sm">
        <RefreshCw className="h-6 w-6 animate-spin text-primary" />
        <p>Computing cash flow routing &amp; MoM anomalies...</p>
      </div>
    )
  }

  const summary = data?.summary || {
    total_inflow: 0,
    total_outflow: 0,
    net_surplus: 0,
    savings_rate: 0,
    fixed_needs_spend: 0,
    discretionary_spend: 0,
    credit_card_share_pct: 0,
    direct_bank_share_pct: 0,
    top_spending_category: '',
    top_spending_amount: 0,
  }

  const currentSelectedMonth = data?.selected_month || ''
  const availableMonths = [...new Set([
    ...(/^\d{4}-\d{2}$/.test(selectedPeriod || currentSelectedMonth) ? [selectedPeriod || currentSelectedMonth] : []),
    ...(data?.available_months || []),
  ])].sort().reverse()
  const previousMonth = data?.previous_month || ''

  return (
    <div className="space-y-8">
      {/* Header & Month Filter Controls */}
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl flex items-center gap-2.5">
            <Layers className="h-6 w-6 text-primary" />
            Cash Flow Sankey &amp; MoM Intelligence
          </h1>
          <p className="text-xs text-muted-foreground mt-1">
            Visual multi-tier money routing, credit card channel flow, and month-over-month shift detection
          </p>
        </div>

        {/* Period Selector Bar */}
        <div className="flex items-center gap-2 bg-card/60 border rounded-lg p-1 shadow-xs">
          <Calendar className="h-3.5 w-3.5 text-muted-foreground ml-2" />
          <select
            value={selectedPeriod || currentSelectedMonth}
            onChange={(e) => handlePeriodChange(e.target.value)}
            className="bg-transparent text-xs font-semibold text-foreground focus:outline-hidden pr-2 cursor-pointer"
          >
            {availableMonths.map((m) => {
              const d = new Date(`${m}-01`)
              const monthLabel = d.toLocaleDateString('en-IN', {
                month: 'long',
                year: 'numeric',
              })
              return (
                <option key={m} value={m} className="bg-popover text-popover-foreground">
                  {monthLabel} ({m})
                </option>
              )
            })}
            <option value="LAST_3_MONTHS" className="bg-popover text-popover-foreground">
              Last 3 Months Rolling
            </option>
            <option value="ALL" className="bg-popover text-popover-foreground">
              All Time History
            </option>
          </select>

          <Button
            size="sm"
            variant="ghost"
            onClick={() => refetch()}
            disabled={isFetching}
            className="h-7 w-7 p-0"
            title="Refresh Data"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${isFetching ? 'animate-spin text-primary' : 'text-muted-foreground'}`} />
          </Button>
        </div>
      </div>

      {(!selectedPeriod || /^\d{4}-\d{2}$/.test(selectedPeriod)) && (
        <MonthlyReview month={selectedPeriod || undefined} onMonthChange={handlePeriodChange} />
      )}

      {/* 4 Top KPI Metric Cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {/* Total Inflow */}
        <div className="rounded-xl border bg-card/70 p-4 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-muted-foreground">Total Inflow</span>
            <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-500">
              <TrendingUp className="h-4 w-4" />
            </div>
          </div>
          <div className="mt-2 flex items-baseline gap-2">
            <span className="text-2xl font-bold tracking-tight text-foreground">
              {formatINR(summary.total_inflow)}
            </span>
          </div>
          <p className="text-[11px] text-muted-foreground mt-1">
            Salary, refunds &amp; incoming deposits
          </p>
        </div>

        {/* Total Outflow */}
        <div className="rounded-xl border bg-card/70 p-4 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-muted-foreground">Total Outflow</span>
            <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-rose-500/10 text-rose-500">
              <TrendingDown className="h-4 w-4" />
            </div>
          </div>
          <div className="mt-2 flex items-baseline gap-2">
            <span className="text-2xl font-bold tracking-tight text-foreground">
              {formatINR(summary.total_outflow)}
            </span>
          </div>
          <div className="mt-1 flex items-center gap-2 text-[11px] text-muted-foreground">
            <span>💳 {summary.credit_card_share_pct.toFixed(0)}% CC</span>
            <span>•</span>
            <span>🏦 {summary.direct_bank_share_pct.toFixed(0)}% Bank</span>
          </div>
        </div>

        {/* Net Savings / Surplus */}
        <div
          className={`rounded-xl border p-4 shadow-xs ${
            summary.net_surplus >= 0
              ? 'bg-emerald-500/5 border-emerald-500/20'
              : 'bg-rose-500/5 border-rose-500/20'
          }`}
        >
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-muted-foreground">
              {summary.net_surplus >= 0 ? 'Net Surplus (Retained)' : 'Monthly Deficit'}
            </span>
            <div
              className={`flex h-7 w-7 items-center justify-center rounded-lg ${
                summary.net_surplus >= 0
                  ? 'bg-emerald-500/10 text-emerald-500'
                  : 'bg-rose-500/10 text-rose-500'
              }`}
            >
              <PiggyBank className="h-4 w-4" />
            </div>
          </div>
          <div className="mt-2 flex items-baseline justify-between gap-2">
            <span
              className={`text-2xl font-bold tracking-tight ${
                summary.net_surplus >= 0
                  ? 'text-emerald-600 dark:text-emerald-400'
                  : 'text-rose-600 dark:text-rose-400'
              }`}
            >
              {formatINR(summary.net_surplus)}
            </span>
            <Badge
              variant="outline"
              className={`text-[10px] font-bold ${
                summary.savings_rate >= 30
                  ? 'bg-emerald-500/15 text-emerald-600 border-emerald-500/30'
                  : summary.savings_rate >= 0
                  ? 'bg-blue-500/15 text-blue-600 border-blue-500/30'
                  : 'bg-rose-500/15 text-rose-600 border-rose-500/30'
              }`}
            >
              {summary.savings_rate.toFixed(1)}% Rate
            </Badge>
          </div>
          <p className="text-[11px] text-muted-foreground mt-1">
            {summary.savings_rate >= 20
              ? 'Healthy savings runway for wealth creation'
              : 'Monitor discretionary spend velocity'}
          </p>
        </div>

        {/* Needs vs Wants Budget Allocation */}
        <div className="rounded-xl border bg-card/70 p-4 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-muted-foreground">Needs vs Wants</span>
            <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-indigo-500/10 text-indigo-500">
              <PieChart className="h-4 w-4" />
            </div>
          </div>
          <div className="mt-2 flex items-baseline justify-between">
            <span className="text-sm font-bold text-foreground">
              {formatINR(summary.fixed_needs_spend)} / {formatINR(summary.discretionary_spend)}
            </span>
          </div>
          <div className="mt-2 w-full bg-muted/60 h-2 rounded-full overflow-hidden flex">
            <div
              className="bg-blue-500 h-full transition-all"
              style={{
                width: `${summary.total_outflow > 0 ? (summary.fixed_needs_spend / summary.total_outflow) * 100 : 50}%`,
              }}
              title={`Needs: ${formatINR(summary.fixed_needs_spend)}`}
            />
            <div
              className="bg-amber-500 h-full transition-all"
              style={{
                width: `${summary.total_outflow > 0 ? (summary.discretionary_spend / summary.total_outflow) * 100 : 50}%`,
              }}
              title={`Wants: ${formatINR(summary.discretionary_spend)}`}
            />
          </div>
          <div className="mt-1 flex items-center justify-between text-[10px] text-muted-foreground">
            <span>Essential Needs</span>
            <span>Discretionary</span>
          </div>
        </div>
      </div>

      {/* Main Sankey Diagram Card */}
      <div className="rounded-xl border bg-card/70 backdrop-blur-xs p-6 shadow-xs space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b pb-4">
          <div>
            <h2 className="text-base font-bold text-foreground flex items-center gap-2">
              <Layers className="h-4 w-4 text-primary" />
              Cash Flow Routing Topology
            </h2>
            <p className="text-xs text-muted-foreground mt-0.5">
              Hover nodes or flow ribbons to inspect pathway shares; click any node to drill into contributing transactions
            </p>
          </div>

          <div className="flex items-center gap-2 text-xs">
            <Badge variant="outline" className="text-[11px] font-medium bg-emerald-500/10 text-emerald-600 border-emerald-500/20">
              Inflows ({formatINR(summary.total_inflow)})
            </Badge>
            <span>→</span>
            <Badge variant="outline" className="text-[11px] font-medium bg-rose-500/10 text-rose-600 border-rose-500/20">
              Outflows ({formatINR(summary.total_outflow)})
            </Badge>
          </div>
        </div>

        {/* The SVG Sankey Diagram */}
        <SankeyDiagram
          data={data?.sankey || { nodes: [], links: [], total_inflow: 0, total_outflow: 0, net_surplus: 0, savings_rate: 0 }}
          onNodeClick={handleNodeClick}
          selectedNodeId={activeNode?.id}
        />
      </div>

      {/* MoM Anomaly & Shift Intelligence */}
      <AnomaliesSection
        anomalies={data?.anomalies || []}
        previousMonth={previousMonth}
        selectedMonth={currentSelectedMonth}
      />

      {/* Category Shifts & 6-Month Trajectory Table */}
      <CategoryComparisonTable
        items={data?.category_comparisons || []}
        selectedMonth={currentSelectedMonth}
        previousMonth={previousMonth}
        totalOutflow={summary.total_outflow}
        onCategoryClick={handleCategoryRowClick}
      />

      {/* Drilldown Modal for Selected Node */}
      <Dialog open={isDrilldownOpen} onOpenChange={setIsDrilldownOpen}>
        <DialogContent className="max-w-2xl max-h-[80vh] flex flex-col">
          <DialogHeader>
            <DialogTitle className="flex items-center justify-between pr-6">
              <div className="flex items-center gap-2.5">
                <div
                  className="h-3 w-3 rounded-full"
                  style={{ backgroundColor: activeNode?.color_hex || '#3B82F6' }}
                />
                <span>{activeNode?.name} Transactions</span>
              </div>
              <Badge variant="outline" className="font-bold text-xs">
                {activeNode && formatINR(activeNode.total_value)}
              </Badge>
            </DialogTitle>
          </DialogHeader>

          <div className="flex-1 overflow-y-auto mt-3">
            {isNodeTxsLoading ? (
              <div className="flex h-40 items-center justify-center text-xs text-muted-foreground">
                <RefreshCw className="h-4 w-4 animate-spin mr-2" />
                Loading node transactions...
              </div>
            ) : !nodeTxs || nodeTxs.length === 0 ? (
              <div className="flex h-32 items-center justify-center text-xs text-muted-foreground">
                No individual ledger records found for this node
              </div>
            ) : (
              <div className="divide-y divide-border/60 text-xs">
                {nodeTxs.map((tx: Transaction) => (
                  <div
                    key={tx.id}
                    className="flex items-center justify-between py-2.5 px-1 hover:bg-muted/40 transition-colors"
                  >
                    <div className="min-w-0 pr-3">
                      <div className="font-semibold text-foreground truncate">
                        {tx.cleaned_payee || tx.raw_narration}
                      </div>
                      <div className="text-[11px] text-muted-foreground flex items-center gap-2 mt-0.5">
                        <span>{formatDate(tx.tx_date)}</span>
                        {tx.payment_mode && (
                          <Badge variant="secondary" className="text-[9px] px-1 py-0 font-normal">
                            {tx.payment_mode}
                          </Badge>
                        )}
                        {tx.account_name && (
                          <span className="truncate opacity-75">{tx.account_name}</span>
                        )}
                      </div>
                    </div>

                    <div className="text-right shrink-0">
                      <div
                        className={`font-bold ${
                          tx.tx_type === 'CREDIT'
                            ? 'text-emerald-600 dark:text-emerald-400'
                            : 'text-foreground'
                        }`}
                      >
                        {tx.tx_type === 'CREDIT' ? '+' : '-'}
                        {formatINR(tx.amount)}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
