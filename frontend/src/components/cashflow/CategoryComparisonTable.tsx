import React, { useState } from 'react'
import { CategoryComparisonItem } from '@/types'
import { formatINR } from '@/lib/utils'
import {
  ArrowUpRight,
  ArrowDownRight,
  ArrowRight,
  Search,
  SlidersHorizontal,
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'

interface CategoryComparisonTableProps {
  items: CategoryComparisonItem[]
  selectedMonth: string
  previousMonth: string
  totalOutflow: number
  onCategoryClick?: (categoryId: string) => void
}

export const CategoryComparisonTable: React.FC<CategoryComparisonTableProps> = ({
  items,
  selectedMonth,
  previousMonth,
  totalOutflow,
  onCategoryClick,
}) => {
  const [search, setSearch] = useState('')
  const [sortBy, setSortBy] = useState<'spend' | 'delta_asc' | 'delta_desc' | 'pct'>('spend')

  const filteredItems = items
    .filter((it) =>
      it.category_name.toLowerCase().includes(search.toLowerCase()) ||
      (it.top_payees && it.top_payees.some((p) => p.payee.toLowerCase().includes(search.toLowerCase())))
    )
    .sort((a, b) => {
      if (sortBy === 'spend') return b.current_spend - a.current_spend
      if (sortBy === 'delta_desc') return b.delta_amount - a.delta_amount
      if (sortBy === 'delta_asc') return a.delta_amount - b.delta_amount
      if (sortBy === 'pct') return b.percentage_change - a.percentage_change
      return 0
    })

  // Helper to render mini sparkline
  const renderSparkline = (history: { month: string; amount: number }[], color: string) => {
    if (!history || history.length < 2) {
      return <span className="text-[10px] text-muted-foreground">-</span>
    }

    const width = 80
    const height = 22
    const maxVal = Math.max(...history.map((h) => h.amount), 1)
    const minVal = Math.min(...history.map((h) => h.amount), 0)
    const range = maxVal - minVal || 1

    const points = history.map((h, i) => {
      const x = (i / (history.length - 1)) * (width - 4) + 2
      const y = height - 2 - ((h.amount - minVal) / range) * (height - 6)
      return `${x.toFixed(1)},${y.toFixed(1)}`
    }).join(' ')

    return (
      <svg width={width} height={height} className="overflow-visible">
        <polyline
          fill="none"
          stroke={color || '#3B82F6'}
          strokeWidth="1.5"
          strokeLinecap="round"
          strokeLinejoin="round"
          points={points}
        />
        {history.map((h, i) => {
          const x = (i / (history.length - 1)) * (width - 4) + 2
          const y = height - 2 - ((h.amount - minVal) / range) * (height - 6)
          return (
            <circle
              key={i}
              cx={x}
              cy={y}
              r={i === history.length - 1 ? 2.5 : 1}
              fill={color || '#3B82F6'}
            />
          )
        })}
      </svg>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-2">
            <SlidersHorizontal className="h-3.5 w-3.5 text-primary" />
            Category Shifts &amp; 6-Month Trajectory
          </h3>
          <p className="text-[11px] text-muted-foreground mt-0.5">
            Compare category spending against {previousMonth} and 3-month rolling averages
          </p>
        </div>

        <div className="flex items-center gap-2 w-full sm:w-auto">
          <div className="relative flex-1 sm:w-56">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search category / payee..."
              className="h-8 pl-8 text-xs"
            />
          </div>

          <div className="flex items-center gap-1 text-xs">
            <button
              onClick={() => setSortBy('spend')}
              className={`px-2 py-1 rounded-md text-[11px] font-medium transition-colors ${
                sortBy === 'spend'
                  ? 'bg-primary/10 text-primary font-semibold'
                  : 'text-muted-foreground hover:bg-muted'
              }`}
            >
              Highest Spend
            </button>
            <button
              onClick={() => setSortBy('delta_desc')}
              className={`px-2 py-1 rounded-md text-[11px] font-medium transition-colors ${
                sortBy === 'delta_desc'
                  ? 'bg-primary/10 text-primary font-semibold'
                  : 'text-muted-foreground hover:bg-muted'
              }`}
            >
              Surges
            </button>
            <button
              onClick={() => setSortBy('delta_asc')}
              className={`px-2 py-1 rounded-md text-[11px] font-medium transition-colors ${
                sortBy === 'delta_asc'
                  ? 'bg-primary/10 text-primary font-semibold'
                  : 'text-muted-foreground hover:bg-muted'
              }`}
            >
              Drops
            </button>
          </div>
        </div>
      </div>

      <div className="rounded-xl border bg-card/60 backdrop-blur-xs overflow-hidden shadow-xs">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-muted/50 border-b text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
              <tr>
                <th className="py-3 px-4">Category</th>
                <th className="py-3 px-4 text-right">Spend ({selectedMonth})</th>
                <th className="py-3 px-4 text-right">Share</th>
                <th className="py-3 px-4 text-right">MoM Shift</th>
                <th className="py-3 px-4 text-right">3-Mo Avg</th>
                <th className="py-3 px-4 text-center">6-Mo Trend</th>
                <th className="py-3 px-4">Top Payees</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {filteredItems.map((item) => {
                const sharePct = totalOutflow > 0 ? (item.current_spend / totalOutflow) * 100 : 0

                return (
                  <tr
                    key={item.category_id}
                    onClick={() => onCategoryClick && onCategoryClick(item.category_id)}
                    className="hover:bg-muted/40 transition-colors cursor-pointer"
                  >
                    {/* Category */}
                    <td className="py-3 px-4 font-semibold text-foreground">
                      <div className="flex items-center gap-2.5">
                        <div
                          className="h-2.5 w-2.5 rounded-full shrink-0"
                          style={{ backgroundColor: item.category_color }}
                        />
                        <span className="truncate">{item.category_name}</span>
                      </div>
                    </td>

                    {/* Spend */}
                    <td className="py-3 px-4 text-right font-bold text-foreground">
                      {formatINR(item.current_spend)}
                    </td>

                    {/* Share */}
                    <td className="py-3 px-4 text-right text-muted-foreground">
                      {sharePct.toFixed(1)}%
                    </td>

                    {/* MoM Shift */}
                    <td className="py-3 px-4 text-right">
                      <div className="flex items-center justify-end gap-1.5 font-semibold">
                        {item.delta_amount > 0 ? (
                          <span className="flex items-center gap-0.5 text-amber-600 dark:text-amber-400">
                            <ArrowUpRight className="h-3.5 w-3.5 shrink-0" />
                            +{formatINR(item.delta_amount)}
                            <span className="text-[10px] font-normal opacity-85">
                              (+{item.percentage_change.toFixed(0)}%)
                            </span>
                          </span>
                        ) : item.delta_amount < 0 ? (
                          <span className="flex items-center gap-0.5 text-emerald-600 dark:text-emerald-400">
                            <ArrowDownRight className="h-3.5 w-3.5 shrink-0" />
                            -{formatINR(Math.abs(item.delta_amount))}
                            <span className="text-[10px] font-normal opacity-85">
                              ({item.percentage_change.toFixed(0)}%)
                            </span>
                          </span>
                        ) : (
                          <span className="flex items-center gap-0.5 text-muted-foreground">
                            <ArrowRight className="h-3 w-3 shrink-0" />
                            ₹0 (0%)
                          </span>
                        )}
                      </div>
                    </td>

                    {/* 3-Month Average */}
                    <td className="py-3 px-4 text-right text-muted-foreground font-medium">
                      {item.three_month_avg > 0 ? formatINR(item.three_month_avg) : '-'}
                    </td>

                    {/* 6-Month Sparkline */}
                    <td className="py-3 px-4 text-center">
                      <div className="flex justify-center">
                        {renderSparkline(item.history, item.category_color)}
                      </div>
                    </td>

                    {/* Top Payees */}
                    <td className="py-3 px-4">
                      {item.top_payees && item.top_payees.length > 0 ? (
                        <div className="flex flex-wrap gap-1">
                          {item.top_payees.map((p, idx) => (
                            <Badge
                              key={idx}
                              variant="secondary"
                              className="text-[10px] px-1.5 py-0 font-normal bg-muted text-muted-foreground"
                            >
                              {p.payee}
                            </Badge>
                          ))}
                        </div>
                      ) : (
                        <span className="text-muted-foreground text-[11px]">-</span>
                      )}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
