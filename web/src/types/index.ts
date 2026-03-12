export type TaskStatus = 'backlog' | 'planned' | 'in_progress' | 'review' | 'done'
export type AutonomyLevel = 'supervised' | 'checkpoint' | 'autonomous'
export type SessionStatus =
  | 'pending'
  | 'planning'
  | 'awaiting_approval'
  | 'running'
  | 'paused'
  | 'done'
  | 'error'
export type SessionType = 'plan' | 'execute'
export type StreamEventType =
  | 'claude:start'
  | 'claude:stream'
  | 'claude:done'
  | 'claude:error'
  | 'session:status'

export interface Project {
  id: string
  owner_id: string
  name: string
  description: string | null
  repo_url: string
  created_at: string
}

export interface Task {
  id: string
  project_id: string
  title: string
  description: string | null
  status: TaskStatus
  autonomy_level: AutonomyLevel
  position: number
  created_at: string
}

export interface Session {
  id: string
  task_id: string
  project_id: string
  session_type: SessionType
  status: SessionStatus
  claude_session_id: string | null
  plan_steps: PlanStep[] | null
  claude_notes: string | null
  error_message: string | null
  started_at: string | null
  completed_at: string | null
  created_at: string
}

export interface PlanStep {
  index: number
  description: string
  completed: boolean
}

export interface ContextChunk {
  id: string
  project_id: string
  session_id: string | null
  chunk_type: 'session_output' | 'code_diff' | 'task_note'
  content: string
  created_at: string
}

export interface AgentProvider {
  id: string
  name: string
  provider_type: 'claude' | 'codex' | 'ollama'
  bin_path: string | null
  is_default: boolean
  config: Record<string, unknown> | null
}

export interface StreamEvent {
  type: StreamEventType
  data: string
  session_id?: string
  status?: SessionStatus
}

export type StreamLineType = 'default' | 'tool' | 'thinking' | 'error' | 'success'

export interface StreamLine {
  id: string
  type: StreamLineType
  text: string
  timestamp: number
}
