import React from 'react'

interface MerchantBrand {
  name: string
  icon: string
  bg: string
  textColor: string
  borderColor: string
}

export function getMerchantBrand(rawPayee: string): MerchantBrand {
  const p = (rawPayee || '').toUpperCase()

  if (p.includes('SWIGGY')) {
    return { name: 'Swiggy', icon: '🍔', bg: 'bg-orange-500/15', textColor: 'text-orange-500', borderColor: 'border-orange-500/30' }
  }
  if (p.includes('ZOMATO')) {
    return { name: 'Zomato', icon: '🍽️', bg: 'bg-red-500/15', textColor: 'text-red-500', borderColor: 'border-red-500/30' }
  }
  if (p.includes('BLINKIT') || p.includes('GROFERS')) {
    return { name: 'Blinkit', icon: '🛒', bg: 'bg-amber-500/15', textColor: 'text-amber-500', borderColor: 'border-amber-500/30' }
  }
  if (p.includes('ZEPTO')) {
    return { name: 'Zepto', icon: '⚡', bg: 'bg-purple-500/15', textColor: 'text-purple-400', borderColor: 'border-purple-500/30' }
  }
  if (p.includes('AMAZON')) {
    return { name: 'Amazon', icon: '📦', bg: 'bg-yellow-500/15', textColor: 'text-yellow-600 dark:text-yellow-400', borderColor: 'border-yellow-500/30' }
  }
  if (p.includes('FLIPKART')) {
    return { name: 'Flipkart', icon: '🛍️', bg: 'bg-blue-500/15', textColor: 'text-blue-400', borderColor: 'border-blue-500/30' }
  }
  if (p.includes('MYNTRA')) {
    return { name: 'Myntra', icon: '👗', bg: 'bg-pink-500/15', textColor: 'text-pink-400', borderColor: 'border-pink-500/30' }
  }
  if (p.includes('UBER')) {
    return { name: 'Uber', icon: '🚗', bg: 'bg-slate-500/15', textColor: 'text-slate-300', borderColor: 'border-slate-500/30' }
  }
  if (p.includes('OLA')) {
    return { name: 'Ola', icon: '🚕', bg: 'bg-lime-500/15', textColor: 'text-lime-400', borderColor: 'border-lime-500/30' }
  }
  if (p.includes('NETFLIX')) {
    return { name: 'Netflix', icon: '🎬', bg: 'bg-rose-500/15', textColor: 'text-rose-500', borderColor: 'border-rose-500/30' }
  }
  if (p.includes('SPOTIFY')) {
    return { name: 'Spotify', icon: '🎵', bg: 'bg-emerald-500/15', textColor: 'text-emerald-400', borderColor: 'border-emerald-500/30' }
  }
  if (p.includes('SHELL') || p.includes('PETROL') || p.includes('HPCL') || p.includes('BPCL') || p.includes('IOCL') || p.includes('FUEL')) {
    return { name: 'Fuel', icon: '⛽', bg: 'bg-amber-500/15', textColor: 'text-amber-400', borderColor: 'border-amber-500/30' }
  }
  if (p.includes('MAKEMYTRIP') || p.includes('INDIGO') || p.includes('AIR INDIA') || p.includes('FLIGHT') || p.includes('HOTEL')) {
    return { name: 'Travel', icon: '✈️', bg: 'bg-sky-500/15', textColor: 'text-sky-400', borderColor: 'border-sky-500/30' }
  }
  if (p.includes('CRED') || p.includes('BILL PAYMENT')) {
    return { name: 'Card Bill', icon: '💳', bg: 'bg-indigo-500/15', textColor: 'text-indigo-400', borderColor: 'border-indigo-500/30' }
  }
  if (p.includes('ELECTRICITY') || p.includes('BESCOM') || p.includes('TATA POWER') || p.includes('ADANI') || p.includes('UTILITY') || p.includes('WATER') || p.includes('GAS')) {
    return { name: 'Utility', icon: '⚡', bg: 'bg-yellow-500/15', textColor: 'text-yellow-500', borderColor: 'border-yellow-500/30' }
  }
  if (p.includes('AIRTEL') || p.includes('JIO') || p.includes('VI ') || p.includes('VODAFONE')) {
    return { name: 'Telecom', icon: '📱', bg: 'bg-red-500/15', textColor: 'text-red-400', borderColor: 'border-red-500/30' }
  }
  if (p.includes('SALARY') || p.includes('PAYROLL') || p.includes('INTEREST')) {
    return { name: 'Income', icon: '💰', bg: 'bg-emerald-500/15', textColor: 'text-emerald-400', borderColor: 'border-emerald-500/30' }
  }
  if (p.includes('UPI')) {
    return { name: 'UPI', icon: '📱', bg: 'bg-teal-500/15', textColor: 'text-teal-400', borderColor: 'border-teal-500/30' }
  }

  // Fallback: Use first letter initials
  const initials = rawPayee ? rawPayee.slice(0, 2).toUpperCase() : 'TX'
  return {
    name: rawPayee,
    icon: initials,
    bg: 'bg-muted/80',
    textColor: 'text-foreground font-mono text-[10px] font-bold',
    borderColor: 'border-border/80',
  }
}

interface MerchantAvatarProps {
  payee: string
  size?: 'xs' | 'sm' | 'md'
  className?: string
}

export const MerchantAvatar: React.FC<MerchantAvatarProps> = ({
  payee,
  size = 'sm',
  className = '',
}) => {
  const brand = getMerchantBrand(payee)

  const sizeClasses = {
    xs: 'h-5 w-5 text-[11px] rounded-md',
    sm: 'h-6 w-6 text-xs rounded-lg',
    md: 'h-8 w-8 text-sm rounded-xl',
  }[size]

  return (
    <div
      className={`inline-flex shrink-0 items-center justify-center border font-sans select-none shadow-2xs ${brand.bg} ${brand.textColor} ${brand.borderColor} ${sizeClasses} ${className}`}
      title={payee}
    >
      <span>{brand.icon}</span>
    </div>
  )
}
