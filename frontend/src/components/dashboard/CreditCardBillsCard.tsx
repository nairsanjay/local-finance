import React from 'react'
import { CreditCardBill } from '@/types'
import { formatDate, formatINR } from '@/lib/utils'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Calendar, CreditCard, Gift } from 'lucide-react'
import { AnimatedNumber } from '@/components/ui/animated-number'

interface CreditCardBillsCardProps {
  bills: CreditCardBill[]
}

export const CreditCardBillsCard: React.FC<CreditCardBillsCardProps> = ({ bills }) => {
  if (!bills || bills.length === 0) {
    return null
  }

  return (
    <Card className="border-border/80 bg-card shadow-xs">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-base font-semibold flex items-center gap-2">
              <CreditCard className="h-4 w-4 text-amber-400" /> Credit Card Statements &amp; Dues
            </CardTitle>
            <CardDescription className="text-xs mt-0.5">
              Extracted billing statements, due dates, and reward points balances
            </CardDescription>
          </div>
          <Badge variant="secondary" className="font-mono text-xs">
            {bills.length} {bills.length === 1 ? 'Statement' : 'Statements'}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          {bills.map((bill) => {
            const isUnpaid = bill.payment_status === 'UNPAID'
            return (
              <div
                key={bill.id}
                className="rounded-lg border border-border/70 bg-muted/40 p-4 flex flex-col justify-between transition-all duration-200 hover:border-primary/40 hover:bg-muted/60"
              >
                <div>
                  <div className="flex items-start justify-between gap-2">
                    <div>
                      <h5 className="text-xs font-semibold text-foreground tracking-tight">
                        {bill.bank_name || 'Credit Card'}
                      </h5>
                      <p className="text-[11px] text-muted-foreground mt-0.5 flex items-center gap-1">
                        <Calendar className="h-3 w-3" /> Due: <span className="font-medium text-foreground">{formatDate(bill.payment_due_date)}</span>
                      </p>
                    </div>
                    <Badge
                      variant={isUnpaid ? 'destructive' : 'secondary'}
                      className="text-[10px] uppercase font-semibold tracking-wider px-1.5 py-0"
                    >
                      {bill.payment_status}
                    </Badge>
                  </div>

                  <div className="mt-3 flex items-baseline justify-between">
                    <div>
                      <p className="text-[10px] uppercase font-medium text-muted-foreground tracking-wider">Total Due</p>
                      <div className="text-base font-bold font-mono text-foreground tracking-tight tabular-nums mt-0.5">
                        <AnimatedNumber value={bill.total_due_amount} formatFn={formatINR} />
                      </div>
                    </div>
                    {bill.minimum_due_amount && bill.minimum_due_amount > 0 && (
                      <div className="text-right">
                        <p className="text-[10px] uppercase font-medium text-muted-foreground tracking-wider">Min Due</p>
                        <p className="text-xs font-semibold font-mono text-muted-foreground tabular-nums mt-0.5">
                          {formatINR(bill.minimum_due_amount)}
                        </p>
                      </div>
                    )}
                  </div>
                </div>

                {(bill.reward_points_balance > 0 || (bill.cashback_earned !== undefined && bill.cashback_earned > 0)) && (
                  <div className="mt-3 pt-2.5 border-t border-border/50 flex items-center justify-between text-[11px]">
                    <span className="text-muted-foreground flex items-center gap-1">
                      <Gift className="h-3 w-3 text-amber-400" /> Rewards / Cashback
                    </span>
                    <span className="font-mono font-semibold text-foreground tabular-nums">
                      {bill.cashback_earned ? `₹${bill.cashback_earned.toFixed(2)}` : `${bill.reward_points_balance.toLocaleString('en-IN')} pts`}
                    </span>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </CardContent>
    </Card>
  )
}
