import type {
  Project,
  Task,
  Session,
  ContextChunk,
  AgentProvider,
  TaskStatus,
  AutonomyLevel,
  SessionType,
} from '../types'

const BASE_URL = (import.meta.env.VITE_API_URL as string | undefined) ?? 'http://localhost:8080'
const FORGE_KEY = import.meta.env.VITE_FORGE_KEY as string | undefined

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(FORGE_KEY ? { 'X-Forge-Key': FORGE_KEY } : {}),
    ...(init.headers ?? {}),
  }

  const delays = [200, 400, 800]
  let lastError: Error = new Error('Unknown error')

  for (let attempt = 0; attempt <= delays.length; attempt++) {
    try {
      const res = await fetch(`${BASE_URL}${path}`, { ...init, headers })
      if (!res.ok) {
        let message = res.statusText
        try {
          const body = await res.json()
          if (body.error) message = body.error
          else if (body.message) message = body.message
        } catch {
          // ignore parse error
        }
        throw new ApiError(res.status, message)
      }
      if (res.status === 204) return undefined as T
      return res.json() as Promise<T>
    } catch (err) {
      if (err instanceof ApiError) throw err
      lastError = err as Error
      if (attempt < delays.length) {
        await new Promise((r) => setTimeout(r, delays[attempt]))
      }
    }
  }

  throw lastError
}

// Projects
export const getProjects = () => request<Project[]>('/api/projects')
export const createProject = (data: { name: string; description?: string; repo_url: string }) =>
  request<Project>('/api/projects', { method: 'POST', body: JSON.stringify(data) })
export const updateProject = (id: string, data: Partial<Project>) =>
  request<Project>(`/api/projects/${id}`, { method: 'PATCH', body: JSON.stringify(data) })
export const deleteProject = (id: string) =>
  request<void>(`/api/projects/${id}`, { method: 'DELETE' })

// Tasks
export const getTasks = (projectId: string) =>
  request<Task[]>(`/api/projects/${projectId}/tasks`)
export const createTask = (
  projectId: string,
  data: { title: string; description?: string; autonomy_level?: AutonomyLevel },
) =>
  request<Task>(`/api/projects/${projectId}/tasks`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
export const updateTask = (id: string, data: Partial<Omit<Task, 'id' | 'project_id' | 'created_at'>>) =>
  request<Task>(`/api/tasks/${id}`, { method: 'PATCH', body: JSON.stringify(data) })
export const deleteTask = (id: string) =>
  request<void>(`/api/tasks/${id}`, { method: 'DELETE' })
export const reorderTask = (id: string, data: { status: TaskStatus; position: number }) =>
  request<Task>(`/api/tasks/${id}/reorder`, { method: 'PATCH', body: JSON.stringify(data) })

// Sessions
export const getSessions = (taskId: string) =>
  request<Session[]>(`/api/tasks/${taskId}/sessions`)
export const createSession = (data: { task_id: string; project_id: string; session_type: SessionType }) =>
  request<Session>('/api/sessions', { method: 'POST', body: JSON.stringify(data) })
export const getSession = (id: string) => request<Session>(`/api/sessions/${id}`)
export const approveSession = (id: string) =>
  request<Session>(`/api/sessions/${id}/approve`, { method: 'POST' })
export const interruptSession = (id: string) =>
  request<void>(`/api/sessions/${id}/interrupt`, { method: 'POST' })
export const resumeSession = (id: string, prompt: string) =>
  request<Session>(`/api/sessions/${id}/resume`, {
    method: 'POST',
    body: JSON.stringify({ prompt }),
  })

// Context
export const getContext = (projectId: string) =>
  request<ContextChunk[]>(`/api/projects/${projectId}/context`)
export const deleteContextChunk = (id: string) =>
  request<void>(`/api/context/${id}`, { method: 'DELETE' })

// Providers
export const getProviders = () => request<AgentProvider[]>('/api/providers')
export const createProvider = (data: Omit<AgentProvider, 'id'>) =>
  request<AgentProvider>('/api/providers', { method: 'POST', body: JSON.stringify(data) })
export const updateProvider = (id: string, data: Partial<AgentProvider>) =>
  request<AgentProvider>(`/api/providers/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })
