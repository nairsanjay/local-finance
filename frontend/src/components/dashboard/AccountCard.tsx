import React from 'react'
import { Account } from '@/types'
import { formatINR } from '@/lib/utils'
import { CreditCard, Landmark } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

interface AccountCardProps {
  account: Account
}

export const AccountCard: React.FC<AccountCardProps> = ({ account }) => {
  const isCreditCard = account.account_type === 'CREDIT_CARD'

  // Card network badge colors
  const networkBadge = account.card_network ? (
    <Badge variant="outline" className="text-[10px] font-mono font-semibold px-1.5 py-0">
      {account.card_network}
    </Badge>
  ) : null

  const variantBadge = account.card_variant ? (
    <Badge variant="secondary" className="text-[10px] font-medium px-1.5 py-0">
      {account.card_variant}
    </Badge>
  ) : null

  return (
    <Card className="overflow-hidden border-border/80 bg-card shadow-xs transition-colors hover:border-border">
      <CardContent className="p-5 flex flex-col justify-between h-full">
        <div>
          <div className="flex items-start justify-between gap-2">
            <div className="flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-muted text-foreground border">
                {isCreditCard ? (
                  <CreditCard className="h-4 w-4 text-amber-400" />
                ) : (
                  <Landmark className="h-4 w-4 text-sky-400" />
                )}
              </div>
              <div>
                <h4 className="text-sm font-semibold text-foreground tracking-tight">
                  {account.bank_name}
                </h4>
                <div className="flex items-center gap-1.5 mt-0.5 flex-wrap">
                  <Badge variant="outline" className="text-[10px] uppercase font-semibold py-0 px-1.5 text-muted-foreground border-border/80">
                    {account.account_type.replace('_', ' ')}
                  </Badge>
                  <span className="text-[11px] font-mono text-muted-foreground">
                    {account.account_number ? account.account_number : account.account_number_mask}
                  </span>
                  {variantBadge}
                  {account.account_holder_name && (
                    <span className="text-[10px] text-muted-foreground truncate max-w-[120px]">
                      • {account.account_holder_name}
                    </span>
                  )}
                </div>
              </div>
            </div>
            {networkBadge || (
              <Badge variant="outline" className="text-[10px] font-mono px-1.5 py-0">
                {account.currency || 'INR'}
              </Badge>
            )}
          </div>
        </div>

        <div className="mt-5 pt-3 border-t border-border/60 flex items-baseline justify-between">
          <div>
            <p className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider">
              {isCreditCard ? 'Current Outstanding' : 'Available Balance'}
            </p>
            <p
              className={`text-lg font-bold font-mono tracking-tight tabular-nums mt-0.5 ${
                isCreditCard
                  ? account.current_balance > 0
                    ? 'text-rose-400'
                    : 'text-foreground'
                  : account.current_balance < 0
                  ? 'text-rose-400'
                  : 'text-emerald-400'
              }`}
            >
              {formatINR(account.current_balance)}
            </p>
          </div>

          {isCreditCard && account.credit_limit && (
            <div className="text-right">
              <p className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider">Credit Limit</p>
              <p className="text-sm font-semibold font-mono text-muted-foreground tabular-nums mt-0.5">
                {formatINR(account.credit_limit)}
              </p>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
