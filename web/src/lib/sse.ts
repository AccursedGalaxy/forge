import type { SessionStatus, StreamEventType, PlanStep } from '../types'

const BASE_URL = (import.meta.env.VITE_API_URL as string | undefined) ?? 'http://localhost:8080'
const FORGE_KEY = import.meta.env.VITE_FORGE_KEY as string | undefined

// Structured data payloads from each event type
export interface ClaudeStreamData {
  type: string   // "text" | "thinking" | "tool" | "error" | "done"
  content: string
}

export interface ClaudeDoneData {
  phase: 'plan' | 'execute' | 'resume'
  plan_steps?: PlanStep[]
  notes?: string
}

export interface SSEHandlers {
  onStart?: (phase: string) => void
  onStream?: (data: ClaudeStreamData) => void
  onDone?: (data: ClaudeDoneData) => void
  onError?: (data: string) => void
  onStatusChange?: (status: SessionStatus) => void
}

export interface SSEClient {
  connect: () => void
  disconnect: () => void
}

interface RawSSEEvent {
  type: StreamEventType
  data: unknown
}

export function createSSEClient(sessionId: string, handlers: SSEHandlers): SSEClient {
  let source: EventSource | null = null
  let attempts = 0
  const MAX_ATTEMPTS = 5
  let disconnected = false

  function connect() {
    if (disconnected) return

    const url = new URL(`${BASE_URL}/api/sessions/${sessionId}/stream`)
    if (FORGE_KEY) url.searchParams.set('forge_key', FORGE_KEY)

    source = new EventSource(url.toString())

    source.onmessage = (e: MessageEvent) => {
      if (!e.data || e.data === '') return

      try {
        const event = JSON.parse(e.data) as RawSSEEvent

        switch (event.type) {
          case 'claude:start': {
            const d = event.data as { phase?: string }
            handlers.onStart?.(d?.phase ?? '')
            break
          }
          case 'claude:stream': {
            const d = event.data as ClaudeStreamData
            handlers.onStream?.(d)
            break
          }
          case 'claude:done': {
            const d = event.data as ClaudeDoneData
            handlers.onDone?.(d)
            break
          }
          case 'claude:error': {
            const d = event.data as { content?: string }
            handlers.onError?.(d?.content ?? String(event.data))
            break
          }
          case 'session:status': {
            const d = event.data as { status?: SessionStatus }
            if (d?.status) handlers.onStatusChange?.(d.status)
            break
          }
        }
      } catch {
        // Malformed event — ignore
      }
    }

    source.onerror = () => {
      source?.close()
      source = null

      if (disconnected) return

      if (attempts < MAX_ATTEMPTS) {
        attempts++
        setTimeout(connect, 1000 * attempts)
      }
    }

    source.onopen = () => {
      attempts = 0
    }
  }

  function disconnect() {
    disconnected = true
    source?.close()
    source = null
  }

  return { connect, disconnect }
}
