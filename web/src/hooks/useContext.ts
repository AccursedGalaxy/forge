import { useState, useEffect, useCallback } from 'react'
import type { ContextChunk } from '../types'
import * as api from '../lib/api'
import log from '../lib/logger'

interface UseContextResult {
  chunks: ContextChunk[]
  isLoading: boolean
  error: Error | null
  deleteChunk: (id: string) => Promise<boolean>
  refetch: () => void
}

export function useContext(projectId: string | null): UseContextResult {
  const [chunks, setChunks] = useState<ContextChunk[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)
  const [tick, setTick] = useState(0)

  const refetch = useCallback(() => setTick((t) => t + 1), [])

  useEffect(() => {
    if (!projectId) {
      setChunks([])
      setIsLoading(false)
      return
    }

    let cancelled = false
    setIsLoading(true)
    setError(null)

    api
      .getContext(projectId)
      .then((data) => {
        if (!cancelled) {
          setChunks(data)
          setIsLoading(false)
        }
      })
      .catch((err: Error) => {
        if (!cancelled) {
          log.error('useContext', 'Failed to fetch context', err)
          setError(err)
          setChunks([])
          setIsLoading(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [projectId, tick])

  const deleteChunk = useCallback(async (id: string) => {
    try {
      await api.deleteContextChunk(id)
      setChunks((prev) => prev.filter((c) => c.id !== id))
      return true
    } catch (err) {
      log.error('useContext', 'Failed to delete context chunk', err)
      return false
    }
  }, [])

  return { chunks, isLoading, error, deleteChunk, refetch }
}
