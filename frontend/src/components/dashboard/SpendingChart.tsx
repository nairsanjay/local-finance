import React from 'react'
import { ResponsiveContainer, PieChart, Pie, Cell, Tooltip } from 'recharts'
import { CategorySpend } from '@/types'
import { formatINR } from '@/lib/utils'
import { PieChart as PieIcon } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'

interface SpendingChartProps {
  data: CategorySpend[]
}

const PALETTE = [
  '#f43f5e', '#10b981', '#f59e0b', '#3b82f6',
  '#6366f1', '#ec4899', '#14b8a6', '#8b5cf6',
  '#f97316', '#06b6d4', '#84cc16', '#64748b',
]

export const SpendingChart: React.FC<SpendingChartProps> = ({ data }) => {
  if (!data || data.length === 0) {
    return (
      <Card className="flex h-80 flex-col items-center justify-center text-center p-6 border-border/80 bg-card">
        <PieIcon className="mb-2 h-8 w-8 text-muted-foreground/50" />
        <CardTitle className="text-sm font-semibold text-foreground">No spend data available</CardTitle>
        <CardDescription className="text-xs text-muted-foreground mt-1">
          Import a bank or credit card statement to view category breakdown.
        </CardDescription>
      </Card>
    )
  }

  const chartData = data.map((item, index) => ({
    name: item.category_name,
    value: item.total_amount,
    percentage: item.percentage,
    color: item.color_hex || PALETTE[index % PALETTE.length],
  }))

  return (
    <Card className="border-border/80 bg-card shadow-xs">
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-semibold">Expense by Category</CardTitle>
        <CardDescription className="text-xs">Distribution of categorized spending across accounts</CardDescription>
      </CardHeader>
      <CardContent className="pt-2">
        <div className="grid grid-cols-1 md:grid-cols-12 gap-4 items-center">
          <div className="h-60 md:col-span-6">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={chartData}
                  dataKey="value"
                  nameKey="name"
                  cx="50%"
                  cy="50%"
                  innerRadius={55}
                  outerRadius={85}
                  paddingAngle={2}
                >
                  {chartData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} stroke="var(--card)" strokeWidth={2} />
                  ))}
                </Pie>
                <Tooltip
                  formatter={(value: any) => [formatINR(Number(value)), 'Spent']}
                  contentStyle={{
                    backgroundColor: 'var(--popover)',
                    borderColor: 'var(--border)',
                    borderRadius: 'var(--radius)',
                    color: 'var(--popover-foreground)',
                    fontSize: '12px',
                  }}
                  itemStyle={{ color: 'var(--foreground)' }}
                />
              </PieChart>
            </ResponsiveContainer>
          </div>

          {/* Legend list */}
          <div className="space-y-2 md:col-span-6 max-h-60 overflow-y-auto pr-1">
            {chartData.slice(0, 8).map((item) => (
              <div key={item.name} className="flex items-center justify-between text-xs py-1 border-b border-border/40 last:border-0">
                <div className="flex items-center gap-2 truncate pr-2">
                  <span className="h-2.5 w-2.5 rounded-full flex-shrink-0" style={{ backgroundColor: item.color }} />
                  <span className="truncate font-medium text-foreground">{item.name}</span>
                </div>
                <div className="flex items-center gap-2 text-right flex-shrink-0 font-mono">
                  <span className="font-semibold text-foreground tabular-nums">{formatINR(item.value)}</span>
                  <span className="text-muted-foreground w-9 tabular-nums">({item.percentage.toFixed(0)}%)</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
