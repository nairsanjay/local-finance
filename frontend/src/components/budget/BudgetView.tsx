import React, { useState, useMemo, useCallback, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useSearch, useNavigate } from '@tanstack/react-router'
import {
  fetchBudgetSummary,
  upsertCategoryBudget,
  deleteCategoryBudget,
} from '../../lib/api'
import { formatINR } from '../../lib/utils'
import type { UpsertCategoryBudgetRequest, BudgetSearchParams } from '../../types'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '../ui/card'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Badge } from '../ui/badge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../ui/dialog'
import {
  Target,
  ChevronLeft,
  ChevronRight,
  Calendar,
  AlertTriangle,
  CheckCircle2,
  TrendingUp,
  Sliders,
  Plus,
  Trash2,
  PieChart,
  Search,
  Sparkles,
  DollarSign,
} from 'lucide-react'

export const BudgetView: React.FC = () => {
  const queryClient = useQueryClient()
  const searchParams = (useSearch({ strict: false }) as BudgetSearchParams) || {}
  const navigate = useNavigate()

  // Current selected month: format "YYYY-MM"
  const today = useMemo(() => new Date(), [])
  const currentMonthStr = useMemo(
    () => `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}`,
    [today]
  )

  const selectedMonth = searchParams.month || currentMonthStr
  const filterTab = (searchParams.tab as 'ALL' | 'BUDGETED' | 'WARNING' | 'UNBUDGETED') || 'ALL'
  const searchQuery = (searchParams.search || searchParams.q || '').trim()

  const [searchInput, setSearchInput] = useState<string>(searchQuery)

  useEffect(() => {
    setSearchInput(searchQuery)
  }, [searchQuery])

  const updateFilters = useCallback(
    (newParams: Partial<BudgetSearchParams>, replace: boolean = true) => {
      const current: BudgetSearchParams = {
        month: selectedMonth !== currentMonthStr ? selectedMonth : undefined,
        tab: filterTab !== 'ALL' ? filterTab : undefined,
        search: searchQuery || undefined,
      }
      const merged = { ...current, ...newParams }
      const cleaned: Record<string, string | undefined> = {}
      if (merged.month && merged.month !== currentMonthStr) cleaned.month = merged.month
      if (merged.tab && merged.tab !== 'ALL') cleaned.tab = merged.tab
      if (merged.search && merged.search.trim()) cleaned.search = merged.search.trim()

      navigate({
        to: '/budget',
        search: cleaned,
        replace,
      })
    },
    [selectedMonth, currentMonthStr, filterTab, searchQuery, navigate]
  )

  // Debounced search input sync
  useEffect(() => {
    if (searchInput.trim() === searchQuery) return
    const timer = setTimeout(() => {
      updateFilters({ search: searchInput.trim() || undefined }, true)
    }, 300)
    return () => clearTimeout(timer)
  }, [searchInput, searchQuery, updateFilters])

  // Edit / Set Budget Modal state
  const [editingCategory, setEditingCategory] = useState<{
    id: string
    name: string
    color: string
    currentLimit: number
  } | null>(null)
  const [budgetLimitInput, setBudgetLimitInput] = useState<number>(10000)

  // Fetch Budget Summary for selected month
  const { data: summary, isLoading, isError } = useQuery({
    queryKey: ['budgets', selectedMonth],
    queryFn: () => fetchBudgetSummary(selectedMonth),
  })

  // Mutations
  const upsertMutation = useMutation({
    mutationFn: (data: UpsertCategoryBudgetRequest) => upsertCategoryBudget(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['budgets'] })
      queryClient.invalidateQueries({ queryKey: ['categories'] })
      setEditingCategory(null)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: ({ categoryId, month }: { categoryId: string; month?: string }) =>
      deleteCategoryBudget(categoryId, month),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['budgets'] })
      queryClient.invalidateQueries({ queryKey: ['categories'] })
    },
  })

  // Navigation handlers
  const handlePrevMonth = () => {
    const [y, m] = selectedMonth.split('-').map(Number)
    const d = new Date(y, m - 2, 1)
    const nextMonth = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
    updateFilters({ month: nextMonth }, false)
  }

  const handleNextMonth = () => {
    const [y, m] = selectedMonth.split('-').map(Number)
    const d = new Date(y, m, 1)
    const nextMonth = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
    updateFilters({ month: nextMonth }, false)
  }

  const handleCurrentMonth = () => {
    updateFilters({ month: undefined }, false)
  }

  // Format month for display (e.g., "August 2026")
  const displayMonthName = () => {
    const [y, m] = selectedMonth.split('-').map(Number)
    const d = new Date(y, m - 1, 1)
    return d.toLocaleString('en-US', { month: 'long', year: 'numeric' })
  }

  const openBudgetModal = (cat: { id: string; name: string; color: string; currentLimit: number }) => {
    setEditingCategory(cat)
    setBudgetLimitInput(cat.currentLimit > 0 ? cat.currentLimit : 10000)
  }

  const handleSaveBudget = () => {
    if (!editingCategory) return
    upsertMutation.mutate({
      category_id: editingCategory.id,
      monthly_limit: budgetLimitInput,
      month: selectedMonth,
    })
  }

  // Filter budgets
  const filteredBudgets = (summary?.budgets || []).filter((b) => {
    const matchesSearch = b.category_name.toLowerCase().includes(searchQuery.toLowerCase())
    if (!matchesSearch) return false

    if (filterTab === 'BUDGETED') return b.monthly_limit > 0
    if (filterTab === 'WARNING') return b.status === 'WARNING' || b.status === 'EXCEEDED'
    if (filterTab === 'UNBUDGETED') return b.monthly_limit <= 0
    return true
  })

  const monthProgressPct = summary
    ? Math.min((summary.days_elapsed_in_month / Math.max(1, summary.total_days_in_month)) * 100, 100)
    : 0

  return (
    <div className="space-y-6">
      {/* HEADER & MONTH SELECTOR */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2.5">
            <Target className="h-6 w-6 text-primary" /> Category Budgeting & Pacing
          </h2>
          <p className="text-sm text-muted-foreground">
            Set spending limits, monitor pacing velocity, and stay within safe daily allowances
          </p>
        </div>

        {/* Month Navigation Control */}
        <div className="flex items-center gap-2 bg-card border rounded-lg p-1 shadow-xs">
          <Button variant="ghost" size="icon" onClick={handlePrevMonth} className="h-8 w-8">
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <div className="flex items-center gap-1.5 px-2.5 text-xs font-semibold">
            <Calendar className="h-3.5 w-3.5 text-primary" />
            <span>{displayMonthName()}</span>
            {selectedMonth === currentMonthStr && (
              <Badge variant="secondary" className="text-[9px] px-1.5 py-0 h-4">
                Current
              </Badge>
            )}
          </div>
          <Button variant="ghost" size="icon" onClick={handleNextMonth} className="h-8 w-8">
            <ChevronRight className="h-4 w-4" />
          </Button>
          {selectedMonth !== currentMonthStr && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleCurrentMonth}
              className="text-xs h-7 ml-1 font-semibold"
            >
              Today
            </Button>
          )}
        </div>
      </div>

      {isLoading ? (
        <div className="py-24 text-center">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-4 border-primary border-r-transparent"></div>
          <p className="text-sm text-muted-foreground mt-3">Calculating budget pacing & allowances...</p>
        </div>
      ) : isError || !summary ? (
        <Card className="border-destructive/30 bg-destructive/5 text-center p-8">
          <AlertTriangle className="h-8 w-8 text-destructive mx-auto mb-2" />
          <p className="text-sm font-semibold text-destructive">Failed to load budget summary.</p>
        </Card>
      ) : (
        <>
          {/* OVERSPENT / WARNING ALERT BANNER */}
          {(summary.overspent_categories_count > 0 || summary.warning_categories_count > 0) && (
            <div className="rounded-xl border border-amber-500/30 bg-amber-500/10 p-4 space-y-2">
              <div className="flex items-center gap-2 text-amber-500 font-semibold text-sm">
                <AlertTriangle className="h-4.5 w-4.5" />
                <span>Overspend & Pacing Alerts for {displayMonthName()}</span>
              </div>
              <div className="flex flex-wrap gap-2 pt-1">
                {summary.budgets
                  .filter((b) => b.status === 'EXCEEDED' || b.status === 'WARNING')
                  .map((b) => (
                    <Badge
                      key={b.id}
                      variant="outline"
                      className={`text-xs gap-1.5 py-1 px-2.5 ${
                        b.status === 'EXCEEDED'
                          ? 'border-rose-500/40 bg-rose-500/10 text-rose-400 font-semibold'
                          : 'border-amber-500/40 bg-amber-500/10 text-amber-400 font-semibold'
                      }`}
                    >
                      <span className="h-2 w-2 rounded-full" style={{ backgroundColor: b.category_color }} />
                      <strong>{b.category_name}:</strong>{' '}
                      {b.status === 'EXCEEDED'
                        ? `Exceeded by ${formatINR(Math.abs(b.remaining_amount))} (${b.spent_percentage.toFixed(0)}%)`
                        : `${b.spent_percentage.toFixed(0)}% Spent (${formatINR(b.remaining_amount)} remaining)`}
                    </Badge>
                  ))}
              </div>
            </div>
          )}

          {/* KPI RIBBON CARDS */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* 1. Total Monthly Budget */}
            <Card className="border-border/80 bg-card shadow-xs">
              <CardHeader className="pb-2">
                <CardDescription className="text-xs font-semibold text-muted-foreground flex items-center justify-between">
                  <span>TOTAL MONTHLY BUDGET</span>
                  <Target className="h-4 w-4 text-primary" />
                </CardDescription>
                <CardTitle className="text-2xl font-bold text-foreground">
                  {summary.total_budget > 0 ? formatINR(summary.total_budget) : '₹0'}
                </CardTitle>
              </CardHeader>
              <CardContent className="pt-0">
                <p className="text-xs text-muted-foreground">
                  {summary.categories_with_budgets} active category limit{summary.categories_with_budgets === 1 ? '' : 's'} configured
                </p>
              </CardContent>
            </Card>

            {/* 2. Total Spend & Progress */}
            <Card className="border-border/80 bg-card shadow-xs">
              <CardHeader className="pb-2">
                <CardDescription className="text-xs font-semibold text-muted-foreground flex items-center justify-between">
                  <span>TOTAL MONTHLY SPENT</span>
                  <TrendingUp className="h-4 w-4 text-emerald-400" />
                </CardDescription>
                <CardTitle className="text-2xl font-bold text-foreground">
                  {formatINR(summary.total_spent)}
                </CardTitle>
              </CardHeader>
              <CardContent className="pt-0 space-y-1.5">
                <div className="flex justify-between text-xs font-mono">
                  <span className="text-muted-foreground">Budget Utilized:</span>
                  <span
                    className={`font-bold ${
                      summary.overall_spent_percentage > 100
                        ? 'text-rose-400'
                        : summary.overall_spent_percentage > 80
                        ? 'text-amber-400'
                        : 'text-emerald-400'
                    }`}
                  >
                    {summary.overall_spent_percentage.toFixed(1)}%
                  </span>
                </div>
                <div className="h-1.5 w-full rounded-full bg-muted overflow-hidden">
                  <div
                    className={`h-full rounded-full transition-all ${
                      summary.overall_spent_percentage > 100
                        ? 'bg-rose-500'
                        : summary.overall_spent_percentage > 80
                        ? 'bg-amber-500'
                        : 'bg-emerald-500'
                    }`}
                    style={{ width: `${Math.min(summary.overall_spent_percentage, 100)}%` }}
                  />
                </div>
              </CardContent>
            </Card>

            {/* 3. Total Remaining Budget */}
            <Card className="border-border/80 bg-card shadow-xs">
              <CardHeader className="pb-2">
                <CardDescription className="text-xs font-semibold text-muted-foreground flex items-center justify-between">
                  <span>REMAINING BUDGET POOL</span>
                  <PieChart className="h-4 w-4 text-blue-400" />
                </CardDescription>
                <CardTitle
                  className={`text-2xl font-bold ${
                    summary.total_remaining > 0 ? 'text-foreground' : 'text-rose-400'
                  }`}
                >
                  {formatINR(summary.total_remaining)}
                </CardTitle>
              </CardHeader>
              <CardContent className="pt-0">
                <p className="text-xs text-muted-foreground font-mono">
                  {summary.days_remaining_in_month} day{summary.days_remaining_in_month === 1 ? '' : 's'} remaining in month
                </p>
              </CardContent>
            </Card>

            {/* 4. Daily Recommended Safe Allowance */}
            <Card className="border-primary/30 bg-primary/5 shadow-xs">
              <CardHeader className="pb-2">
                <CardDescription className="text-xs font-semibold text-primary flex items-center justify-between">
                  <span>DAILY SAFE ALLOWANCE</span>
                  <Sparkles className="h-4 w-4 text-primary" />
                </CardDescription>
                <CardTitle className="text-2xl font-bold text-primary font-mono">
                  {summary.overall_daily_recommended_allowance > 0
                    ? `${formatINR(summary.overall_daily_recommended_allowance)} / day`
                    : '₹0 / day'}
                </CardTitle>
              </CardHeader>
              <CardContent className="pt-0">
                <p className="text-xs text-muted-foreground">
                  Safe daily spend runway across budgeted categories
                </p>
              </CardContent>
            </Card>
          </div>

          {/* MONTH PACING PROGRESS COMPARISON */}
          <Card className="border-border/80 bg-card shadow-xs">
            <CardContent className="py-4">
              <div className="space-y-2">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between text-xs gap-1">
                  <span className="font-semibold text-foreground flex items-center gap-1.5">
                    <Calendar className="h-3.5 w-3.5 text-primary" /> Month Progress vs Spend Pacing
                  </span>
                  <div className="flex items-center gap-3 text-muted-foreground font-mono">
                    <span>
                      Day <strong>{summary.days_elapsed_in_month}</strong> of {summary.total_days_in_month} ({monthProgressPct.toFixed(0)}% elapsed)
                    </span>
                    <span>•</span>
                    <span>
                      Spend: <strong>{summary.overall_spent_percentage.toFixed(0)}%</strong> of budget
                    </span>
                  </div>
                </div>

                <div className="relative h-3 w-full rounded-full bg-muted overflow-hidden">
                  {/* Spend Bar */}
                  <div
                    className={`h-full rounded-full transition-all ${
                      summary.overall_spent_percentage > 100
                        ? 'bg-rose-500'
                        : summary.overall_spent_percentage > monthProgressPct + 10
                        ? 'bg-amber-500'
                        : 'bg-emerald-500'
                    }`}
                    style={{ width: `${Math.min(summary.overall_spent_percentage, 100)}%` }}
                  />
                  {/* Month Timeline Marker Pin */}
                  <div
                    className="absolute top-0 bottom-0 w-0.5 bg-foreground/80 z-10"
                    style={{ left: `${monthProgressPct}%` }}
                    title={`Day ${summary.days_elapsed_in_month} Marker`}
                  />
                </div>

                <div className="flex justify-between text-[10px] text-muted-foreground">
                  <span>Day 1</span>
                  <span className="font-semibold text-foreground">
                    {summary.overall_spent_percentage <= monthProgressPct
                      ? '✨ Great job! You are spending under the expected calendar pace.'
                      : '⚡ Warning: Spending pace is slightly faster than calendar days elapsed.'}
                  </span>
                  <span>Day {summary.total_days_in_month}</span>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* CATEGORY BUDGETS LEDGER */}
          <Card className="border-border/80 bg-card shadow-xs">
            <CardHeader className="pb-3 border-b">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div>
                  <CardTitle className="text-base font-semibold flex items-center gap-2">
                    <Sliders className="h-4.5 w-4.5 text-primary" /> Category Budgets & Pacing Breakdown
                  </CardTitle>
                  <CardDescription className="text-xs">
                    Inspect category limits, actual expenditures, and daily allowance recommendations
                  </CardDescription>
                </div>

                {/* Filter Tabs & Search */}
                <div className="flex flex-wrap items-center gap-2">
                  <div className="relative w-44">
                    <Search className="h-3.5 w-3.5 absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
                    <Input
                      placeholder="Search category..."
                      value={searchInput}
                      onChange={(e) => setSearchInput(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                          updateFilters({ search: searchInput.trim() || undefined }, true)
                        }
                      }}
                      className="pl-8 h-8 text-xs"
                    />
                  </div>

                  <div className="flex items-center gap-1 bg-muted/40 p-0.5 rounded-lg border text-xs">
                    <Button
                      variant={filterTab === 'ALL' ? 'secondary' : 'ghost'}
                      size="sm"
                      onClick={() => updateFilters({ tab: undefined }, true)}
                      className="h-7 text-xs font-semibold px-2.5"
                    >
                      All ({summary.budgets.length})
                    </Button>
                    <Button
                      variant={filterTab === 'BUDGETED' ? 'secondary' : 'ghost'}
                      size="sm"
                      onClick={() => updateFilters({ tab: 'BUDGETED' }, true)}
                      className="h-7 text-xs font-semibold px-2.5"
                    >
                      Active ({summary.categories_with_budgets})
                    </Button>
                    <Button
                      variant={filterTab === 'WARNING' ? 'secondary' : 'ghost'}
                      size="sm"
                      onClick={() => updateFilters({ tab: 'WARNING' }, true)}
                      className="h-7 text-xs font-semibold px-2.5"
                    >
                      Overspent ({summary.overspent_categories_count + summary.warning_categories_count})
                    </Button>
                    <Button
                      variant={filterTab === 'UNBUDGETED' ? 'secondary' : 'ghost'}
                      size="sm"
                      onClick={() => updateFilters({ tab: 'UNBUDGETED' }, true)}
                      className="h-7 text-xs font-semibold px-2.5"
                    >
                      Unset ({summary.budgets.length - summary.categories_with_budgets})
                    </Button>
                  </div>
                </div>
              </div>
            </CardHeader>

            <CardContent className="pt-4">
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3.5">
                {filteredBudgets.map((b) => {
                  const hasBudget = b.monthly_limit > 0
                  const isExceeded = b.status === 'EXCEEDED'
                  const isWarning = b.status === 'WARNING'

                  return (
                    <div
                      key={b.id}
                      className={`rounded-xl border p-4 transition-all space-y-3 ${
                        isExceeded
                          ? 'border-rose-500/40 bg-rose-500/5 ring-1 ring-rose-500/20'
                          : isWarning
                          ? 'border-amber-500/40 bg-amber-500/5 ring-1 ring-amber-500/20'
                          : hasBudget
                          ? 'border-border/80 bg-card shadow-xs'
                          : 'border-dashed border-border/60 bg-muted/10 opacity-80'
                      }`}
                    >
                      {/* Header: Name, Color & Badge */}
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <span
                            className="h-3 w-3 rounded-full flex-shrink-0"
                            style={{ backgroundColor: b.category_color || '#64748B' }}
                          />
                          <h4 className="font-bold text-sm text-foreground">{b.category_name}</h4>
                        </div>

                        {isExceeded ? (
                          <Badge className="bg-rose-500/10 text-rose-400 border-rose-500/30 text-[10px] font-bold gap-1">
                            <AlertTriangle className="h-3 w-3" /> Exceeded!
                          </Badge>
                        ) : isWarning ? (
                          <Badge className="bg-amber-500/10 text-amber-400 border-amber-500/30 text-[10px] font-bold gap-1">
                            <AlertTriangle className="h-3 w-3" /> {b.spent_percentage.toFixed(0)}% Alert
                          </Badge>
                        ) : hasBudget ? (
                          <Badge className="bg-emerald-500/10 text-emerald-400 border-emerald-500/30 text-[10px] font-semibold gap-1">
                            <CheckCircle2 className="h-3 w-3" /> On Track
                          </Badge>
                        ) : (
                          <Badge variant="outline" className="text-[10px] text-muted-foreground border-dashed">
                            No Budget
                          </Badge>
                        )}
                      </div>

                      {/* Spend & Limit Numbers */}
                      <div className="flex justify-between items-end">
                        <div>
                          <p className="text-[10px] text-muted-foreground">Spent this month</p>
                          <p className="text-lg font-bold font-mono text-foreground">
                            {formatINR(b.actual_spend)}
                          </p>
                        </div>
                        <div className="text-right">
                          <p className="text-[10px] text-muted-foreground">Monthly Target</p>
                          <p className="text-sm font-semibold font-mono text-foreground">
                            {hasBudget ? formatINR(b.monthly_limit) : 'Not Configured'}
                          </p>
                        </div>
                      </div>

                      {/* Visual Pacing Bar */}
                      {hasBudget ? (
                        <div className="space-y-1.5">
                          <div className="relative h-2 w-full rounded-full bg-muted overflow-hidden">
                            <div
                              className={`h-full rounded-full transition-all ${
                                isExceeded ? 'bg-rose-500' : isWarning ? 'bg-amber-500' : 'bg-primary'
                              }`}
                              style={{ width: `${Math.min(b.spent_percentage, 100)}%` }}
                            />
                          </div>

                          <div className="flex justify-between items-center text-[11px] font-mono">
                            <span
                              className={
                                isExceeded ? 'text-rose-400 font-bold' : 'text-muted-foreground'
                              }
                            >
                              {isExceeded
                                ? `Overspent by ${formatINR(Math.abs(b.remaining_amount))}`
                                : `${formatINR(b.remaining_amount)} remaining`}
                            </span>
                            <span className="text-muted-foreground">
                              {b.spent_percentage.toFixed(0)}%
                            </span>
                          </div>
                        </div>
                      ) : (
                        <p className="text-xs text-muted-foreground italic">
                          Click below to set a monthly spending limit for this category.
                        </p>
                      )}

                      {/* Daily Allowance & Pacing Velocity */}
                      {hasBudget && (
                        <div className="rounded-lg border bg-muted/20 p-2 text-xs space-y-1">
                          <div className="flex justify-between items-center">
                            <span className="text-muted-foreground text-[10px]">Daily Allowance:</span>
                            <span
                              className={`font-mono font-semibold text-[11px] ${
                                b.daily_recommended_allowance > 0 ? 'text-emerald-400' : 'text-rose-400'
                              }`}
                            >
                              {b.daily_recommended_allowance > 0
                                ? `${formatINR(b.daily_recommended_allowance)} / day`
                                : '₹0 / day remaining'}
                            </span>
                          </div>
                          {b.projected_month_end_spend > 0 && (
                            <div className="flex justify-between items-center text-[10px]">
                              <span className="text-muted-foreground">Projected Spend:</span>
                              <span
                                className={`font-mono ${
                                  b.pacing_status === 'PACING_HIGH' || b.pacing_status === 'PACING_EXCEEDED'
                                    ? 'text-amber-400 font-semibold'
                                    : 'text-muted-foreground'
                                }`}
                              >
                                ~{formatINR(b.projected_month_end_spend)}
                              </span>
                            </div>
                          )}
                        </div>
                      )}

                      {/* Action Controls */}
                      <div className="flex items-center justify-end gap-2 pt-1 border-t border-border/50">
                        {hasBudget && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() =>
                              deleteMutation.mutate({ categoryId: b.category_id, month: selectedMonth })
                            }
                            disabled={deleteMutation.isPending}
                            className="h-7 text-xs text-muted-foreground hover:text-rose-400 px-2"
                            title="Remove budget target"
                          >
                            <Trash2 className="h-3 w-3" />
                          </Button>
                        )}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() =>
                            openBudgetModal({
                              id: b.category_id,
                              name: b.category_name,
                              color: b.category_color,
                              currentLimit: b.monthly_limit,
                            })
                          }
                          className="h-7 text-xs font-semibold gap-1"
                        >
                          <Plus className="h-3 w-3" />
                          {hasBudget ? 'Edit Target' : 'Set Budget'}
                        </Button>
                      </div>
                    </div>
                  )
                })}
              </div>

              {filteredBudgets.length === 0 && (
                <div className="py-12 text-center text-muted-foreground text-xs">
                  No categories match your search filter.
                </div>
              )}
            </CardContent>
          </Card>
        </>
      )}

      {/* SET / EDIT BUDGET MODAL */}
      <Dialog open={!!editingCategory} onOpenChange={(open) => !open && setEditingCategory(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="text-base font-semibold flex items-center gap-2">
              <Target className="h-4.5 w-4.5 text-primary" />
              Set Budget for {editingCategory?.name}
            </DialogTitle>
            <DialogDescription className="text-xs">
              Configure target monthly spending limit for {displayMonthName()}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2 text-xs">
            <div className="space-y-1.5">
              <label className="font-semibold text-foreground">Monthly Spending Limit (₹) *</label>
              <div className="relative">
                <DollarSign className="h-4 w-4 absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
                <Input
                  type="number"
                  step="500"
                  placeholder="e.g. 10000"
                  value={budgetLimitInput || ''}
                  onChange={(e) => setBudgetLimitInput(parseFloat(e.target.value) || 0)}
                  className="pl-8 font-mono font-bold text-sm h-9"
                />
              </div>
            </div>

            {/* Quick Preset Buttons */}
            <div className="space-y-1.5">
              <label className="text-[11px] text-muted-foreground">Quick Preset Limits:</label>
              <div className="grid grid-cols-3 gap-1.5">
                {[2000, 5000, 10000, 15000, 25000, 50000].map((preset) => (
                  <Button
                    key={preset}
                    variant={budgetLimitInput === preset ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setBudgetLimitInput(preset)}
                    className="text-xs font-mono h-7"
                  >
                    {formatINR(preset)}
                  </Button>
                ))}
              </div>
            </div>
          </div>

          <DialogFooter className="gap-2 pt-2">
            <Button variant="outline" size="sm" onClick={() => setEditingCategory(null)}>
              Cancel
            </Button>
            <Button
              size="sm"
              onClick={handleSaveBudget}
              disabled={upsertMutation.isPending || budgetLimitInput <= 0}
              className="font-semibold"
            >
              {upsertMutation.isPending ? 'Saving...' : 'Save Budget Target'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
