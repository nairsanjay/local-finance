import { useQuery } from '@tanstack/react-query'
import { Link, useNavigate, useSearch } from '@tanstack/react-router'
import { ArrowRight, CalendarDays } from 'lucide-react'
import { fetchAnalytics, fetchAnalyticsRange } from '@/lib/api'
import { formatINR } from '@/lib/utils'
import {
  spendingMonthLabel,
  spendingMonthRange,
  validSpendingMonth,
  type SpendingSearch,
} from '@/lib/spending'
import { usePrivacy } from '@/components/privacy-provider'
import { SpendingChart } from '@/components/dashboard/SpendingChart'
import { SpendingTransactions } from '@/components/dashboard/SpendingTransactions'
import { Button, buttonVariants } from '@/components/ui/button'
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Empty,
  EmptyHeader,
  EmptyTitle,
  EmptyDescription,
} from '@/components/ui/empty'
import { PageLoading, QueryError } from '@/components/ui/query-state'

export function SpendingDashboard() {
  const { month, category } = useSearch({ strict: false }) as SpendingSearch
  const navigate = useNavigate()
  const { isPrivacyMode } = usePrivacy()
  const baseline = useQuery({
    queryKey: ['analytics'],
    queryFn: fetchAnalytics,
  })
  const overview = useQuery({
    queryKey: month ? ['analytics', 'spending', month] : ['analytics'],
    queryFn: () => fetchAnalyticsRange(spendingMonthRange(month)),
  })
  const selectMonth = (value?: string) => {
    void navigate({ to: '/spending', search: { month: value }, resetScroll: false })
  }
  const categories = overview.data?.category_breakdown ?? []
  const selected = categories.find((item) => item.category_id === category)
  const purchases = categories.reduce((sum, item) => sum + item.tx_count, 0)
  const recentMonths = baseline.data?.monthly_trends ?? []
  const largestMonth = Math.max(1, ...recentMonths.map((item) => item.expense))

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
            Spending
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            See where your money went, then open the purchases behind it.
          </p>
        </div>
        <Link to="/import" className={buttonVariants({ variant: 'outline' })}>
          Import a statement <ArrowRight />
        </Link>
      </div>
      <Card>
        <CardContent className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <p className="flex items-center gap-2 text-sm font-medium">
              <CalendarDays className="size-4 text-primary" />{' '}
              {spendingMonthLabel(month)}
            </p>
            <p className="mt-1 text-xs text-muted-foreground">
              Personal spending across your accounts and Splitwise.
            </p>
          </div>
          <div className="flex flex-wrap items-end gap-3">
            <Button
              variant={month ? 'outline' : 'secondary'}
              aria-pressed={!month}
              onClick={() => selectMonth()}
            >
              All time
            </Button>
            <div className="space-y-1.5">
              <Label htmlFor="spending-month" className="text-xs">
                Choose month
              </Label>
              <Input
                id="spending-month"
                type="month"
                min="0001-01"
                max="9999-12"
                value={month ?? ''}
                className="w-44"
                onChange={(event) => {
                  const value = event.target.value
                  if (!value || validSpendingMonth(value))
                    selectMonth(value || undefined)
                }}
              />
            </div>
          </div>
        </CardContent>
      </Card>
      {overview.isPending ? (
        <PageLoading label="Loading spending" />
      ) : overview.isError ? (
        <QueryError
          message="Unable to load spending"
          onRetry={overview.refetch}
        />
      ) : (
        <>
          <div className="grid gap-4 sm:grid-cols-3">
            {[
              {
                title: 'Total spent',
                value: formatINR(overview.data.total_expense, isPrivacyMode),
                description: `${purchases} ${purchases === 1 ? 'purchase' : 'purchases'}`,
              },
              {
                title: 'Biggest category',
                value: categories[0]?.category_name ?? 'No spending yet',
                description: categories[0]
                  ? formatINR(categories[0].total_amount, isPrivacyMode)
                  : 'Import a statement to begin',
              },
              {
                title: 'Average purchase',
                value: formatINR(
                  purchases ? overview.data.total_expense / purchases : 0,
                  isPrivacyMode
                ),
                description: `${categories.length} spending ${categories.length === 1 ? 'category' : 'categories'}`,
              },
            ].map((item) => (
              <Card key={item.title}>
                <CardHeader>
                  <CardDescription>{item.title}</CardDescription>
                  <CardTitle className="text-xl tabular-nums">
                    {item.value}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-xs text-muted-foreground">
                    {item.description}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
          {recentMonths.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle>Recent months</CardTitle>
                <CardDescription>
                  Latest 12 recorded months. Choose one to explore, or use the
                  month picker for older months.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div
                  className="grid grid-cols-2 gap-2 sm:grid-cols-4 lg:grid-cols-6"
                  role="group"
                  aria-label="Monthly spending"
                >
                  {recentMonths.map((item) => (
                    <Button
                      key={item.month}
                      variant={month === item.month ? 'secondary' : 'ghost'}
                      aria-pressed={month === item.month}
                      className="h-auto min-w-0 flex-col items-start gap-2 px-3 py-3"
                      onClick={() => selectMonth(item.month)}
                    >
                      <span className="text-xs text-muted-foreground">
                        {spendingMonthLabel(item.month)}
                      </span>
                      <span className="text-sm tabular-nums">
                        {formatINR(item.expense, isPrivacyMode)}
                      </span>
                      <span
                        aria-hidden="true"
                        className="h-1.5 w-full overflow-hidden rounded-full bg-muted"
                      >
                        <span
                          className="block h-full rounded-full bg-primary"
                          style={{
                            width: `${(item.expense / largestMonth) * 100}%`,
                          }}
                        />
                      </span>
                    </Button>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
          {baseline.isError && month && (
            <QueryError
              message="Unable to load recent months"
              onRetry={baseline.refetch}
            />
          )}
          <SpendingChart
            data={categories}
            selectedCategory={category}
            onCategorySelect={(id) => {
              void navigate({
                to: '/spending',
                search: { month, category: id === category ? undefined : id },
                resetScroll: false,
              })
            }}
          />
          {selected ? (
            <SpendingTransactions
              key={`${month ?? 'all'}:${selected.category_id}`}
              category={selected}
              month={month}
            />
          ) : categories.length > 0 ? (
            <Empty className="rounded-xl border border-dashed">
              <EmptyHeader>
                <EmptyTitle>Explore a category</EmptyTitle>
                <EmptyDescription>
                  Choose a category above to see its individual purchases.
                </EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            <p role="status" className="text-sm text-muted-foreground">
              {month
                ? 'No spending recorded this month. Choose another month or import a statement.'
                : 'Import a statement to start exploring your spending.'}
            </p>
          )}
          <p className="text-xs text-muted-foreground">
            Transfers and excluded transactions are omitted. Splitwise uses your
            personal share. Credits and refunds are not deducted. Totals reflect
            imported data; missing statements can make spending look lower.
          </p>
        </>
      )}
    </div>
  )
}
