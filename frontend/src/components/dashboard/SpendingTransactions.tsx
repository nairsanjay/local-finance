import { useEffect, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { fetchTransactions } from '@/lib/api'
import { formatDate, formatINR } from '@/lib/utils'
import { spendingMonthLabel, spendingMonthRange } from '@/lib/spending'
import type { CategorySpend } from '@/types'
import { usePrivacy } from '@/components/privacy-provider'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { QueryError } from '@/components/ui/query-state'
import { Skeleton } from '@/components/ui/skeleton'

export function SpendingTransactions({
  category,
  month,
}: {
  category: CategorySpend
  month?: string
}) {
  const [page, setPage] = useState(1)
  const panel = useRef<HTMLDivElement>(null)
  useEffect(() => {
    panel.current?.focus({ preventScroll: true })
    panel.current?.scrollIntoView({
      block: 'nearest',
      behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches
        ? 'instant'
        : 'smooth',
    })
  }, [])
  const { isPrivacyMode } = usePrivacy()
  const { data, isPending, isError, refetch } = useQuery({
    queryKey: ['transactions', 'spending', month ?? 'all', category.category_id, page],
    queryFn: () =>
      fetchTransactions({
        spending_only: true,
        category_id: category.category_id,
        ...spendingMonthRange(month),
        page,
        page_size: 10,
      }),
  })
  return (
    <Card
      role="region"
      ref={panel}
      tabIndex={-1}
      className="scroll-mt-20 outline-none focus-visible:ring-2 focus-visible:ring-ring"
      aria-label={`${category.category_name} purchases`}
    >
      <CardHeader>
        <CardTitle>{category.category_name} purchases</CardTitle>
        <CardDescription>
          {spendingMonthLabel(month)} · {formatINR(category.total_amount, isPrivacyMode)} across{' '}
          {category.tx_count} {category.tx_count === 1 ? 'purchase' : 'purchases'}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {isPending ? (
          <div role="status" aria-label="Loading category purchases" className="space-y-3">
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-8 w-full" />
          </div>
        ) : isError ? (
          <QueryError message="Unable to load category purchases" onRetry={refetch} />
        ) : data && data.items.length > 0 ? (
          <>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Date</TableHead>
                  <TableHead>Payee / account</TableHead>
                  <TableHead className="text-right">Personal spending</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.items.map((transaction) => (
                  <TableRow key={transaction.id}>
                    <TableCell className="whitespace-nowrap text-xs">
                      {formatDate(transaction.tx_date)}
                    </TableCell>
                    <TableCell className="max-w-64 whitespace-normal break-words">
                      {isPrivacyMode
                        ? 'Payee hidden'
                        : transaction.cleaned_payee || 'Unknown payee'}
                      <p className="mt-1 text-xs text-muted-foreground">
                        {isPrivacyMode ? 'Account hidden' : transaction.account_name}
                      </p>
                    </TableCell>
                    <TableCell className="text-right whitespace-nowrap tabular-nums">
                      {formatINR(transaction.amount, isPrivacyMode)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
            <div className="flex flex-wrap items-center justify-between gap-3 text-xs text-muted-foreground">
              <span aria-live="polite">
                {data.total} {data.total === 1 ? 'purchase' : 'purchases'} · Page {data.page} of{' '}
                {data.total_pages}
              </span>
              {data.total_pages > 1 && (
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page === 1}
                    onClick={() => setPage(page - 1)}
                  >
                    Previous
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page >= data.total_pages}
                    onClick={() => setPage(page + 1)}
                  >
                    Next
                  </Button>
                </div>
              )}
            </div>
          </>
        ) : (
          <p role="status" className="text-sm text-muted-foreground">
            No purchases found for this category and period.
          </p>
        )}
      </CardContent>
    </Card>
  )
}
