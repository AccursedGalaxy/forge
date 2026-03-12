import { useState, useEffect, useCallback } from 'react'
import type { Project } from '../types'
import * as api from '../lib/api'
import log from '../lib/logger'

interface UseProjectsResult {
  projects: Project[]
  isLoading: boolean
  error: Error | null
  createProject: (data: { name: string; description?: string; repo_url: string }) => Promise<Project | null>
  deleteProject: (id: string) => Promise<boolean>
  refetch: () => void
}

export function useProjects(): UseProjectsResult {
  const [projects, setProjects] = useState<Project[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [tick, setTick] = useState(0)

  const refetch = useCallback(() => setTick((t) => t + 1), [])

  useEffect(() => {
    let cancelled = false
    setIsLoading(true)
    setError(null)

    api
      .getProjects()
      .then((data) => {
        if (!cancelled) {
          setProjects(data)
          setIsLoading(false)
        }
      })
      .catch((err: Error) => {
        if (!cancelled) {
          log.error('useProjects', 'Failed to fetch projects', err)
          setError(err)
          setProjects([])
          setIsLoading(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [tick])

  const createProject = useCallback(
    async (data: { name: string; description?: string; repo_url: string }) => {
      try {
        const project = await api.createProject(data)
        setProjects((prev) => [...prev, project])
        return project
      } catch (err) {
        log.error('useProjects', 'Failed to create project', err)
        return null
      }
    },
    [],
  )

  const deleteProject = useCallback(async (id: string) => {
    try {
      await api.deleteProject(id)
      setProjects((prev) => prev.filter((p) => p.id !== id))
      return true
    } catch (err) {
      log.error('useProjects', 'Failed to delete project', err)
      return false
    }
  }, [])

  return { projects, isLoading, error, createProject, deleteProject, refetch }
}
