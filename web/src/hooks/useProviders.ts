import { useState, useEffect, useCallback } from 'react'
import type { AgentProvider } from '../types'
import * as api from '../lib/api'
import log from '../lib/logger'

interface UseProvidersResult {
  providers: AgentProvider[]
  isLoading: boolean
  error: Error | null
  createProvider: (data: Omit<AgentProvider, 'id'>) => Promise<AgentProvider | null>
  updateProvider: (id: string, data: Partial<AgentProvider>) => Promise<AgentProvider | null>
  refetch: () => void
}

export function useProviders(): UseProvidersResult {
  const [providers, setProviders] = useState<AgentProvider[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [tick, setTick] = useState(0)

  const refetch = useCallback(() => setTick((t) => t + 1), [])

  useEffect(() => {
    let cancelled = false
    setIsLoading(true)
    setError(null)

    api
      .getProviders()
      .then((data) => {
        if (!cancelled) {
          setProviders(data)
          setIsLoading(false)
        }
      })
      .catch((err: Error) => {
        if (!cancelled) {
          log.error('useProviders', 'Failed to fetch providers', err)
          setError(err)
          setProviders([])
          setIsLoading(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [tick])

  const createProvider = useCallback(async (data: Omit<AgentProvider, 'id'>) => {
    try {
      const provider = await api.createProvider(data)
      setProviders((prev) => [...prev, provider])
      return provider
    } catch (err) {
      log.error('useProviders', 'Failed to create provider', err)
      return null
    }
  }, [])

  const updateProvider = useCallback(async (id: string, data: Partial<AgentProvider>) => {
    try {
      const updated = await api.updateProvider(id, data)
      setProviders((prev) => prev.map((p) => (p.id === id ? updated : p)))
      return updated
    } catch (err) {
      log.error('useProviders', 'Failed to update provider', err)
      return null
    }
  }, [])

  return { providers, isLoading, error, createProvider, updateProvider, refetch }
}
