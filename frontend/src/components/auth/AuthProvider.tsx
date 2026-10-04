import React, { createContext, useContext, useEffect, useState, useCallback, useRef } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { fetchAuthStatus, logoutAuth, clearAuthToken } from '@/lib/api'
import { AuthStatusResponse } from '@/types'
import { playPrivacyToggleSound } from '@/lib/audio'

interface AuthContextType {
  authStatus: AuthStatusResponse | undefined
  isLoading: boolean
  isLocked: boolean
  lockApp: () => Promise<void>
  unlock: () => void
  refetchAuth: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const queryClient = useQueryClient()
  const [manualLocked, setManualLocked] = useState(false)
  const lastActivityRef = useRef<number>(Date.now())

  const { data: authStatus, isLoading, refetch } = useQuery({
    queryKey: ['auth-status'],
    queryFn: fetchAuthStatus,
    staleTime: 1000 * 30, // 30 seconds
  })

  // Calculate if app is currently locked
  const isLocked = Boolean(
    authStatus?.auth_enabled && (!authStatus.is_authenticated || manualLocked)
  )

  const lockApp = useCallback(async () => {
    try {
      playPrivacyToggleSound(true)
      await logoutAuth()
    } catch {
      clearAuthToken()
    } finally {
      setManualLocked(true)
      queryClient.setQueryData(['auth-status'], (prev: AuthStatusResponse | undefined) =>
        prev ? { ...prev, is_authenticated: false } : prev
      )
      queryClient.invalidateQueries()
    }
  }, [queryClient])

  const unlock = useCallback(() => {
    setManualLocked(false)
    lastActivityRef.current = Date.now()
    playPrivacyToggleSound(false)
    refetch()
    queryClient.invalidateQueries()
  }, [queryClient, refetch])

  // Listen for 401 AUTH_REQUIRED events from apiFetch
  useEffect(() => {
    const handleAuthRequired = () => {
      setManualLocked(true)
      queryClient.setQueryData(['auth-status'], (prev: AuthStatusResponse | undefined) =>
        prev ? { ...prev, is_authenticated: false } : prev
      )
    }

    window.addEventListener('auth:required', handleAuthRequired)
    return () => window.removeEventListener('auth:required', handleAuthRequired)
  }, [queryClient])

  // Auto-lock on inactivity if enabled
  useEffect(() => {
    if (!authStatus?.auth_enabled || isLocked) return

    const timeoutMinutes = authStatus.auto_lock_minutes || 60
    if (timeoutMinutes <= 0) return

    const timeoutMs = timeoutMinutes * 60 * 1000

    const resetTimer = () => {
      lastActivityRef.current = Date.now()
    }

    const activityEvents = ['mousedown', 'keydown', 'scroll', 'touchstart']
    activityEvents.forEach((evt) => window.addEventListener(evt, resetTimer, { passive: true }))

    const interval = setInterval(() => {
      if (Date.now() - lastActivityRef.current > timeoutMs) {
        lockApp()
      }
    }, 15000)

    return () => {
      activityEvents.forEach((evt) => window.removeEventListener(evt, resetTimer))
      clearInterval(interval)
    }
  }, [authStatus?.auth_enabled, authStatus?.auto_lock_minutes, isLocked, lockApp])

  return (
    <AuthContext.Provider
      value={{
        authStatus,
        isLoading,
        isLocked,
        lockApp,
        unlock,
        refetchAuth: refetch,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
