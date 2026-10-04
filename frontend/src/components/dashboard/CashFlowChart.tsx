import React from 'react'
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  Legend,
} from 'recharts'
import { MonthlyCashFlow } from '@/types'
import { formatINR } from '@/lib/utils'
import { BarChart3 } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'

interface CashFlowChartProps {
  data: MonthlyCashFlow[]
}

export const CashFlowChart: React.FC<CashFlowChartProps> = ({ data }) => {
  if (!data || data.length === 0) {
    return (
      <Card className="flex h-80 flex-col items-center justify-center text-center p-6 border-border/80 bg-card">
        <BarChart3 className="mb-2 h-8 w-8 text-muted-foreground/50" />
        <CardTitle className="text-sm font-semibold text-foreground">No cash flow trends</CardTitle>
        <CardDescription className="text-xs text-muted-foreground mt-1">
          Historical statements will populate monthly income and expense trends.
        </CardDescription>
      </Card>
    )
  }

  return (
    <Card className="border-border/80 bg-card shadow-xs">
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-semibold">Monthly Cash Flow</CardTitle>
        <CardDescription className="text-xs">Comparison of monthly inflow (Income) vs outflow (Expenses)</CardDescription>
      </CardHeader>
      <CardContent className="pt-2">
        <div className="h-60">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={data} margin={{ top: 10, right: 10, left: -15, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" opacity={0.6} vertical={false} />
              <XAxis
                dataKey="month"
                stroke="var(--muted-foreground)"
                fontSize={11}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                stroke="var(--muted-foreground)"
                fontSize={11}
                tickLine={false}
                axisLine={false}
                tickFormatter={(v) => `₹${v >= 1000 ? `${(v / 1000).toFixed(0)}k` : v}`}
              />
              <Tooltip
                formatter={(value: any) => [formatINR(Number(value)), '']}
                contentStyle={{
                  backgroundColor: 'var(--popover)',
                  borderColor: 'var(--border)',
                  borderRadius: 'var(--radius)',
                  color: 'var(--popover-foreground)',
                  fontSize: '12px',
                }}
              />
              <Legend
                wrapperStyle={{ paddingTop: '8px' }}
                formatter={(value) => <span className="text-xs text-muted-foreground font-medium capitalize">{value}</span>}
              />
              <Bar dataKey="income" name="Income" fill="#10b981" radius={[4, 4, 0, 0]} maxBarSize={28} />
              <Bar dataKey="expense" name="Expense" fill="#f43f5e" radius={[4, 4, 0, 0]} maxBarSize={28} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  )
}
