import React from 'react'
import { LucideIcon } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { cn, formatINR } from '@/lib/utils'
import { AnimatedNumber } from '@/components/ui/animated-number'

interface OverviewCardProps {
  title: string
  amount: string
  numericValue?: number
  subtitle?: string
  icon: LucideIcon
  variant?: 'default' | 'income' | 'expense' | 'savings'
}

export const OverviewCard: React.FC<OverviewCardProps> = ({
  title,
  amount,
  numericValue,
  subtitle,
  icon: Icon,
  variant = 'default',
}) => {
  const variantStyles = {
    default: {
      iconBg: 'bg-primary/10 text-primary border-primary/20',
      textAccent: 'text-foreground',
    },
    income: {
      iconBg: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20',
      textAccent: 'text-emerald-600 dark:text-emerald-400',
    },
    expense: {
      iconBg: 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20',
      textAccent: 'text-foreground',
    },
    savings: {
      iconBg: 'bg-sky-500/10 text-sky-600 dark:text-sky-400 border-sky-500/20',
      textAccent: 'text-sky-600 dark:text-sky-400',
    },
  }

  const { iconBg } = variantStyles[variant]

  return (
    <Card className="overflow-hidden border-border/80 bg-card shadow-xs transition-all duration-200 hover:border-primary/30 hover:shadow-sm">
      <CardContent className="p-5">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
            {title}
          </span>
          <div className={cn('flex h-8 w-8 items-center justify-center rounded-lg border transition-transform duration-200 group-hover:scale-105', iconBg)}>
            <Icon className="h-4 w-4" />
          </div>
        </div>
        <div className="mt-3">
          <div className="text-2xl font-bold font-mono tracking-tight text-foreground tabular-nums privacy-blur">
            {numericValue !== undefined ? (
              <AnimatedNumber value={numericValue} formatFn={formatINR} />
            ) : (
              amount
            )}
          </div>
          {subtitle && (
            <p className="mt-1 text-xs text-muted-foreground flex items-center gap-1.5 font-medium">
              {subtitle}
            </p>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
