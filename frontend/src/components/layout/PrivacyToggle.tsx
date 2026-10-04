import React from 'react'
import { usePrivacy } from '@/components/privacy-provider'
import { Eye, EyeOff } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'

export const PrivacyToggle: React.FC = () => {
  const { isPrivacyMode, togglePrivacyMode } = usePrivacy()

  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <Button
            variant={isPrivacyMode ? 'secondary' : 'ghost'}
            size="sm"
            onClick={togglePrivacyMode}
            className={`h-8 px-2 sm:px-2.5 gap-1.5 text-xs transition-all ${
              isPrivacyMode
                ? 'bg-amber-500/15 text-amber-600 dark:text-amber-400 border border-amber-500/30 hover:bg-amber-500/25'
                : 'text-muted-foreground hover:text-foreground'
            }`}
            aria-label={isPrivacyMode ? 'Disable Discreet Mode' : 'Enable Discreet Mode'}
          >
            {isPrivacyMode ? (
              <EyeOff className="h-3.5 w-3.5 text-amber-500 shrink-0" />
            ) : (
              <Eye className="h-3.5 w-3.5 shrink-0" />
            )}
            <span className="hidden md:inline font-medium">
              {isPrivacyMode ? 'Discreet' : 'Public'}
            </span>
            <kbd className="hidden lg:inline-flex h-4 select-none items-center gap-0.5 rounded border border-border/80 bg-muted/60 px-1 font-mono text-[9px] font-medium text-muted-foreground">
              P
            </kbd>
          </Button>
        }
      />
      <TooltipContent side="bottom" align="end" className="text-xs">
        <p className="font-semibold">
          {isPrivacyMode ? 'Discreet Mode Active' : 'Enable Discreet Mode'}
        </p>
        <p className="text-[11px] text-muted-foreground">
          {isPrivacyMode
            ? 'Balances are blurred/masked. Press P or ⌘. to restore.'
            : 'Mask balances & numbers in public spaces (Hotkey: P)'}
        </p>
      </TooltipContent>
    </Tooltip>
  )
}
