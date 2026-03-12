import type { SessionStatus, StreamEventType } from '../types'

const BASE_URL = (import.meta.env.VITE_API_URL as string | undefined) ?? 'http://localhost:8080'
const FORGE_KEY = import.meta.env.VITE_FORGE_KEY as string | undefined

export interface SSEHandlers {
  onStart?: (data: string) => void
  onStream?: (data: string) => void
  onDone?: (data: string) => void
  onError?: (data: string) => void
  onStatusChange?: (status: SessionStatus) => void
}

export interface SSEClient {
  connect: () => void
  disconnect: () => void
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
      // Ignore SSE heartbeat comment lines (they come as empty data)
      if (!e.data || e.data === '') return

      try {
        const event = JSON.parse(e.data) as { type: StreamEventType; data: string; status?: SessionStatus }

        switch (event.type) {
          case 'claude:start':
            handlers.onStart?.(event.data)
            break
          case 'claude:stream':
            handlers.onStream?.(event.data)
            break
          case 'claude:done':
            handlers.onDone?.(event.data)
            break
          case 'claude:error':
            handlers.onError?.(event.data)
            break
          case 'session:status':
            if (event.status) handlers.onStatusChange?.(event.status)
            break
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
