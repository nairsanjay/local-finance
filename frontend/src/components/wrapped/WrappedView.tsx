import React, { useState, useEffect, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import { fetchWrappedStory } from '@/lib/api'
import { formatINR, formatDate } from '@/lib/utils'
import { playSuccessChime, playSoftClick } from '@/lib/audio'
import {
  Sparkles,
  Trophy,
  Flame,
  ArrowRight,
  ArrowLeft,
  Calendar,
  Gift,
  ShieldCheck,
  Copy,
  Check,
  LayoutGrid,
  Maximize2,
  Store,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { MerchantAvatar } from '@/components/ui/merchant-avatar'

export const WrappedView: React.FC = () => {
  const [selectedYear, setSelectedYear] = useState<string>('latest')
  const [currentSlide, setCurrentSlide] = useState<number>(0)
  const [viewMode, setViewMode] = useState<'slides' | 'poster'>('slides')
  const [copied, setCopied] = useState(false)

  const { data: story, isLoading } = useQuery({
    queryKey: ['wrapped', selectedYear],
    queryFn: () => fetchWrappedStory(selectedYear === 'latest' ? undefined : selectedYear),
  })

  const totalSlides = 6

  const handleNext = useCallback(() => {
    setCurrentSlide((prev) => {
      const next = Math.min(prev + 1, totalSlides - 1)
      playSoftClick(550)
      if (next === 1 || next === totalSlides - 1) {
        playSuccessChime()
      }
      return next
    })
  }, [totalSlides])

  const handlePrev = useCallback(() => {
    setCurrentSlide((prev) => {
      const p = Math.max(prev - 1, 0)
      playSoftClick(480)
      return p
    })
  }, [])

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (viewMode !== 'slides') return
      if (e.key === 'ArrowRight' || e.key === 'Space') {
        e.preventDefault()
        handleNext()
      } else if (e.key === 'ArrowLeft') {
        e.preventDefault()
        handlePrev()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleNext, handlePrev, viewMode])

  // Copy text summary
  const copySummaryText = () => {
    if (!story) return
    const text = `🎉 My ${story.year} LocalFinance Wrapped:\n` +
      `• Persona: ${story.persona_title} (${story.persona_badge})\n` +
      `• Inflow: ${formatINR(story.total_income)}\n` +
      `• Net Savings Kept: ${formatINR(story.net_savings)} (${story.savings_rate.toFixed(1)}%)\n` +
      `• Crown Merchant: ${story.crown_merchant?.payee || 'N/A'} (${story.crown_merchant?.order_count || 0} visits)\n` +
      `• Free Cashback Earned: ${formatINR(story.total_cashback)}\n` +
      `🔒 100% computed locally & offline via LocalFinance.`

    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      playSuccessChime()
      setTimeout(() => setCopied(false), 2500)
    })
  }

  if (isLoading) {
    return (
      <div className="flex h-96 items-center justify-center text-muted-foreground text-sm">
        <Sparkles className="h-5 w-5 animate-pulse mr-2 text-primary" />
        Summoning your offline financial story...
      </div>
    )
  }

  if (!story || story.total_transactions === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-80 text-center p-6 space-y-4">
        <Gift className="h-10 w-10 text-muted-foreground/60" />
        <h3 className="text-lg font-bold text-foreground">No Transactions for this Period</h3>
        <p className="text-xs text-muted-foreground max-w-md">
          Import your bank statements or sample files first to generate your personalized offline Wrapped story.
        </p>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* Top Controls Bar */}
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-border/60 pb-4">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-amber-500/20 to-primary/20 text-amber-500 border border-amber-500/30 shadow-xs">
            <Trophy className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl sm:text-2xl font-black tracking-tight text-foreground">
              LocalFinance Wrapped {story.year}
            </h1>
            <p className="text-xs text-muted-foreground">
              Your offline financial year in review &amp; milestone recap
            </p>
          </div>
        </div>

        {/* Year Selector & View Toggle */}
        <div className="flex items-center gap-2.5">
          <Select value={selectedYear} onValueChange={(val) => val && setSelectedYear(val)}>
            <SelectTrigger className="h-8.5 w-32 text-xs font-semibold bg-card">
              <SelectValue placeholder="Year" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="latest">Latest Year</SelectItem>
              {story.available_years.map((y) => (
                <SelectItem key={y} value={y}>
                  Year {y}
                </SelectItem>
              ))}
              <SelectItem value="ALL">All-Time</SelectItem>
            </SelectContent>
          </Select>

          <div className="flex items-center rounded-lg border bg-muted/40 p-0.5">
            <Button
              size="sm"
              variant={viewMode === 'slides' ? 'secondary' : 'ghost'}
              onClick={() => {
                setViewMode('slides')
                playSoftClick()
              }}
              className="h-7 text-xs px-2.5"
            >
              <Maximize2 className="h-3.5 w-3.5 mr-1.5" />
              Story
            </Button>
            <Button
              size="sm"
              variant={viewMode === 'poster' ? 'secondary' : 'ghost'}
              onClick={() => {
                setViewMode('poster')
                playSoftClick()
              }}
              className="h-7 text-xs px-2.5"
            >
              <LayoutGrid className="h-3.5 w-3.5 mr-1.5" />
              Poster
            </Button>
          </div>
        </div>
      </div>

      {/* ================= VIEW MODE 1: INTERACTIVE SLIDES ================= */}
      {viewMode === 'slides' && (
        <div className="space-y-4">
          {/* Progress Indicator Dots */}
          <div className="flex items-center justify-between px-1">
            <div className="flex items-center gap-1.5">
              {Array.from({ length: totalSlides }).map((_, idx) => (
                <button
                  key={idx}
                  onClick={() => {
                    setCurrentSlide(idx)
                    playSoftClick()
                  }}
                  className={`h-1.5 rounded-full transition-all cursor-pointer ${
                    idx === currentSlide
                      ? 'w-8 bg-primary'
                      : idx < currentSlide
                      ? 'w-3 bg-primary/40'
                      : 'w-3 bg-muted'
                  }`}
                  aria-label={`Go to slide ${idx + 1}`}
                />
              ))}
            </div>

            <span className="text-[10px] font-mono text-muted-foreground font-semibold">
              Slide {currentSlide + 1} of {totalSlides}
            </span>
          </div>

          {/* Active Slide Card */}
          <Card className="border-border/80 bg-gradient-to-b from-card to-card/60 shadow-md min-h-[460px] flex flex-col justify-between overflow-hidden relative">
            <CardContent className="p-6 sm:p-10 flex-1 flex flex-col justify-center">
              {/* SLIDE 0: Money In & Out */}
              {currentSlide === 0 && (
                <div className="space-y-8 animate-in fade-in zoom-in-95 duration-300">
                  <div className="space-y-1">
                    <Badge variant="outline" className="text-[10px] uppercase font-mono tracking-wider">
                      The Big Numbers
                    </Badge>
                    <h2 className="text-2xl sm:text-4xl font-extrabold tracking-tight text-foreground">
                      In {story.year}, you handled{' '}
                      <span className="text-primary font-mono privacy-blur">
                        {formatINR(story.total_income + story.total_expense)}
                      </span>
                    </h2>
                    <p className="text-xs sm:text-sm text-muted-foreground">
                      Across {story.total_transactions} transactions recorded in your offline ledger.
                    </p>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                    <div className="rounded-xl border bg-emerald-500/10 border-emerald-500/20 p-4 space-y-1">
                      <span className="text-[10px] font-bold uppercase tracking-wider text-emerald-600 dark:text-emerald-400 block">
                        Total Inflow
                      </span>
                      <div className="text-xl sm:text-2xl font-bold font-mono text-foreground tabular-nums privacy-blur">
                        +{formatINR(story.total_income)}
                      </div>
                      <p className="text-[11px] text-muted-foreground">Salary &amp; Credits</p>
                    </div>

                    <div className="rounded-xl border bg-rose-500/10 border-rose-500/20 p-4 space-y-1">
                      <span className="text-[10px] font-bold uppercase tracking-wider text-rose-500 block">
                        Total Outflow
                      </span>
                      <div className="text-xl sm:text-2xl font-bold font-mono text-foreground tabular-nums privacy-blur">
                        -{formatINR(story.total_expense)}
                      </div>
                      <p className="text-[11px] text-muted-foreground">Debits &amp; Card Spends</p>
                    </div>

                    <div className="rounded-xl border bg-sky-500/10 border-sky-500/20 p-4 space-y-1">
                      <span className="text-[10px] font-bold uppercase tracking-wider text-sky-600 dark:text-sky-400 block">
                        Net Kept
                      </span>
                      <div className="text-xl sm:text-2xl font-bold font-mono text-emerald-500 tabular-nums privacy-blur">
                        {story.net_savings >= 0 ? '+' : ''}{formatINR(story.net_savings)}
                      </div>
                      <p className="text-[11px] text-muted-foreground font-mono">
                        {story.savings_rate.toFixed(1)}% Savings Rate
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {/* SLIDE 1: Persona Reveal */}
              {currentSlide === 1 && (
                <div className="space-y-6 text-center max-w-xl mx-auto animate-in fade-in zoom-in-95 duration-300">
                  <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-amber-500/15 border border-amber-500/30 text-amber-500 text-xs font-semibold">
                    <Sparkles className="h-3.5 w-3.5" />
                    <span>Your Financial Archetype</span>
                  </div>

                  <div className="space-y-2">
                    <h2 className="text-3xl sm:text-5xl font-black tracking-tight text-foreground">
                      {story.persona_title}
                    </h2>
                    <Badge variant="secondary" className="text-xs sm:text-sm font-semibold px-3 py-1 mt-1">
                      {story.persona_badge}
                    </Badge>
                  </div>

                  <p className="text-xs sm:text-sm text-muted-foreground leading-relaxed pt-2">
                    {story.persona_description}
                  </p>
                </div>
              )}

              {/* SLIDE 2: Crown Merchant */}
              {currentSlide === 2 && (
                <div className="space-y-6 animate-in fade-in zoom-in-95 duration-300">
                  <div className="space-y-1">
                    <Badge variant="outline" className="text-[10px] uppercase font-mono tracking-wider">
                      Favorite Destination
                    </Badge>
                    <h2 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-foreground">
                      Crown Merchant of {story.year}
                    </h2>
                  </div>

                  {story.crown_merchant ? (
                    <div className="rounded-2xl border border-amber-500/30 bg-gradient-to-r from-amber-500/10 via-card to-card p-6 flex flex-col sm:flex-row items-center gap-5">
                      <div className="h-16 w-16 shrink-0 rounded-2xl bg-amber-500/20 border border-amber-500/40 flex items-center justify-center text-3xl shadow-sm">
                        👑
                      </div>
                      <div className="space-y-1 text-center sm:text-left min-w-0">
                        <h3 className="text-xl font-bold text-foreground truncate">
                          {story.crown_merchant.payee}
                        </h3>
                        <p className="text-xs text-muted-foreground">
                          You patronized this merchant{' '}
                          <span className="font-bold text-foreground font-mono">
                            {story.crown_merchant.order_count} times
                          </span>{' '}
                          totaling{' '}
                          <span className="font-bold text-foreground font-mono privacy-blur">
                            {formatINR(story.crown_merchant.total_spent)}
                          </span>
                          .
                        </p>
                      </div>
                    </div>
                  ) : (
                    <p className="text-xs text-muted-foreground">No merchant orders found</p>
                  )}

                  {/* Top 4 runner-up merchants */}
                  {story.top_merchants.length > 1 && (
                    <div className="space-y-2 pt-2">
                      <span className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground block">
                        Top Runner-Ups
                      </span>
                      <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
                        {story.top_merchants.slice(1, 5).map((m, idx) => (
                          <div
                            key={m.payee}
                            className="rounded-lg border bg-background/50 p-2.5 text-xs space-y-1"
                          >
                            <span className="text-[10px] font-mono text-muted-foreground font-semibold">
                              #{idx + 2}
                            </span>
                            <p className="font-bold text-foreground truncate">{m.payee}</p>
                            <p className="text-[11px] font-mono text-muted-foreground privacy-blur">
                              {formatINR(m.total_spent)}
                            </p>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )}

              {/* SLIDE 3: Peak Moments */}
              {currentSlide === 3 && (
                <div className="space-y-6 animate-in fade-in zoom-in-95 duration-300">
                  <div className="space-y-1">
                    <Badge variant="outline" className="text-[10px] uppercase font-mono tracking-wider">
                      Extreme Moments
                    </Badge>
                    <h2 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-foreground">
                      Peak Spending Moments
                    </h2>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    {/* Biggest Purchase */}
                    <div className="rounded-xl border bg-card p-5 space-y-3">
                      <div className="flex items-center justify-between">
                        <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground flex items-center gap-1.5">
                          <Flame className="h-3.5 w-3.5 text-rose-500" />
                          Biggest Single Purchase
                        </span>
                        {story.biggest_purchase && (
                          <Badge variant="outline" className="text-[10px] font-mono">
                            {formatDate(story.biggest_purchase.tx_date)}
                          </Badge>
                        )}
                      </div>

                      {story.biggest_purchase ? (
                        <div>
                          <div className="text-2xl font-extrabold font-mono text-foreground tabular-nums privacy-blur">
                            {formatINR(story.biggest_purchase.amount)}
                          </div>
                          <p className="text-xs font-semibold text-foreground truncate mt-1">
                            {story.biggest_purchase.cleaned_payee}
                          </p>
                          <p className="text-[10px] font-mono text-muted-foreground truncate">
                            {story.biggest_purchase.payment_mode} • {story.biggest_purchase.raw_narration}
                          </p>
                        </div>
                      ) : (
                        <p className="text-xs text-muted-foreground">No purchases recorded</p>
                      )}
                    </div>

                    {/* Busiest Day */}
                    <div className="rounded-xl border bg-card p-5 space-y-3">
                      <div className="flex items-center justify-between">
                        <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground flex items-center gap-1.5">
                          <Calendar className="h-3.5 w-3.5 text-primary" />
                          Busiest Spending Day
                        </span>
                        {story.busiest_day && (
                          <Badge variant="outline" className="text-[10px] font-mono">
                            {story.busiest_day_tx_count} orders
                          </Badge>
                        )}
                      </div>

                      {story.busiest_day ? (
                        <div>
                          <div className="text-2xl font-extrabold font-mono text-foreground tabular-nums privacy-blur">
                            {formatINR(story.busiest_day_spend || 0)}
                          </div>
                          <p className="text-xs font-semibold text-foreground mt-1">
                            {formatDate(story.busiest_day)}
                          </p>
                          <p className="text-[10px] text-muted-foreground">
                            Highest cumulative debit volume recorded on a single day.
                          </p>
                        </div>
                      ) : (
                        <p className="text-xs text-muted-foreground">No data</p>
                      )}
                    </div>
                  </div>
                </div>
              )}

              {/* SLIDE 4: Free Money & Cashbacks */}
              {currentSlide === 4 && (
                <div className="space-y-6 animate-in fade-in zoom-in-95 duration-300">
                  <div className="space-y-1">
                    <Badge variant="outline" className="text-[10px] uppercase font-mono tracking-wider">
                      Financial Float
                    </Badge>
                    <h2 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-foreground">
                      Free Money &amp; Rewards Unlocked
                    </h2>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                    <div className="rounded-xl border bg-emerald-500/10 border-emerald-500/20 p-4 space-y-1">
                      <span className="text-[10px] font-bold uppercase tracking-wider text-emerald-600 dark:text-emerald-400 block">
                        Cashback Credited
                      </span>
                      <div className="text-2xl font-bold font-mono text-foreground tabular-nums privacy-blur">
                        +{formatINR(story.total_cashback)}
                      </div>
                      <p className="text-[11px] text-muted-foreground">Credited directly to bills</p>
                    </div>

                    <div className="rounded-xl border bg-primary/10 border-primary/20 p-4 space-y-1">
                      <span className="text-[10px] font-bold uppercase tracking-wider text-primary block">
                        Reward Points
                      </span>
                      <div className="text-2xl font-bold font-mono text-foreground tabular-nums privacy-blur">
                        {story.total_reward_points.toLocaleString()} pts
                      </div>
                      <p className="text-[11px] text-muted-foreground">Earned across credit cards</p>
                    </div>

                    <div className="rounded-xl border bg-muted/40 p-4 space-y-1">
                      <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground block">
                        UPI vs Cards
                      </span>
                      <div className="text-xs font-mono text-foreground space-y-1 pt-1">
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">UPI Txs:</span>
                          <span className="font-bold">{story.upi_tx_count}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">Card Swipes:</span>
                          <span className="font-bold">{story.card_tx_count}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* SLIDE 5: The Grand Wrapped Summary */}
              {currentSlide === 5 && (
                <div className="space-y-6 animate-in fade-in zoom-in-95 duration-300">
                  <div className="flex items-center justify-between border-b pb-3">
                    <div>
                      <Badge variant="outline" className="text-[10px] uppercase font-mono tracking-wider">
                        Complete Recap
                      </Badge>
                      <h2 className="text-2xl font-extrabold tracking-tight text-foreground">
                        {story.year} Financial Snapshot
                      </h2>
                    </div>

                    <Button
                      size="sm"
                      variant="outline"
                      onClick={copySummaryText}
                      className="gap-1.5 text-xs font-semibold shadow-xs"
                    >
                      {copied ? <Check className="h-3.5 w-3.5 text-emerald-500" /> : <Copy className="h-3.5 w-3.5" />}
                      <span>{copied ? 'Copied!' : 'Copy Summary'}</span>
                    </Button>
                  </div>

                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
                    <div className="rounded-lg border bg-muted/20 p-3 space-y-1">
                      <span className="text-[10px] text-muted-foreground uppercase font-bold">Archetype</span>
                      <p className="font-bold text-foreground truncate">{story.persona_title}</p>
                    </div>
                    <div className="rounded-lg border bg-muted/20 p-3 space-y-1">
                      <span className="text-[10px] text-muted-foreground uppercase font-bold">Net Kept</span>
                      <p className="font-bold font-mono text-emerald-500 privacy-blur">{formatINR(story.net_savings)}</p>
                    </div>
                    <div className="rounded-lg border bg-muted/20 p-3 space-y-1">
                      <span className="text-[10px] text-muted-foreground uppercase font-bold">Savings %</span>
                      <p className="font-bold font-mono text-foreground">{story.savings_rate.toFixed(1)}%</p>
                    </div>
                    <div className="rounded-lg border bg-muted/20 p-3 space-y-1">
                      <span className="text-[10px] text-muted-foreground uppercase font-bold">#1 Merchant</span>
                      <p className="font-bold text-foreground truncate">{story.crown_merchant?.payee || 'None'}</p>
                    </div>
                  </div>

                  {/* 100% Offline Privacy Guarantee Ribbon */}
                  <div className="rounded-xl border border-emerald-500/20 bg-emerald-500/5 p-4 flex items-center gap-3 text-xs text-foreground">
                    <ShieldCheck className="h-5 w-5 text-emerald-500 shrink-0" />
                    <div>
                      <span className="font-bold block">100% Offline Computation</span>
                      <p className="text-[11px] text-muted-foreground">
                        This story was calculated entirely on your local machine in milliseconds. Zero bytes were transmitted across the internet.
                      </p>
                    </div>
                  </div>
                </div>
              )}
            </CardContent>

            {/* Slide Navigation Footer Bar */}
            <div className="flex items-center justify-between border-t border-border/60 bg-muted/20 px-6 py-3">
              <Button
                variant="ghost"
                size="sm"
                onClick={handlePrev}
                disabled={currentSlide === 0}
                className="gap-1 text-xs"
              >
                <ArrowLeft className="h-3.5 w-3.5" />
                <span>Previous</span>
              </Button>

              <div className="flex items-center gap-2">
                <span className="text-[10px] text-muted-foreground hidden sm:inline">
                  Use arrow keys to navigate
                </span>
                {currentSlide < totalSlides - 1 ? (
                  <Button
                    size="sm"
                    onClick={handleNext}
                    className="gap-1 text-xs font-semibold shadow-xs"
                  >
                    <span>Next</span>
                    <ArrowRight className="h-3.5 w-3.5" />
                  </Button>
                ) : (
                  <Button
                    size="sm"
                    variant="default"
                    onClick={() => {
                      setCurrentSlide(0)
                      playSoftClick()
                    }}
                    className="gap-1 text-xs font-semibold shadow-xs"
                  >
                    <span>Replay Story</span>
                  </Button>
                )}
              </div>
            </div>
          </Card>
        </div>
      )}

      {/* ================= VIEW MODE 2: FULL POSTER ================= */}
      {viewMode === 'poster' && (
        <div className="rounded-2xl border bg-card p-6 sm:p-8 shadow-sm space-y-8">
          <div className="flex flex-wrap items-center justify-between gap-4 border-b pb-4">
            <div>
              <div className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-amber-500/15 text-amber-500 text-[11px] font-semibold mb-1">
                <Trophy className="h-3 w-3" />
                <span>{story.persona_badge}</span>
              </div>
              <h2 className="text-2xl sm:text-3xl font-black text-foreground">
                {story.persona_title}
              </h2>
              <p className="text-xs text-muted-foreground mt-0.5">
                {story.year} Financial Year-in-Review
              </p>
            </div>

            <Button
              size="sm"
              variant="outline"
              onClick={copySummaryText}
              className="gap-1.5 text-xs font-semibold shadow-xs"
            >
              {copied ? <Check className="h-3.5 w-3.5 text-emerald-500" /> : <Copy className="h-3.5 w-3.5" />}
              <span>{copied ? 'Copied' : 'Copy Summary'}</span>
            </Button>
          </div>

          {/* Key Metric Tiles */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
            <div className="rounded-xl border bg-muted/30 p-4 space-y-1">
              <span className="text-[10px] uppercase font-bold text-muted-foreground">Total Income</span>
              <div className="text-xl font-bold font-mono text-foreground privacy-blur">
                +{formatINR(story.total_income)}
              </div>
            </div>
            <div className="rounded-xl border bg-muted/30 p-4 space-y-1">
              <span className="text-[10px] uppercase font-bold text-muted-foreground">Total Spent</span>
              <div className="text-xl font-bold font-mono text-foreground privacy-blur">
                -{formatINR(story.total_expense)}
              </div>
            </div>
            <div className="rounded-xl border bg-muted/30 p-4 space-y-1">
              <span className="text-[10px] uppercase font-bold text-muted-foreground">Net Kept</span>
              <div className="text-xl font-bold font-mono text-emerald-500 privacy-blur">
                +{formatINR(story.net_savings)}
              </div>
            </div>
            <div className="rounded-xl border bg-muted/30 p-4 space-y-1">
              <span className="text-[10px] uppercase font-bold text-muted-foreground">Savings Rate</span>
              <div className="text-xl font-bold font-mono text-foreground">
                {story.savings_rate.toFixed(1)}%
              </div>
            </div>
          </div>

          {/* Crown Merchant & Biggest Purchase Row */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {story.crown_merchant && (
              <div className="rounded-xl border p-4 space-y-2">
                <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground flex items-center gap-1.5">
                  <Store className="h-3.5 w-3.5 text-amber-500" />
                  Crown Merchant
                </span>
                <div className="flex items-center gap-3">
                  <MerchantAvatar payee={story.crown_merchant.payee} size="md" />
                  <div>
                    <h4 className="font-bold text-sm text-foreground">{story.crown_merchant.payee}</h4>
                    <p className="text-xs font-mono text-muted-foreground">
                      {story.crown_merchant.order_count} visits • <span className="privacy-blur">{formatINR(story.crown_merchant.total_spent)}</span>
                    </p>
                  </div>
                </div>
              </div>
            )}

            {story.biggest_purchase && (
              <div className="rounded-xl border p-4 space-y-2">
                <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground flex items-center gap-1.5">
                  <Flame className="h-3.5 w-3.5 text-rose-500" />
                  Largest Single Purchase
                </span>
                <div>
                  <h4 className="font-bold text-sm text-foreground truncate">
                    {story.biggest_purchase.cleaned_payee}
                  </h4>
                  <p className="text-xs font-mono text-muted-foreground">
                    <span className="font-bold text-foreground privacy-blur">{formatINR(story.biggest_purchase.amount)}</span> on {formatDate(story.biggest_purchase.tx_date)}
                  </p>
                </div>
              </div>
            )}
          </div>

          {/* Top Spending Categories */}
          {story.top_categories.length > 0 && (
            <div className="space-y-3">
              <span className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
                Top Spending Categories
              </span>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                {story.top_categories.slice(0, 3).map((c) => (
                  <div key={c.category_id} className="rounded-xl border p-3.5 space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="font-semibold text-xs text-foreground truncate">{c.category_name}</span>
                      <Badge variant="secondary" className="text-[10px] font-mono">
                        {c.percentage.toFixed(1)}%
                      </Badge>
                    </div>
                    <div className="text-base font-bold font-mono text-foreground privacy-blur">
                      {formatINR(c.total_amount)}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Footer Note */}
          <div className="border-t pt-4 text-center text-xs text-muted-foreground flex items-center justify-center gap-2">
            <ShieldCheck className="h-4 w-4 text-emerald-500" />
            <span>LocalFinance • 100% Offline Local Intelligence</span>
          </div>
        </div>
      )}
    </div>
  )
}
