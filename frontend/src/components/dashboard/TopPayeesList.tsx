import React from 'react'
import { PayeeSpend } from '@/types'
import { formatINR } from '@/lib/utils'
import { Store, Zap } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { MerchantAvatar } from '@/components/ui/merchant-avatar'

interface TopPayeesListProps {
  data: PayeeSpend[]
}

export const TopPayeesList: React.FC<TopPayeesListProps> = ({ data }) => {
  if (!data || data.length === 0) {
    return (
      <Card className="flex h-60 flex-col items-center justify-center text-center p-6 border-border/80 bg-card">
        <Store className="mb-2 h-8 w-8 text-muted-foreground/50" />
        <CardTitle className="text-sm font-semibold text-foreground">No payee details</CardTitle>
        <CardDescription className="text-xs text-muted-foreground mt-1">
          Cleaned UPI and card swipe payees will appear here once statements are uploaded.
        </CardDescription>
      </Card>
    )
  }

  return (
    <Card className="border-border/80 bg-card shadow-xs">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-base font-semibold">Top Payees &amp; Merchants</CardTitle>
            <CardDescription className="text-xs">Highest debit volume payees extracted across statements</CardDescription>
          </div>
          <Badge variant="outline" className="text-xs font-mono">Top {data.slice(0, 6).length}</Badge>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="divide-y divide-border/50">
          {data.slice(0, 6).map((payee, idx) => (
            <div key={`${payee.payee}-${idx}`} className="flex items-center justify-between py-2.5 transition-colors hover:bg-muted/30 px-2 rounded-lg -mx-2">
              <div className="flex items-center gap-3 min-w-0">
                <div className="relative shrink-0">
                  <MerchantAvatar payee={payee.payee} size="md" />
                  <span className="absolute -top-1 -right-1 flex h-4 w-4 items-center justify-center rounded-full bg-muted font-mono text-[9px] font-bold text-muted-foreground border border-border">
                    {idx + 1}
                  </span>
                </div>
                <div className="min-w-0">
                  <p className="text-sm font-semibold text-foreground tracking-tight truncate">{payee.payee}</p>
                  <div className="flex items-center gap-2 mt-0.5">
                    <Badge variant="secondary" className="text-[10px] font-mono px-1.5 py-0">
                      <Zap className="mr-1 h-2.5 w-2.5 text-primary" />
                      {payee.payment_mode || 'PAYMENT'}
                    </Badge>
                    <span className="text-[11px] text-muted-foreground">{payee.tx_count} {payee.tx_count === 1 ? 'txn' : 'txns'}</span>
                  </div>
                </div>
              </div>
              <div className="text-right shrink-0 pl-3">
                <span className="text-sm font-bold font-mono text-rose-400 tabular-nums">
                  {formatINR(payee.total_spent)}
                </span>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
