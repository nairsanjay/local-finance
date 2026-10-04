import React, { useState } from 'react'
import { Check, Copy } from 'lucide-react'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'

interface CopyButtonProps {
  value: string
  label?: string
  className?: string
  size?: 'xs' | 'sm' | 'default'
}

export const CopyButton: React.FC<CopyButtonProps> = ({
  value,
  label = 'Copy to clipboard',
  className = '',
  size = 'xs',
}) => {
  const [copied, setCopied] = useState(false)

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation()
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      // Fallback
      setCopied(false)
    }
  }

  const iconSize = size === 'xs' ? 'h-3 w-3' : 'h-3.5 w-3.5'

  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <button
            type="button"
            onClick={handleCopy}
            className={`inline-flex items-center justify-center h-5 w-5 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-all duration-150 active:scale-90 cursor-pointer ${
              copied ? 'text-emerald-500 hover:text-emerald-500 bg-emerald-500/10' : ''
            } ${className}`}
          >
            {copied ? (
              <Check className={`${iconSize} text-emerald-500 animate-in zoom-in-50 duration-150`} />
            ) : (
              <Copy className={`${iconSize} transition-transform`} />
            )}
          </button>
        }
      />
      <TooltipContent side="top" className="text-[10px] py-1 px-2">
        {copied ? 'Copied!' : label}
      </TooltipContent>
    </Tooltip>
  )
}
