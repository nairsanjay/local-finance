import React, { useState, useMemo } from 'react'
import { SankeyData, SankeyNode } from '@/types'
import { formatINR } from '@/lib/utils'
import {
  Briefcase,
  RotateCcw,
  TrendingUp,
  ArrowDownLeft,
  Landmark,
  CreditCard,
  PiggyBank,
  AlertTriangle,
  HelpCircle,
  Utensils,
  ShoppingBag,
  ShoppingCart,
  Receipt,
  Car,
  Film,
  HeartPulse,
  Percent,
  Wallet,
} from 'lucide-react'

interface SankeyDiagramProps {
  data: SankeyData
  onNodeClick?: (node: SankeyNode) => void
  selectedNodeId?: string | null
}

const getIconComponent = (iconName?: string, type?: string) => {
  switch (iconName) {
    case 'Briefcase': return Briefcase
    case 'RotateCcw': return RotateCcw
    case 'TrendingUp': return TrendingUp
    case 'ArrowDownLeft': return ArrowDownLeft
    case 'Landmark': return Landmark
    case 'CreditCard': return CreditCard
    case 'PiggyBank': return PiggyBank
    case 'AlertTriangle': return AlertTriangle
    case 'Utensils': return Utensils
    case 'ShoppingBag': return ShoppingBag
    case 'ShoppingCart': return ShoppingCart
    case 'Receipt': return Receipt
    case 'Car': return Car
    case 'Film': return Film
    case 'HeartPulse': return HeartPulse
    case 'Percent': return Percent
    case 'Wallet': return Wallet
    default:
      if (type === 'INCOME_SOURCE') return Briefcase
      if (type === 'ACCOUNT') return Landmark
      if (type === 'CHANNEL') return CreditCard
      if (type === 'SURPLUS') return PiggyBank
      if (type === 'DEFICIT') return AlertTriangle
      return HelpCircle
  }
}

interface LayoutNode extends SankeyNode {
  x: number
  y: number
  width: number
  height: number
  layer: number
}

interface LayoutLink {
  source: LayoutNode
  target: LayoutNode
  value: number
  sourceY: number
  sourceHeight: number
  targetY: number
  targetHeight: number
  color: string
}

export const SankeyDiagram: React.FC<SankeyDiagramProps> = ({
  data,
  onNodeClick,
  selectedNodeId,
}) => {
  const [hoveredNodeId, setHoveredNodeId] = useState<string | null>(null)
  const [hoveredLink, setHoveredLink] = useState<{
    sourceName: string
    targetName: string
    value: number
    pct: number
    x: number
    y: number
  } | null>(null)

  const { layoutNodes, layoutLinks, layers, width, height } = useMemo(() => {
    const defaultWidth = 1040
    const nodeWidth = 175
    const minNodeHeight = 38
    const nodeGap = 12
    const paddingTop = 28
    const paddingBottom = 20

    if (!data || !data.nodes || data.nodes.length === 0) {
      return { layoutNodes: [], layoutLinks: [], layers: [], width: defaultWidth, height: 420 }
    }

    // Group nodes by layer
    const rawLayers: SankeyNode[][] = [[], [], [], []]
    const nodeMap = new Map<string, SankeyNode>()

    data.nodes.forEach((n) => {
      nodeMap.set(n.id, n)
      const layerIdx = Math.min(Math.max(n.layer, 0), 3)
      rawLayers[layerIdx].push(n)
    })

    // Filter out empty intermediate layers and calculate X positions
    const activeLayerIndices = rawLayers
      .map((arr, idx) => (arr.length > 0 ? idx : -1))
      .filter((idx) => idx !== -1)

    const totalActiveLayers = activeLayerIndices.length
    const colSpacing =
      totalActiveLayers > 1
        ? (defaultWidth - nodeWidth * totalActiveLayers) / (totalActiveLayers - 1)
        : 0

    // Find max layer total for proportional height scaling
    const maxLayerSum = Math.max(
      ...rawLayers.map((nodes) => nodes.reduce((sum, n) => sum + n.total_value, 0)),
      1
    )

    // Estimate total canvas height needed
    const maxNodesInAnyLayer = Math.max(...rawLayers.map((l) => l.length), 1)
    const dynamicHeight = Math.max(
      460,
      maxNodesInAnyLayer * (minNodeHeight + nodeGap) + paddingTop + paddingBottom + 40
    )
    const availableHeight = dynamicHeight - paddingTop - paddingBottom

    const computedNodes: LayoutNode[] = []
    const layoutNodeMap = new Map<string, LayoutNode>()

    activeLayerIndices.forEach((layerIdx, colIdx) => {
      const nodesInLayer = rawLayers[layerIdx]
      const layerTotal = nodesInLayer.reduce((s, n) => s + n.total_value, 0)
      const x = colIdx * (nodeWidth + colSpacing)

      // Calculate total gaps
      const totalGaps = (nodesInLayer.length - 1) * nodeGap
      const spaceForNodes = Math.max(availableHeight - totalGaps, nodesInLayer.length * minNodeHeight)

      // Calculate heights
      let currentY = paddingTop
      nodesInLayer.forEach((n) => {
        const proportionalHeight =
          layerTotal > 0
            ? Math.round((n.total_value / maxLayerSum) * spaceForNodes)
            : minNodeHeight
        const nodeH = Math.max(minNodeHeight, proportionalHeight)

        const lNode: LayoutNode = {
          ...n,
          x,
          y: currentY,
          width: nodeWidth,
          height: nodeH,
          layer: layerIdx,
        }
        computedNodes.push(lNode)
        layoutNodeMap.set(n.id, lNode)
        currentY += nodeH + nodeGap
      })
    })

    // Track offsets for incoming and outgoing links per node
    const sourceOffsets = new Map<string, number>()
    const targetOffsets = new Map<string, number>()

    computedNodes.forEach((n) => {
      sourceOffsets.set(n.id, 0)
      targetOffsets.set(n.id, 0)
    })

    const computedLinks: LayoutLink[] = []

    data.links.forEach((link) => {
      const src = layoutNodeMap.get(link.source)
      const tgt = layoutNodeMap.get(link.target)
      if (!src || !tgt || link.value <= 0) return

      const srcCurrentOffset = sourceOffsets.get(src.id) || 0
      const tgtCurrentOffset = targetOffsets.get(tgt.id) || 0

      const srcRatio = src.total_value > 0 ? link.value / src.total_value : 1
      const tgtRatio = tgt.total_value > 0 ? link.value / tgt.total_value : 1

      const srcLinkHeight = Math.max(2, src.height * srcRatio)
      const tgtLinkHeight = Math.max(2, tgt.height * tgtRatio)

      const linkObj: LayoutLink = {
        source: src,
        target: tgt,
        value: link.value,
        sourceY: src.y + srcCurrentOffset,
        sourceHeight: srcLinkHeight,
        targetY: tgt.y + tgtCurrentOffset,
        targetHeight: tgtLinkHeight,
        color: link.color_hex || src.color_hex,
      }

      computedLinks.push(linkObj)

      sourceOffsets.set(src.id, srcCurrentOffset + srcLinkHeight)
      targetOffsets.set(tgt.id, tgtCurrentOffset + tgtLinkHeight)
    })

    const layerHeaders = [
      { label: 'Income Sources', layer: 0 },
      { label: 'Holding Accounts', layer: 1 },
      { label: 'Payment Channels', layer: 2 },
      { label: 'Categories & Surplus', layer: 3 },
    ].filter((h) => activeLayerIndices.includes(h.layer))

    return {
      layoutNodes: computedNodes,
      layoutLinks: computedLinks,
      layers: layerHeaders,
      width: defaultWidth,
      height: dynamicHeight,
    }
  }, [data])

  // Helper to determine if a link or node is highlighted
  const isNodeConnected = (nodeId: string) => {
    if (!hoveredNodeId) return false
    if (nodeId === hoveredNodeId) return true
    return layoutLinks.some(
      (l) =>
        (l.source.id === hoveredNodeId && l.target.id === nodeId) ||
        (l.target.id === hoveredNodeId && l.source.id === nodeId)
    )
  }

  const isLinkHighlighted = (link: LayoutLink) => {
    if (!hoveredNodeId) return false
    return link.source.id === hoveredNodeId || link.target.id === hoveredNodeId
  }

  const generatePath = (link: LayoutLink) => {
    const x0 = link.source.x + link.source.width
    const y0 = link.sourceY
    const h0 = link.sourceHeight
    const x1 = link.target.x
    const y1 = link.targetY
    const h1 = link.targetHeight

    const dx = x1 - x0
    const cpx1 = x0 + dx * 0.45
    const cpx2 = x1 - dx * 0.45

    return `
      M ${x0} ${y0}
      C ${cpx1} ${y0}, ${cpx2} ${y1}, ${x1} ${y1}
      L ${x1} ${y1 + h1}
      C ${cpx2} ${y1 + h1}, ${cpx1} ${y0 + h0}, ${x0} ${y0 + h0}
      Z
    `
  }

  if (layoutNodes.length === 0) {
    return (
      <div className="flex h-64 flex-col items-center justify-center rounded-xl border border-dashed p-8 text-center text-muted-foreground">
        <PiggyBank className="h-8 w-8 text-muted-foreground/50 mb-2" />
        <p className="text-sm font-medium">No cash flow transaction data in this period</p>
        <p className="text-xs text-muted-foreground mt-0.5">
          Upload bank statements or select a different month to generate the Sankey diagram
        </p>
      </div>
    )
  }

  return (
    <div className="relative w-full overflow-x-auto select-none">
      {/* Floating Tooltip */}
      {hoveredLink && (
        <div
          className="pointer-events-none absolute z-50 rounded-lg border bg-popover/95 backdrop-blur-md px-3 py-2 text-xs shadow-xl text-popover-foreground transition-all duration-75"
          style={{
            left: `${Math.min(Math.max(hoveredLink.x - 70, 10), width - 180)}px`,
            top: `${hoveredLink.y - 45}px`,
          }}
        >
          <div className="flex items-center gap-1.5 font-semibold">
            <span>{hoveredLink.sourceName}</span>
            <span className="text-muted-foreground">→</span>
            <span>{hoveredLink.targetName}</span>
          </div>
          <div className="flex items-center gap-2 mt-1 text-muted-foreground">
            <span className="font-bold text-foreground">{formatINR(hoveredLink.value)}</span>
            <span>•</span>
            <span>{hoveredLink.pct.toFixed(1)}% of flow</span>
          </div>
        </div>
      )}

      {/* Layer Labels Header */}
      <div className="flex justify-between px-2 mb-2 min-w-[1040px]">
        {layers.map((l) => (
          <div
            key={l.layer}
            className="w-[175px] text-[11px] font-bold uppercase tracking-wider text-muted-foreground/75 text-center"
          >
            {l.label}
          </div>
        ))}
      </div>

      <svg
        viewBox={`0 0 ${width} ${height}`}
        className="w-full h-auto min-w-[1040px] overflow-visible"
        style={{ height }}
      >
        <defs>
          {layoutLinks.map((link, idx) => {
            const gradId = `grad-${link.source.id}-${link.target.id}-${idx}`
            return (
              <linearGradient
                key={gradId}
                id={gradId}
                gradientUnits="userSpaceOnUse"
                x1={link.source.x + link.source.width}
                y1="0"
                x2={link.target.x}
                y2="0"
              >
                <stop offset="0%" stopColor={link.source.color_hex} stopOpacity="0.45" />
                <stop offset="100%" stopColor={link.target.color_hex} stopOpacity="0.45" />
              </linearGradient>
            )
          })}
        </defs>

        {/* 1. Sankey Links / Flow Ribbons */}
        <g className="links-layer">
          {layoutLinks.map((link, idx) => {
            const gradId = `grad-${link.source.id}-${link.target.id}-${idx}`
            const highlighted = isLinkHighlighted(link)
            const dimmed = hoveredNodeId && !highlighted

            return (
              <path
                key={`link-${idx}`}
                d={generatePath(link)}
                fill={`url(#${gradId})`}
                className="transition-all duration-200 cursor-pointer"
                style={{
                  opacity: highlighted ? 0.9 : dimmed ? 0.12 : 0.65,
                  stroke: highlighted ? link.color : 'none',
                  strokeWidth: highlighted ? 1 : 0,
                }}
                onMouseEnter={(e) => {
                  const rect = e.currentTarget.getBoundingClientRect()
                  const parentRect = e.currentTarget.closest('svg')?.getBoundingClientRect()
                  const relX = (rect.left + rect.right) / 2 - (parentRect?.left || 0)
                  const relY = (rect.top + rect.bottom) / 2 - (parentRect?.top || 0)

                  const totalRef = link.source.total_value || 1
                  setHoveredLink({
                    sourceName: link.source.name,
                    targetName: link.target.name,
                    value: link.value,
                    pct: (link.value / totalRef) * 100,
                    x: relX,
                    y: relY,
                  })
                }}
                onMouseLeave={() => setHoveredLink(null)}
              />
            )
          })}
        </g>

        {/* 2. Sankey Nodes (HTML-in-SVG foreignObjects for pixel-perfect card styling) */}
        <g className="nodes-layer">
          {layoutNodes.map((node) => {
            const Icon = getIconComponent(node.icon, node.type)
            const isHovered = hoveredNodeId === node.id
            const isConnected = isNodeConnected(node.id)
            const isSelected = selectedNodeId === node.id
            const isDimmed = hoveredNodeId && !isHovered && !isConnected

            const isSurplus = node.type === 'SURPLUS'
            const isDeficit = node.type === 'DEFICIT'

            return (
              <foreignObject
                key={node.id}
                x={node.x}
                y={node.y}
                width={node.width}
                height={node.height}
                className="overflow-visible cursor-pointer"
                onMouseEnter={() => setHoveredNodeId(node.id)}
                onMouseLeave={() => setHoveredNodeId(null)}
                onClick={() => onNodeClick && onNodeClick(node)}
              >
                <div
                  className={`group relative flex h-full w-full flex-col justify-between rounded-xl border p-2.5 transition-all duration-150 ${
                    isSelected
                      ? 'ring-2 ring-primary ring-offset-2 ring-offset-background bg-card shadow-lg'
                      : isHovered
                      ? 'scale-[1.02] shadow-md border-primary/60 bg-card/95'
                      : isDimmed
                      ? 'opacity-35 bg-card/40 border-border/40'
                      : isSurplus
                      ? 'bg-emerald-500/10 border-emerald-500/30 dark:bg-emerald-950/20'
                      : isDeficit
                      ? 'bg-amber-500/10 border-amber-500/30 dark:bg-amber-950/20'
                      : 'bg-card/90 hover:bg-card border-border shadow-xs'
                  }`}
                  style={{
                    borderLeftColor: node.color_hex,
                    borderLeftWidth: '4px',
                  }}
                >
                  {/* Top: Icon + Name */}
                  <div className="flex items-center gap-2 min-w-0">
                    <div
                      className="flex h-5 w-5 shrink-0 items-center justify-center rounded-md text-white text-[10px] shadow-xs"
                      style={{ backgroundColor: node.color_hex }}
                    >
                      <Icon className="h-3 w-3" />
                    </div>
                    <span className="text-[11px] font-semibold text-foreground truncate tracking-tight">
                      {node.name}
                    </span>
                  </div>

                  {/* Bottom: Amount & Percentage */}
                  <div className="flex items-baseline justify-between gap-1 mt-0.5">
                    <span className="text-xs font-bold text-foreground">
                      {formatINR(node.total_value)}
                    </span>
                    {node.tx_count && node.tx_count > 0 && (
                      <span className="text-[9px] font-medium text-muted-foreground">
                        {node.tx_count} {node.tx_count === 1 ? 'tx' : 'txs'}
                      </span>
                    )}
                  </div>
                </div>
              </foreignObject>
            )
          })}
        </g>
      </svg>
    </div>
  )
}
