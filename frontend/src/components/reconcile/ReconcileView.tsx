import React, { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  fetchReconciliationSummary,
  scanReconciliation,
  linkTransferPair,
  unlinkTransferPair,
  toggleExcludeTransaction,
} from '../../lib/api'
import { formatINR } from '../../lib/utils'
import type { TransferPair } from '../../types'
import {
  Card,
  CardContent,
} from '../ui/card'
import { Button } from '../ui/button'
import { Badge } from '../ui/badge'
import { MerchantAvatar } from '../ui/merchant-avatar'
import {
  ArrowLeftRight,
  ShieldCheck,
  Sparkles,
  Ban,
  CheckCircle2,
  RefreshCw,
  Wallet,
  Landmark,
  CreditCard,
  Unlink,
  Check,
  HelpCircle,
} from 'lucide-react'

export const ReconcileView: React.FC = () => {
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<'CANDIDATES' | 'PAIRED' | 'WALLET'>('CANDIDATES')
  const [notification, setNotification] = useState<string | null>(null)

  const { data: summary, isLoading, isError } = useQuery({
    queryKey: ['reconciliation-summary'],
    queryFn: fetchReconciliationSummary,
  })

  // Mutations
  const scanMutation = useMutation({
    mutationFn: scanReconciliation,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['reconciliation-summary'] })
      queryClient.invalidateQueries({ queryKey: ['analytics'] })
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
      setNotification(data.message)
      setTimeout(() => setNotification(null), 5000)
    },
  })

  const linkMutation = useMutation({
    mutationFn: (pair: { debit_tx_id: string; credit_tx_id: string; match_reason?: string }) =>
      linkTransferPair(pair),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['reconciliation-summary'] })
      queryClient.invalidateQueries({ queryKey: ['analytics'] })
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
    },
  })

  const unlinkMutation = useMutation({
    mutationFn: (txId: string) => unlinkTransferPair(txId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['reconciliation-summary'] })
      queryClient.invalidateQueries({ queryKey: ['analytics'] })
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
    },
  })

  const toggleExcludeMutation = useMutation({
    mutationFn: (txId: string) => toggleExcludeTransaction(txId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['reconciliation-summary'] })
      queryClient.invalidateQueries({ queryKey: ['analytics'] })
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
    },
  })

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center text-muted-foreground text-sm">
        Analyzing transactions for internal transfers and reconciliations...
      </div>
    )
  }

  if (isError || !summary) {
    return (
      <div className="rounded-xl border border-destructive/30 bg-destructive/10 p-8 text-center text-destructive">
        Failed to load transfer reconciliation data.
      </div>
    )
  }

  const candidates = summary.candidates || []
  const pairs = summary.pairs || []
  const walletTransactions = summary.wallet_transactions || []

  return (
    <div className="space-y-8 max-w-7xl mx-auto pb-12">
      {/* Header Section */}
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
              Transfer Reconciler
            </h1>
            <Badge variant="secondary" className="gap-1 font-semibold text-xs py-0.5">
              <ShieldCheck className="h-3.5 w-3.5 text-emerald-500" />
              Double-Count Protection
            </Badge>
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            Pair transfers between your bank accounts and credit card payments. Shared payment references identify bank transfers automatically; missing or ambiguous references require your review.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button
            onClick={() => scanMutation.mutate()}
            disabled={scanMutation.isPending}
            className="gap-2 font-semibold shadow-xs"
          >
            <RefreshCw className={`h-4 w-4 ${scanMutation.isPending ? 'animate-spin' : ''}`} />
            <span>{scanMutation.isPending ? 'Scanning Ledger...' : 'Scan & Auto-Pair'}</span>
          </Button>
        </div>
      </div>

      {/* Notification Toast Banner */}
      {notification && (
        <div className="flex items-center gap-3 rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-4 text-sm text-emerald-600 dark:text-emerald-400 animate-in fade-in slide-in-from-top-2">
          <CheckCircle2 className="h-5 w-5 shrink-0" />
          <span className="font-medium">{notification}</span>
        </div>
      )}

      {/* KPI Ribbon */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card className="border-border shadow-xs bg-card/60 backdrop-blur-xs">
          <CardContent className="p-5">
            <div className="flex items-center justify-between">
              <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                Double-Count Prevented
              </p>
              <div className="rounded-md bg-emerald-500/10 p-2 text-emerald-500">
                <ShieldCheck className="h-4 w-4" />
              </div>
            </div>
            <p className="mt-2 text-2xl font-bold tracking-tight text-foreground">
              {formatINR(summary.double_count_prevented_amount || 0)}
            </p>
            <p className="mt-1 text-[11px] text-muted-foreground">
              Across {summary.total_paired_transfers || 0} linked internal transfers
            </p>
          </CardContent>
        </Card>

        <Card className="border-border shadow-xs bg-card/60 backdrop-blur-xs">
          <CardContent className="p-5">
            <div className="flex items-center justify-between">
              <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                Pending Pair Candidates
              </p>
              <div className="rounded-md bg-amber-500/10 p-2 text-amber-500">
                <Sparkles className="h-4 w-4" />
              </div>
            </div>
            <p className="mt-2 text-2xl font-bold tracking-tight text-foreground">
              {summary.pending_candidates_count || 0}
            </p>
            <p className="mt-1 text-[11px] text-muted-foreground">
              Potential unlinked bill payments found
            </p>
          </CardContent>
        </Card>

        <Card className="border-border shadow-xs bg-card/60 backdrop-blur-xs">
          <CardContent className="p-5">
            <div className="flex items-center justify-between">
              <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                Linked CC Payments
              </p>
              <div className="rounded-md bg-blue-500/10 p-2 text-blue-500">
                <ArrowLeftRight className="h-4 w-4" />
              </div>
            </div>
            <p className="mt-2 text-2xl font-bold tracking-tight text-foreground">
              {summary.total_paired_transfers || 0}
            </p>
            <p className="mt-1 text-[11px] text-muted-foreground">
              Marked with is_transfer = 1
            </p>
          </CardContent>
        </Card>

        <Card className="border-border shadow-xs bg-card/60 backdrop-blur-xs">
          <CardContent className="p-5">
            <div className="flex items-center justify-between">
              <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                Excluded Wallet Loads
              </p>
              <div className="rounded-md bg-purple-500/10 p-2 text-purple-500">
                <Wallet className="h-4 w-4" />
              </div>
            </div>
            <p className="mt-2 text-2xl font-bold tracking-tight text-foreground">
              {formatINR(summary.wallet_excluded_amount || 0)}
            </p>
            <p className="mt-1 text-[11px] text-muted-foreground">
              {summary.wallet_excluded_count || 0} UPI Lite & wallet top-up transactions
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Tabs Navigation */}
      <div className="flex border-b border-border overflow-x-auto">
        <button
          onClick={() => setActiveTab('CANDIDATES')}
          className={`flex items-center gap-2 border-b-2 px-4 sm:px-5 py-3 text-xs font-semibold transition-colors shrink-0 ${
            activeTab === 'CANDIDATES'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
        >
          <Sparkles className="h-4 w-4" />
          <span>Candidate Matches to Pair</span>
          {(summary.pending_candidates_count || 0) > 0 && (
            <Badge variant="default" className="ml-1 px-1.5 py-0 text-[10px] font-bold">
              {summary.pending_candidates_count}
            </Badge>
          )}
        </button>

        <button
          onClick={() => setActiveTab('PAIRED')}
          className={`flex items-center gap-2 border-b-2 px-4 sm:px-5 py-3 text-xs font-semibold transition-colors shrink-0 ${
            activeTab === 'PAIRED'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
        >
          <ArrowLeftRight className="h-4 w-4" />
          <span>Linked Internal Transfers ({summary.total_paired_transfers || 0})</span>
        </button>

        <button
          onClick={() => setActiveTab('WALLET')}
          className={`flex items-center gap-2 border-b-2 px-4 sm:px-5 py-3 text-xs font-semibold transition-colors shrink-0 ${
            activeTab === 'WALLET'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
        >
          <Ban className="h-4 w-4" />
          <span>Excluded Wallet Loads ({summary.wallet_excluded_count || 0})</span>
        </button>
      </div>

      {/* Tab 1: Candidates to Pair */}
      {activeTab === 'CANDIDATES' && (
        <div className="space-y-4">
          {candidates.length === 0 ? (
            <Card className="border-border bg-card/40 p-12 text-center">
              <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-500 mb-3">
                <CheckCircle2 className="h-6 w-6" />
              </div>
              <h3 className="text-base font-semibold text-foreground">All Set! No Unlinked Bill Payments</h3>
              <p className="text-xs text-muted-foreground mt-1 max-w-md mx-auto">
                All confirmed transfers have been matched and reconciled.
              </p>
            </Card>
          ) : (
            <div className="grid grid-cols-1 gap-4">
              {candidates.map((pair) => (
                <CandidatePairCard
                  key={pair.id}
                  pair={pair}
                  onLink={() =>
                    linkMutation.mutate({
                      debit_tx_id: pair.debit_tx.id,
                      credit_tx_id: pair.credit_tx.id,
                      match_reason: pair.match_reason,
                    })
                  }
                  isLinking={linkMutation.isPending}
                />
              ))}
            </div>
          )}
        </div>
      )}

      {/* Tab 2: Linked Transfers */}
      {activeTab === 'PAIRED' && (
        <div className="space-y-4">
          {pairs.length === 0 ? (
            <Card className="border-border bg-card/40 p-12 text-center">
              <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-muted text-muted-foreground mb-3">
                <HelpCircle className="h-6 w-6" />
              </div>
              <h3 className="text-base font-semibold text-foreground">No Linked Transfers Yet</h3>
              <p className="text-xs text-muted-foreground mt-1 max-w-md mx-auto">
                Click &quot;Scan &amp; Auto-Pair&quot; to detect bank transfers and credit card payments.
              </p>
            </Card>
          ) : (
            <div className="grid grid-cols-1 gap-4">
              {pairs.map((pair) => (
                <PairedTransferCard
                  key={pair.id}
                  pair={pair}
                  onUnlink={() => unlinkMutation.mutate(pair.debit_tx.id)}
                  isUnlinking={unlinkMutation.isPending}
                />
              ))}
            </div>
          )}
        </div>
      )}

      {/* Tab 3: Wallet Loads */}
      {activeTab === 'WALLET' && (
        <div className="space-y-4">
          {walletTransactions.length === 0 ? (
            <Card className="border-border bg-card/40 p-12 text-center">
              <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-muted text-muted-foreground mb-3">
                <Wallet className="h-6 w-6" />
              </div>
              <h3 className="text-base font-semibold text-foreground">No Wallet Load Transactions Found</h3>
              <p className="text-xs text-muted-foreground mt-1 max-w-md mx-auto">
                Transactions from UPI Lite or Paytm Wallet top-ups will appear here to prevent double counting.
              </p>
            </Card>
          ) : (
            <div className="rounded-xl border border-border bg-card overflow-hidden">
              <div className="divide-y divide-border">
                {walletTransactions.map((tx) => (
                  <div
                    key={tx.id}
                    className="flex flex-wrap items-center justify-between p-4 hover:bg-muted/30 transition-colors gap-3"
                  >
                    <div className="flex items-center gap-3">
                      <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-purple-500/10 text-purple-500">
                        <Wallet className="h-4 w-4" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-semibold text-sm text-foreground">
                            {tx.cleaned_payee || tx.raw_narration}
                          </span>
                          <Badge variant="outline" className="text-[10px] font-mono">
                            {tx.tx_date}
                          </Badge>
                          {tx.is_excluded && (
                            <Badge variant="secondary" className="text-[10px] text-amber-500 bg-amber-500/10 font-semibold">
                              Excluded from Spend
                            </Badge>
                          )}
                        </div>
                        <p className="text-xs text-muted-foreground mt-0.5 font-mono truncate max-w-lg">
                          {tx.raw_narration} • {tx.account_name}
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center gap-4">
                      <span className="text-sm font-bold text-foreground">
                        {formatINR(tx.amount)}
                      </span>
                      <Button
                        size="sm"
                        variant={tx.is_excluded ? 'outline' : 'secondary'}
                        onClick={() => toggleExcludeMutation.mutate(tx.id)}
                        disabled={toggleExcludeMutation.isPending}
                        className="text-xs gap-1.5"
                      >
                        <Ban className="h-3.5 w-3.5" />
                        <span>{tx.is_excluded ? 'Include in Expenses' : 'Exclude'}</span>
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// Candidate Match Card Component
const CandidatePairCard: React.FC<{
  pair: TransferPair
  onLink: () => void
  isLinking: boolean
}> = ({ pair, onLink, isLinking }) => {
  const confidencePercent = Math.round(pair.match_confidence * 100)

  return (
    <Card className="border-border shadow-xs bg-card/80 overflow-hidden">
      <div className="p-5">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border pb-3 mb-4">
          <div className="flex items-center gap-2">
            <Badge
              variant="outline"
              className={`text-xs font-semibold px-2 py-0.5 ${
                confidencePercent >= 90
                  ? 'border-emerald-500/50 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
                  : 'border-amber-500/50 bg-amber-500/10 text-amber-600 dark:text-amber-400'
              }`}
            >
              <Sparkles className="h-3 w-3 mr-1 inline" />
              {pair.match_reason === 'OWN_ACCOUNT_REFERENCE_MATCH'
                ? 'Shared payment reference'
                : pair.match_reason?.startsWith('OWN_ACCOUNT_')
                  ? 'Review required: reference missing or ambiguous'
                  : `${confidencePercent}% Match Confidence`}
            </Badge>
            <span className="text-xs text-muted-foreground font-medium">
              {pair.date_difference_days === 0
                ? 'Same-day transaction'
                : `${pair.date_difference_days} day difference`}
            </span>
          </div>

          <Button
            size="sm"
            onClick={onLink}
            disabled={isLinking}
            className="gap-2 font-semibold text-xs bg-emerald-600 hover:bg-emerald-700 text-white"
          >
            <Check className="h-3.5 w-3.5" />
            <span>Confirm &amp; Link Transfer</span>
          </Button>
        </div>

        <div className="relative grid grid-cols-1 md:grid-cols-2 gap-4 items-stretch">
          {/* Debit Bank Side */}
          <div className="rounded-xl border border-rose-500/20 bg-rose-500/5 p-4 space-y-2.5 transition-all group-hover:border-rose-500/40">
            <div className="flex items-center justify-between">
              <span className="text-[11px] font-semibold uppercase tracking-wider text-rose-500 flex items-center gap-1.5">
                <Landmark className="h-3.5 w-3.5" /> Bank Outflow
              </span>
              <span className="text-xs font-mono text-muted-foreground">{pair.debit_tx.tx_date}</span>
            </div>
            <div className="flex items-center justify-between gap-2">
              <div className="flex items-center gap-2.5 min-w-0">
                <MerchantAvatar payee={pair.debit_tx.cleaned_payee || pair.debit_tx.raw_narration} size="sm" />
                <div className="min-w-0">
                  <p className="font-semibold text-sm text-foreground truncate">
                    {pair.debit_tx.cleaned_payee || pair.debit_tx.raw_narration}
                  </p>
                  <p className="text-xs text-muted-foreground mt-0.5 truncate">{pair.debit_tx.account_name}</p>
                </div>
              </div>
              <p className="text-base font-bold font-mono text-rose-500 shrink-0 tabular-nums">
                - {formatINR(pair.debit_tx.amount)}
              </p>
            </div>
            <p className="text-[11px] font-mono text-muted-foreground/80 truncate pt-1 border-t border-rose-500/10">
              {pair.debit_tx.raw_narration}
            </p>
          </div>

          {/* Credit Card Side */}
          <div className="rounded-xl border border-emerald-500/20 bg-emerald-500/5 p-4 space-y-2.5 transition-all group-hover:border-emerald-500/40">
            <div className="flex items-center justify-between">
              <span className="text-[11px] font-semibold uppercase tracking-wider text-emerald-500 flex items-center gap-1.5">
                <CreditCard className="h-3.5 w-3.5" /> Account Inflow
              </span>
              <span className="text-xs font-mono text-muted-foreground">{pair.credit_tx.tx_date}</span>
            </div>
            <div className="flex items-center justify-between gap-2">
              <div className="flex items-center gap-2.5 min-w-0">
                <MerchantAvatar payee={pair.credit_tx.cleaned_payee || pair.credit_tx.raw_narration} size="sm" />
                <div className="min-w-0">
                  <p className="font-semibold text-sm text-foreground truncate">
                    {pair.credit_tx.cleaned_payee || pair.credit_tx.raw_narration}
                  </p>
                  <p className="text-xs text-muted-foreground mt-0.5 truncate">{pair.credit_tx.account_name}</p>
                </div>
              </div>
              <p className="text-base font-bold font-mono text-emerald-500 shrink-0 tabular-nums">
                + {formatINR(pair.credit_tx.amount)}
              </p>
            </div>
            <p className="text-[11px] font-mono text-muted-foreground/80 truncate pt-1 border-t border-emerald-500/10">
              {pair.credit_tx.raw_narration}
            </p>
          </div>
        </div>
      </div>
    </Card>
  )
}

// Paired Transfer Card Component
const PairedTransferCard: React.FC<{
  pair: TransferPair
  onUnlink: () => void
  isUnlinking: boolean
}> = ({ pair, onUnlink, isUnlinking }) => {
  return (
    <Card className="border-border shadow-xs bg-card/60">
      <div className="p-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
              <ArrowLeftRight className="h-4 w-4" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-semibold text-sm text-foreground">
                  {pair.debit_tx.account_name} ➔ {pair.credit_tx.account_name}
                </span>
                <Badge variant="outline" className="text-[10px] text-emerald-500 border-emerald-500/30">
                  Linked Transfer
                </Badge>
              </div>
              <p className="text-xs text-muted-foreground mt-0.5">
                {pair.debit_tx.tx_date} • {pair.debit_tx.raw_narration}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <span className="text-base font-bold text-foreground">
              {formatINR(pair.debit_tx.amount)}
            </span>
            <Button
              size="sm"
              variant="outline"
              onClick={onUnlink}
              disabled={isUnlinking}
              className="text-xs gap-1.5 text-muted-foreground hover:text-destructive"
            >
              <Unlink className="h-3.5 w-3.5" />
              <span>Unlink</span>
            </Button>
          </div>
        </div>
      </div>
    </Card>
  )
}
