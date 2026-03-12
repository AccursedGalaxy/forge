import { useState, useEffect, useCallback } from 'react'
import type { Task, TaskStatus, AutonomyLevel } from '../types'
import * as api from '../lib/api'
import log from '../lib/logger'

type TasksByStatus = Record<TaskStatus, Task[]>

const ALL_STATUSES: TaskStatus[] = ['backlog', 'planned', 'in_progress', 'review', 'done']

function groupByStatus(tasks: Task[]): TasksByStatus {
  const grouped: TasksByStatus = {
    backlog: [],
    planned: [],
    in_progress: [],
    review: [],
    done: [],
  }
  for (const task of tasks) {
    grouped[task.status].push(task)
  }
  for (const status of ALL_STATUSES) {
    grouped[status].sort((a, b) => a.position - b.position)
  }
  return grouped
}

interface UseTasksResult {
  tasks: Task[]
  tasksByStatus: TasksByStatus
  isLoading: boolean
  error: Error | null
  createTask: (data: {
    title: string
    description?: string
    autonomy_level?: AutonomyLevel
  }) => Promise<Task | null>
  updateTask: (id: string, data: Partial<Omit<Task, 'id' | 'project_id' | 'created_at'>>) => Promise<Task | null>
  deleteTask: (id: string) => Promise<boolean>
  reorderTask: (id: string, status: TaskStatus, position: number) => Promise<Task | null>
  refetch: () => void
}

export function useTasks(projectId: string | null): UseTasksResult {
  const [tasks, setTasks] = useState<Task[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)
  const [tick, setTick] = useState(0)

  const refetch = useCallback(() => setTick((t) => t + 1), [])

  useEffect(() => {
    if (!projectId) {
      setTasks([])
      setIsLoading(false)
      return
    }

    let cancelled = false
    setIsLoading(true)
    setError(null)

    api
      .getTasks(projectId)
      .then((data) => {
        if (!cancelled) {
          setTasks(data)
          setIsLoading(false)
        }
      })
      .catch((err: Error) => {
        if (!cancelled) {
          log.error('useTasks', 'Failed to fetch tasks', err)
          setError(err)
          setTasks([])
          setIsLoading(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [projectId, tick])

  const createTask = useCallback(
    async (data: { title: string; description?: string; autonomy_level?: AutonomyLevel }) => {
      if (!projectId) return null
      try {
        const task = await api.createTask(projectId, data)
        setTasks((prev) => [...prev, task])
        return task
      } catch (err) {
        log.error('useTasks', 'Failed to create task', err)
        return null
      }
    },
    [projectId],
  )

  const updateTask = useCallback(
    async (id: string, data: Partial<Omit<Task, 'id' | 'project_id' | 'created_at'>>) => {
      try {
        const updated = await api.updateTask(id, data)
        setTasks((prev) => prev.map((t) => (t.id === id ? updated : t)))
        return updated
      } catch (err) {
        log.error('useTasks', 'Failed to update task', err)
        return null
      }
    },
    [],
  )

  const deleteTask = useCallback(async (id: string) => {
    try {
      await api.deleteTask(id)
      setTasks((prev) => prev.filter((t) => t.id !== id))
      return true
    } catch (err) {
      log.error('useTasks', 'Failed to delete task', err)
      return false
    }
  }, [])

  const reorderTask = useCallback(async (id: string, status: TaskStatus, position: number) => {
    try {
      const updated = await api.reorderTask(id, { status, position })
      setTasks((prev) => prev.map((t) => (t.id === id ? updated : t)))
      return updated
    } catch (err) {
      log.error('useTasks', 'Failed to reorder task', err)
      return null
    }
  }, [])

  return {
    tasks,
    tasksByStatus: groupByStatus(tasks),
    isLoading,
    error,
    createTask,
    updateTask,
    deleteTask,
    reorderTask,
    refetch,
  }
}
