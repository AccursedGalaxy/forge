import { useState, useEffect, useCallback } from 'react'
import type { Session } from '../types'
import * as api from '../lib/api'
import log from '../lib/logger'

interface UseSessionResult {
  session: Session | null
  isLoading: boolean
  error: Error | null
  approve: () => Promise<Session | null>
  interrupt: () => Promise<boolean>
  resume: (prompt: string) => Promise<Session | null>
}

export function useSession(sessionId: string | null): UseSessionResult {
  const [session, setSession] = useState<Session | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    if (!sessionId) {
      setSession(null)
      setIsLoading(false)
      return
    }

    let cancelled = false
    setIsLoading(true)
    setError(null)

    api
      .getSession(sessionId)
      .then((data) => {
        if (!cancelled) {
          setSession(data)
          setIsLoading(false)
        }
      })
      .catch((err: Error) => {
        if (!cancelled) {
          log.error('useSession', 'Failed to fetch session', err)
          setError(err)
          setSession(null)
          setIsLoading(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [sessionId])

  const approve = useCallback(async () => {
    if (!sessionId) return null
    try {
      const updated = await api.approveSession(sessionId)
      setSession(updated)
      return updated
    } catch (err) {
      log.error('useSession', 'Failed to approve session', err)
      return null
    }
  }, [sessionId])

  const interrupt = useCallback(async () => {
    if (!sessionId) return false
    try {
      await api.interruptSession(sessionId)
      return true
    } catch (err) {
      log.error('useSession', 'Failed to interrupt session', err)
      return false
    }
  }, [sessionId])

  const resume = useCallback(
    async (prompt: string) => {
      if (!sessionId) return null
      try {
        const updated = await api.resumeSession(sessionId, prompt)
        setSession(updated)
        return updated
      } catch (err) {
        log.error('useSession', 'Failed to resume session', err)
        return null
      }
    },
    [sessionId],
  )

  return { session, isLoading, error, approve, interrupt, resume }
}
