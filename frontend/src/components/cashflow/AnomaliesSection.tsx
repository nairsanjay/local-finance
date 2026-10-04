import React from 'react'
import { MoMAnomaly } from '@/types'
import { formatINR } from '@/lib/utils'
import {
  TrendingUp,
  TrendingDown,
  Sparkles,
  AlertTriangle,
  CheckCircle2,
  HelpCircle,
  PlusCircle,
  PiggyBank,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'

interface AnomaliesSectionProps {
  anomalies: MoMAnomaly[]
  previousMonth: string
  selectedMonth: string
}

export const AnomaliesSection: React.FC<AnomaliesSectionProps> = ({
  anomalies,
  previousMonth,
}) => {
  if (!anomalies || anomalies.length === 0) {
    return (
      <div className="flex items-center gap-3 rounded-xl border border-emerald-500/20 bg-emerald-500/5 p-4 text-xs text-foreground">
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-500">
          <CheckCircle2 className="h-4 w-4" />
        </div>
        <div>
          <h4 className="font-semibold text-emerald-700 dark:text-emerald-400">
            Stable Spending Velocity
          </h4>
          <p className="text-muted-foreground mt-0.5">
            No unusual spending spikes or anomalous shifts detected compared to {previousMonth || 'last month'}.
          </p>
        </div>
      </div>
    )
  }

  const getSeverityStyle = (severity: string, type: string) => {
    switch (severity) {
      case 'warning':
        return {
          border: 'border-amber-500/30 dark:border-amber-500/20',
          bg: 'bg-amber-500/5 dark:bg-amber-950/20',
          iconBg: 'bg-amber-500/15 text-amber-600 dark:text-amber-400',
          badge: 'bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30',
          icon: TrendingUp,
        }
      case 'danger':
        return {
          border: 'border-rose-500/30 dark:border-rose-500/20',
          bg: 'bg-rose-500/5 dark:bg-rose-950/20',
          iconBg: 'bg-rose-500/15 text-rose-600 dark:text-rose-400',
          badge: 'bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/30',
          icon: AlertTriangle,
        }
      case 'success':
        return {
          border: 'border-emerald-500/30 dark:border-emerald-500/20',
          bg: 'bg-emerald-500/5 dark:bg-emerald-950/20',
          iconBg: 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400',
          badge: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30',
          icon: type === 'SAVINGS_MILESTONE' ? PiggyBank : TrendingDown,
        }
      case 'info':
      default:
        return {
          border: 'border-blue-500/30 dark:border-blue-500/20',
          bg: 'bg-blue-500/5 dark:bg-blue-950/20',
          iconBg: 'bg-blue-500/15 text-blue-600 dark:text-blue-400',
          badge: 'bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/30',
          icon: type === 'NEW_SPEND' ? PlusCircle : Sparkles,
        }
    }
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-2">
          <Sparkles className="h-3.5 w-3.5 text-primary" />
          MoM Spending Shifts &amp; Anomalies ({anomalies.length})
        </h3>
        <span className="text-[11px] text-muted-foreground">
          Compared to {previousMonth}
        </span>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
        {anomalies.map((a) => {
          const style = getSeverityStyle(a.severity, a.type)
          const Icon = style.icon || HelpCircle

          return (
            <div
              key={a.id}
              className={`flex flex-col justify-between rounded-xl border p-3.5 transition-all hover:shadow-xs ${style.border} ${style.bg}`}
            >
              <div>
                <div className="flex items-start justify-between gap-2">
                  <div className="flex items-center gap-2">
                    <div
                      className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-lg ${style.iconBg}`}
                    >
                      <Icon className="h-4 w-4" />
                    </div>
                    <span className="text-xs font-bold text-foreground">
                      {a.title}
                    </span>
                  </div>

                  {a.percentage_change !== 0 && (
                    <Badge
                      variant="outline"
                      className={`text-[10px] font-bold px-1.5 py-0.5 shrink-0 ${style.badge}`}
                    >
                      {a.percentage_change > 0 ? '+' : ''}
                      {a.percentage_change.toFixed(0)}%
                    </Badge>
                  )}
                </div>

                <p className="text-[11px] text-muted-foreground mt-2 leading-relaxed">
                  {a.description}
                </p>
              </div>

              <div className="flex items-center justify-between border-t border-border/40 pt-2.5 mt-3 text-[11px]">
                {a.top_contributor ? (
                  <span className="text-muted-foreground truncate font-medium">
                    {a.top_contributor}
                  </span>
                ) : (
                  <span className="text-muted-foreground">
                    Current: <strong className="text-foreground">{formatINR(a.current_amount)}</strong>
                  </span>
                )}

                {a.delta_amount !== 0 && (
                  <span
                    className={`font-semibold shrink-0 ${
                      a.delta_amount > 0 ? 'text-amber-600 dark:text-amber-400' : 'text-emerald-600 dark:text-emerald-400'
                    }`}
                  >
                    {a.delta_amount > 0 ? '+' : ''}
                    {formatINR(a.delta_amount)}
                  </span>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
