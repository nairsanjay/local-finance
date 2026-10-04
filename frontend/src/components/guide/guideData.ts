import React from 'react'
import {
  UploadCloud,
  Layers,
  ReceiptText,
  Target,
  CreditCard,
  Sparkles,
  Store,
  ArrowLeftRight,
  Settings,
  Building2,
} from 'lucide-react'

export interface GuideStep {
  number: number
  title: string
  content: string
  proTip?: string
  calloutType?: 'tip' | 'warning' | 'security' | 'info'
  sampleData?: {
    label: string
    rows: { key: string; value: string }[]
  }
}

export interface GuideTopic {
  id: string
  title: string
  category: 'Getting Started' | 'Financial Intelligence' | 'Daily Tracking' | 'System & Storage'
  shortTitle: string
  readTime: string
  icon: React.ComponentType<{ className?: string }>
  iconColor: string
  summary: string
  ctaText: string
  ctaUrl: string
  steps: GuideStep[]
  footerNotes?: string
}

export const GUIDE_CATEGORIES = [
  'Getting Started',
  'Financial Intelligence',
  'Daily Tracking',
  'System & Storage',
] as const

export const GUIDE_TOPICS: GuideTopic[] = [
  {
    id: 'providers',
    title: 'Supported Banks & Statement Formats',
    shortTitle: 'Supported Providers',
    category: 'Getting Started',
    readTime: '3 min read',
    icon: Building2,
    iconColor: '#3B82F6',
    summary:
      'Complete catalog of all Indian banks, credit cards, file formats, and password recipes natively supported by the LocalFinance offline parser engine.',
    ctaText: 'Test Statement Ingestion',
    ctaUrl: '/import',
    steps: [
      {
        number: 1,
        title: 'Overview of Indian Bank Statement Support',
        content:
          'LocalFinance includes a modular bank adapter registry that automatically inspects file headers, byte signatures, and table column layouts to detect the bank issuer and statement type with high confidence. No cloud parser or third-party scraper is ever contacted.',
        calloutType: 'security',
        proTip:
          'All PDF password decryption happens in volatile memory using pure Go. Your passwords and PDF contents are never written to disk or sent across a network.',
      },
      {
        number: 2,
        title: 'HDFC Bank (Savings, Current & Credit Cards)',
        content:
          'HDFC Bank statements are supported in PDF, Excel (.xls/.xlsx), and CSV formats across both net banking downloads and email statements.',
        sampleData: {
          label: 'HDFC Bank Formats & Passwords',
          rows: [
            { key: 'Savings Account (PDF)', value: 'Password is your 8-to-9 digit NetBanking Customer ID.' },
            { key: 'Savings Account (Excel / CSV)', value: 'Exported directly from HDFC NetBanking > Enquire > Download Statement.' },
            { key: 'Credit Card (PDF)', value: 'First 4 letters of your name (in lowercase) followed by DOB in DDMM format (e.g., rahul1508).' },
            { key: 'Credit Card (CSV)', value: 'Exported from HDFC Cards NetBanking portal without password.' },
          ],
        },
      },
      {
        number: 3,
        title: 'ICICI Bank (Credit Cards & Savings)',
        content:
          'ICICI Bank statements in PDF and CSV format are sniffed automatically, with native support for popular co-branded cards like Amazon Pay ICICI, Coral, Sapphiro, and Rubyx.',
        sampleData: {
          label: 'ICICI Bank Formats & Passwords',
          rows: [
            { key: 'Credit Card (PDF)', value: 'First 4 characters of name in lowercase + Date & Month of Birth in DDMM (e.g. amit0407).' },
            { key: 'Amazon Pay ICICI Card', value: 'Supported natively including cashback earned metrics and statement billing dates.' },
            { key: 'Account CSV', value: 'Direct CSV download from ICICI Internet Banking > Detailed Statement.' },
          ],
        },
      },
      {
        number: 4,
        title: 'Axis Bank (Credit Cards & Savings)',
        content:
          'Axis Bank statements (including Axis Flipkart, Ace, Magnus, Atlas, and Neo) are recognized via transaction line signatures and summary blocks.',
        sampleData: {
          label: 'Axis Bank Formats & Passwords',
          rows: [
            { key: 'Credit Card (PDF)', value: 'First 4 letters of name in ALL CAPS/lowercase + last 4 digits of credit card number.' },
            { key: 'Savings Account (CSV/PDF)', value: 'Axis NetBanking e-statement export.' },
          ],
        },
      },
      {
        number: 5,
        title: 'State Bank of India (SBI) & Generic CSV Extractor',
        content:
          'LocalFinance includes a universal CSV sniffer that automatically identifies column delimiters (comma, semicolon, tab, pipe), removes UTF-8 and UTF-16 Byte Order Marks (BOM), and extracts date, narration, reference number, withdrawal, and deposit fields from SBI and any standard bank format.',
        proTip:
          'If your bank is not yet in the pre-configured registry, simply export your statement as a CSV. The generic extractor can parse it automatically.',
      },
    ],
    footerNotes:
      'Want support for Kotak Mahindra, IDFC First, IndusInd, or Amex India? Additional parsers can be plugged directly into the modular parser registry.',
  },
  {
    id: 'importing',
    title: 'How to Import Statements (Step-by-Step)',
    shortTitle: 'Statement Ingestion',
    category: 'Getting Started',
    readTime: '4 min read',
    icon: UploadCloud,
    iconColor: '#10B981',
    summary:
      'A complete step-by-step walkthrough to downloading, decrypting, previewing, and importing bank and card statements into your local ledger.',
    ctaText: 'Open Statement Importer',
    ctaUrl: '/import',
    steps: [
      {
        number: 1,
        title: 'Download Clean Statements from Your Bank Portal',
        content:
          'Log in to your bank’s official NetBanking portal or mobile app and download your statements in PDF, Excel (.xls/.xlsx), or CSV. For the best accuracy, choose digital statements rather than scanned image copies.',
        calloutType: 'tip',
        proTip:
          'You can download overlapping monthly statement cycles without worry. LocalFinance checks every single transaction hash before saving to prevent duplicate entries.',
      },
      {
        number: 2,
        title: 'Drag and Drop into the Ingestion Dropzone',
        content:
          'Navigate to the Import Statements page (/import). Drag and drop one or multiple statement files directly onto the upload zone. You can also click the dropzone to browse your local files.',
      },
      {
        number: 3,
        title: 'Unlock Protected PDF Statements in Memory',
        content:
          'If a PDF statement is password-protected, LocalFinance will prompt you with an inline password field. Enter the statement password (such as your NetBanking customer ID or name + DOB recipe). The file is decrypted purely in memory.',
        calloutType: 'security',
        proTip:
          'Your password is never saved to the SQLite database or written anywhere. Once the file is parsed into structured records, the password is discarded.',
      },
      {
        number: 4,
        title: 'Inspect the Interactive Pre-Commit Audit Table',
        content:
          'Before any transactions are committed to your database, LocalFinance opens an interactive preview table showing: detected bank name, account type, statement date range, opening/closing balance, total debits, total credits, and a live breakdown of new vs already-imported duplicate transactions.',
      },
      {
        number: 5,
        title: 'Commit Transactions to Local SQLite Database',
        content:
          'Review the preview summary and click "Commit to Database". All new transactions are immediately indexed, categorized using your prioritization rules, and linked to payment accounts.',
      },
    ],
    footerNotes:
      'Test without real files: Click "Load Sample Statement" on the Import page to try out sample HDFC, ICICI, and Axis statements pre-loaded with sample Indian transactions.',
  },
  {
    id: 'cashflow',
    title: 'Cash Flow Sankey & MoM Shift Intelligence',
    shortTitle: 'Cash Flow & Sankey',
    category: 'Financial Intelligence',
    readTime: '5 min read',
    icon: Layers,
    iconColor: '#14B8A6',
    summary:
      'Visualize complete money flow routing from income sources through accounts and payment channels to spending categories, paired with automatic surge detection.',
    ctaText: 'View Cash Flow Sankey',
    ctaUrl: '/cashflow',
    steps: [
      {
        number: 1,
        title: 'The 4-Tier Money Flow Topology',
        content:
          'The Sankey diagram maps how capital moves across four distinct layers:\n\n1. Layer 0 (Income Sources): Salary, Refunds, Investments, and Other credits.\n2. Layer 1 (Holding Accounts): HDFC Savings, ICICI Savings, etc.\n3. Layer 2 (Payment Channels): Direct Bank & UPI debits vs Credit Card channels.\n4. Layer 3 (Destinations): Spending categories (Dining, Groceries, Shopping) + Net Savings Surplus.',
      },
      {
        number: 2,
        title: 'Balancing Surpluses and Deficits',
        content:
          'When income exceeds spending, the remaining money routes into a green "Net Savings Surplus" node at Layer 3. If expenditures exceed income for a month, a "Prior Balance Deficit" node appears at Layer 0 to highlight that past reserves were tapped.',
        calloutType: 'info',
      },
      {
        number: 3,
        title: 'Bidirectional Ribbon Highlighting',
        content:
          'Hover your cursor over any node or connecting bezier ribbon. LocalFinance automatically dims unrelated pathways and illuminates all interconnected inflows and outflows, showing exact amounts and percentage share of total flow.',
      },
      {
        number: 4,
        title: 'Automatic MoM Anomaly Sniffer',
        content:
          'LocalFinance compares the active period against previous statements to detect notable patterns:\n\n• Spending Surges: Flags categories with ≥30% increase and >₹1,000 delta, highlighting the top contributor.\n• Significant Drops: Celebrates categories that decreased by >30%.\n• New Outflows: Identifies categories with first-time expenses.\n• High Savings Milestones: Celebrates retaining ≥40% of income as surplus.',
      },
      {
        number: 5,
        title: 'Category Trajectory & 6-Month Mini Sparklines',
        content:
          'Beneath the Sankey diagram, inspect the Category Comparison Table. Compare the current month against the prior month and 3-month rolling averages, and scan the 6-month mini SVG trendlines to spot spending trajectory over time.',
      },
    ],
    footerNotes:
      'Tip: Click any category node in the Sankey diagram to open an instant modal showing all individual transactions that make up that flow.',
  },
  {
    id: 'reconciliation',
    title: 'Smart Transfer Reconciler (Double-Counting Fix)',
    shortTitle: 'Transfer Reconciler',
    category: 'Financial Intelligence',
    readTime: '4 min read',
    icon: ArrowLeftRight,
    iconColor: '#06B6D4',
    summary:
      'Automatically match bank account debits with credit card payments to eliminate double-counted expenses and exclude intermediate wallet loads.',
    ctaText: 'Open Transfer Reconciler',
    ctaUrl: '/reconcile',
    steps: [
      {
        number: 1,
        title: 'The Double-Counting Problem in Personal Finance',
        content:
          'When you swipe a credit card for ₹5,000 at Swiggy, that is your actual expense. Later, when you pay your ₹5,000 card bill from your HDFC savings account via Cred or NetBanking, standard tools count both transactions as expenses, falsely reporting ₹10,000 in spend.',
        calloutType: 'warning',
      },
      {
        number: 2,
        title: 'Cross-Account 4-Day Heuristic Auto-Pairing',
        content:
          'LocalFinance scans your ledger to identify savings account debits (marked as Cred, BBPS, IMPS, or NetBanking) and pairs them with credit card payment credits occurring within a 4-day window with matched amounts. The candidate pairs are assigned a match confidence score (e.g. 95% Match).',
      },
      {
        number: 3,
        title: '1-Click Confirm & Link Transfer',
        content:
          'Review candidate matches in the Transfer Reconciler workbench (/reconcile). Click "Confirm & Link Transfer" to link the transactions. Once linked, the bill payment transfer is excluded from monthly spend while your itemized Swiggy purchases remain intact.',
      },
      {
        number: 4,
        title: 'UPI Lite & Wallet Top-Up Exclusion',
        content:
          'When you load ₹2,000 into Paytm Wallet or UPI Lite, that is a transfer between your own balances. LocalFinance sniffs wallet top-up narrations and automatically excludes them from spending totals, with simple 1-click override controls.',
      },
    ],
  },
  {
    id: 'cards',
    title: 'Credit Card Rewards & Grace Period Optimizer',
    shortTitle: 'Cards & Rewards',
    category: 'Financial Intelligence',
    readTime: '4 min read',
    icon: CreditCard,
    iconColor: '#EC4899',
    summary:
      'Recommend the highest cashback card for any merchant swipe, maximize 50-day interest-free runways, and track annual fee waiver milestones.',
    ctaText: 'Optimize Card Rewards',
    ctaUrl: '/cards',
    steps: [
      {
        number: 1,
        title: 'Interactive Best-Card Recommender',
        content:
          'Before making a purchase, check the Best Card tool (/cards). Search or click quick merchant tags like Swiggy, Blinkit, Amazon, Flipkart, Fuel, or Flights. LocalFinance compares your card portfolio reward rules to tell you which card yields the highest cashback percentage and rupee savings.',
        sampleData: {
          label: 'Card Reward Examples',
          rows: [
            { key: 'Swiggy HDFC Card', value: '10% cashback on Swiggy food delivery, Dineout, Instamart.' },
            { key: 'Amazon Pay ICICI Card', value: '5% unlimited cashback on Amazon Shopping for Prime members.' },
            { key: 'Axis Flipkart Card', value: '5% cashback on Flipkart, Myntra, and Cleartrip.' },
          ],
        },
      },
      {
        number: 2,
        title: '50-Day Interest-Free Runway Optimizer',
        content:
          'Credit cards provide up to 50 days of interest-free credit depending on the swipe date relative to your statement generation date. LocalFinance calculates the exact days remaining until payment is due across all your cards, highlighting the best card to swipe today for maximum cash float.',
      },
      {
        number: 3,
        title: 'Annual Fee Waiver Spend Tracker',
        content:
          'Many premium Indian credit cards waive annual renewal fees if you cross a yearly spend milestone (e.g. ₹2,00,000/year). LocalFinance tracks your cumulative milestone spend progress and projects whether you will reach the waiver threshold before renewal.',
      },
    ],
  },
  {
    id: 'budgets',
    title: 'Category Budgets & Overspend Pacing',
    shortTitle: 'Budgets & Pacing',
    category: 'Financial Intelligence',
    readTime: '3 min read',
    icon: Target,
    iconColor: '#F59E0B',
    summary:
      'Set monthly category limits, monitor visual pacing velocity bars, and calculate safe daily spending allowances.',
    ctaText: 'Manage Budgets',
    ctaUrl: '/budget',
    steps: [
      {
        number: 1,
        title: 'Setting Category Spend Limits',
        content:
          'Open Category Budgets (/budget). You will see all your active categories (Dining, Groceries, Shopping, Fuel, Entertainment). Click "Set Budget" to assign a monthly rupee target with quick preset limit chips (₹5,000, ₹10,000, ₹20,000).',
      },
      {
        number: 2,
        title: 'Reading Dynamic Pacing Progress Bars',
        content:
          'Unlike basic budget bars that only show spend percentage, LocalFinance calculates your pacing velocity against the current day of the month:\n\n• Green: Spending is well below pace for the current calendar day.\n• Amber: Spending has reached 80% or is pacing faster than the calendar.\n• Red: Category has exceeded 100% of the allocated monthly limit.',
      },
      {
        number: 3,
        title: 'Safe Daily Recommended Allowance',
        content:
          'LocalFinance computes your remaining safe daily burn rate for the rest of the month. For example, if you have ₹6,800 left in your Dining budget with 20 days remaining, your safe allowance is ₹340 / day. This gives you an actionable daily benchmark.',
      },
    ],
  },
  {
    id: 'transactions',
    title: 'Transactions Ledger & Indian UPI Cleaner',
    shortTitle: 'Transactions Ledger',
    category: 'Daily Tracking',
    readTime: '4 min read',
    icon: ReceiptText,
    iconColor: '#6366F1',
    summary:
      'High-performance ledger with specialized Indian banking narration cleaner, UPI VPA extraction, time window filters, and 1-click rule creation.',
    ctaText: 'Open Ledger',
    ctaUrl: '/transactions',
    steps: [
      {
        number: 1,
        title: 'How the Indian Narration Cleaner Works',
        content:
          'Indian bank statement lines are notoriously cryptic (e.g., "UPI/DR/423589214781/SWIGGY BANGA/PYTM/UPI/"). LocalFinance runs an offline regex engine that extracts the clean payee ("Swiggy"), strips bank routing prefixes, isolates the 12-digit UTR reference number, and preserves the UPI VPA handle.',
      },
      {
        number: 2,
        title: 'Quick Time Window Preset Filters',
        content:
          'Switch instantly between time horizons using the top preset pills: All Time, Today, Yesterday, This Week, This Month, This Year, or Custom Date Range with HTML5 date pickers.',
      },
      {
        number: 3,
        title: 'Transaction Details Inspector Modal',
        content:
          'Click any transaction row to open the inspector dialog. View the raw unedited bank narration, cleaned payee, payment mode (UPI, POS, IMPS, NEFT, CC), transaction date and value date, card last 4 digits, and associated bank account.',
      },
      {
        number: 4,
        title: '1-Click "Create Auto-Rule from Payee"',
        content:
          'Need to categorize a recurring merchant automatically? Tap "Create Auto-Rule from Payee" inside the transaction modal. It pre-populates a prioritized match rule with the merchant pattern, ready to save with one click.',
      },
    ],
  },
  {
    id: 'subscriptions',
    title: 'Subscriptions & Recurring Bills Tracker',
    shortTitle: 'Subscriptions & Mandates',
    category: 'Daily Tracking',
    readTime: '3 min read',
    icon: Sparkles,
    iconColor: '#8B5CF6',
    summary:
      'Multi-tier auto-detection for UPI AutoPay mandates, OTT media subscriptions, broadband, telecom bills, and loan EMIs.',
    ctaText: 'Track Subscriptions',
    ctaUrl: '/subscriptions',
    steps: [
      {
        number: 1,
        title: 'Automatic Cadence & Mandate Sniffer',
        content:
          'Click "Scan Ledger" on the Subscriptions page (/subscriptions). LocalFinance searches for recurring charges across three tiers:\n\n1. UPI AutoPay & e-Mandates (captures AUTOPAY, MANDATEEXECUTE, NACH).\n2. Known Catalog (Netflix, Spotify, Amazon Prime, YouTube, ChatGPT, Apple, Airtel, Jio).\n3. Statistical Interval Sniffer (detects uncatalogued recurring merchants with 28-to-33 day intervals).',
      },
      {
        number: 2,
        title: 'Monthly Burn Rate & Annual Forecast',
        content:
          'The top KPI ribbon summarizes your monthly recurring burn rate (₹/mo) and projects your annual financial commitment (₹/yr), helping you audit unnecessary recurring charges.',
      },
      {
        number: 3,
        title: 'Upcoming Renewal Schedule with Countdowns',
        content:
          'Review the chronological timeline of upcoming due dates with badges indicating "Due today", "In 3 days", or "In 14 days" so you are never caught off guard by an auto-debit.',
      },
    ],
  },
  {
    id: 'merchants',
    title: 'Merchant Intelligence Deep-Dive',
    shortTitle: 'Merchant Profiles',
    category: 'Daily Tracking',
    readTime: '3 min read',
    icon: Store,
    iconColor: '#06B6D4',
    summary:
      'Lifetime spending profiles, order volume metrics, average order values (AOV), and payment method breakdowns for every merchant.',
    ctaText: 'Inspect Merchants',
    ctaUrl: '/merchants',
    steps: [
      {
        number: 1,
        title: 'Merchant Aggregation Directory',
        content:
          'The Merchant Intelligence page (/merchants) aggregates all transactions across your accounts by cleaned payee name. Sort by Highest Spend, Most Orders, Highest Average Order Value (AOV), or Most Recent Activity.',
      },
      {
        number: 2,
        title: 'Lifetime Merchant Profile Modal',
        content:
          'Click any merchant card (e.g. Swiggy, Amazon, Blinkit, Uber, Shell) to open the deep-dive profile. Inspect lifetime metrics: total amount spent, net order count, mean order value, and first/last transaction dates.',
      },
      {
        number: 3,
        title: 'Payment Method Breakdown',
        content:
          'See exactly which cards or accounts you use most frequently at that merchant (e.g., 78% Swiggy HDFC Card, 22% Savings UPI). This pairs directly with the Card Reward Optimizer to ensure you are maximizing cashbacks.',
      },
    ],
  },
  {
    id: 'rules-and-data',
    title: 'Auto-Rules, Database Backups & 100% Offline Privacy',
    shortTitle: 'Rules & Database',
    category: 'System & Storage',
    readTime: '4 min read',
    icon: Settings,
    iconColor: '#64748B',
    summary:
      'Manage custom categorization rules, download standalone SQLite snapshots with WAL checkpoint, export CSV/JSON dumps, and maintain 100% offline privacy.',
    ctaText: 'Open Settings',
    ctaUrl: '/settings',
    steps: [
      {
        number: 1,
        title: 'Custom Categories & Categorization Rules',
        content:
          'Define your own spending taxonomy with custom category names and theme colors. Create prioritized match rules using contains or regex patterns against cleaned payee or raw narration.',
        proTip:
          'Rules have priority weights (1 to 100). Higher priority rules match first, allowing specific merchants to override broader general categories.',
      },
      {
        number: 2,
        title: '1-Click Re-Application to Historical Records',
        content:
          'When you create or update a categorization rule, tap "Re-Apply Rules to Transactions". LocalFinance immediately re-evaluates all historical ledger records in your database while strictly preserving any manual category edits you have made.',
      },
      {
        number: 3,
        title: '1-Click Standalone SQLite Backup Download',
        content:
          'In Settings > Data & Storage, click "Download SQLite DB (.db)". LocalFinance executes PRAGMA wal_checkpoint to flush memory buffers and downloads a standalone SQLite database snapshot directly to your device.',
        calloutType: 'security',
      },
      {
        number: 4,
        title: 'Clean CSV and Full JSON Data Exports',
        content:
          'Need to run calculations in Excel or store structured archives? Download a clean CSV ledger export or a full JSON database dump with all accounts, bills, statements, and rules.',
      },
    ],
    footerNotes:
      'Zero Cloud Guarantee: LocalFinance will never send telemetry, ping an external API, or connect to remote servers. All your financial data lives exclusively on your local disk.',
  },
]
