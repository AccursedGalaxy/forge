import { useState, useEffect, useRef } from 'react'
import type { SessionStatus, StreamLine, StreamLineType } from '../types'
import { createSSEClient } from '../lib/sse'
import log from '../lib/logger'

let lineIdCounter = 0

function classifyLine(text: string): StreamLineType {
  if (text.startsWith('⚡') || text.includes('[tool]') || text.includes('Tool:')) return 'tool'
  if (text.startsWith('💭') || text.includes('[thinking]')) return 'thinking'
  if (text.startsWith('✗') || text.toLowerCase().startsWith('error')) return 'error'
  if (text.startsWith('✓') || text.toLowerCase().startsWith('success')) return 'success'
  return 'default'
}

interface UseStreamResult {
  lines: StreamLine[]
  isConnected: boolean
  isRunning: boolean
  status: SessionStatus | null
}

export function useStream(sessionId: string | null): UseStreamResult {
  const [lines, setLines] = useState<StreamLine[]>([])
  const [isConnected, setIsConnected] = useState(false)
  const [status, setStatus] = useState<SessionStatus | null>(null)
  const clientRef = useRef<ReturnType<typeof createSSEClient> | null>(null)

  useEffect(() => {
    if (!sessionId) {
      setLines([])
      setIsConnected(false)
      setStatus(null)
      return
    }

    log.info('useStream', 'Connecting to session stream', { sessionId })

    const addLine = (text: string) => {
      setLines((prev) => [
        ...prev,
        {
          id: String(++lineIdCounter),
          type: classifyLine(text),
          text,
          timestamp: Date.now(),
        },
      ])
    }

    const client = createSSEClient(sessionId, {
      onStart: (data) => {
        setIsConnected(true)
        addLine(data)
      },
      onStream: (data) => addLine(data),
      onDone: (data) => {
        addLine(data)
        setIsConnected(false)
      },
      onError: (data) => {
        addLine(`✗ ${data}`)
        setIsConnected(false)
        log.error('useStream', 'Stream error', { data })
      },
      onStatusChange: (s) => setStatus(s),
    })

    clientRef.current = client
    client.connect()

    return () => {
      client.disconnect()
      clientRef.current = null
      setIsConnected(false)
    }
  }, [sessionId])

  const isRunning = status === 'running' || status === 'planning'

  return { lines, isConnected, isRunning, status }
}
