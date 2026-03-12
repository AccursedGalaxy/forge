import { useState, useEffect, useRef } from 'react'
import type { SessionStatus, StreamLine, StreamLineType, PlanStep } from '../types'
import { createSSEClient, type ClaudeStreamData, type ClaudeDoneData } from '../lib/sse'
import log from '../lib/logger'

let lineIdCounter = 0

function classifyStreamData(data: ClaudeStreamData): StreamLineType {
  switch (data.type) {
    case 'tool': return 'tool'
    case 'thinking': return 'thinking'
    case 'error': return 'error'
    default: return 'default'
  }
}

interface UseStreamResult {
  lines: StreamLine[]
  isConnected: boolean
  isRunning: boolean
  status: SessionStatus | null
  phase: string | null
  planSteps: PlanStep[]
  notes: string | null
}

export function useStream(sessionId: string | null): UseStreamResult {
  const [lines, setLines] = useState<StreamLine[]>([])
  const [isConnected, setIsConnected] = useState(false)
  const [status, setStatus] = useState<SessionStatus | null>(null)
  const [phase, setPhase] = useState<string | null>(null)
  const [planSteps, setPlanSteps] = useState<PlanStep[]>([])
  const [notes, setNotes] = useState<string | null>(null)
  const clientRef = useRef<ReturnType<typeof createSSEClient> | null>(null)

  useEffect(() => {
    if (!sessionId) {
      setLines([])
      setIsConnected(false)
      setStatus(null)
      setPhase(null)
      setPlanSteps([])
      setNotes(null)
      return
    }

    log.info('useStream', 'Connecting to session stream', { sessionId })

    const addLine = (text: string, type: StreamLineType = 'default') => {
      setLines((prev) => [
        ...prev,
        {
          id: String(++lineIdCounter),
          type,
          text,
          timestamp: Date.now(),
        },
      ])
    }

    const client = createSSEClient(sessionId, {
      onStart: (p) => {
        setIsConnected(true)
        setPhase(p)
        addLine(`▶ Starting ${p} phase…`, 'success')
      },
      onStream: (data: ClaudeStreamData) => {
        if (data.content) {
          addLine(data.content, classifyStreamData(data))
        }
      },
      onDone: (data: ClaudeDoneData) => {
        setIsConnected(false)
        setPhase(data.phase)
        if (data.plan_steps && data.plan_steps.length > 0) {
          setPlanSteps(data.plan_steps)
        }
        if (data.notes) {
          setNotes(data.notes)
          addLine(`✓ ${data.notes}`, 'success')
        } else {
          addLine(`✓ ${data.phase} phase complete`, 'success')
        }
      },
      onError: (msg) => {
        addLine(`✗ ${msg}`, 'error')
        setIsConnected(false)
        log.error('useStream', 'Stream error', { msg })
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

  return { lines, isConnected, isRunning, status, phase, planSteps, notes }
}
