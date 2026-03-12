import { useEffect, useRef, useState, useCallback, type CSSProperties } from 'react'

const BASE_URL = (import.meta.env.VITE_API_URL as string | undefined) ?? 'http://localhost:8080'

type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR'

interface LogLine {
  id: string
  time: string
  level: LogLevel
  msg: string
  attrs: Record<string, unknown>
}

type LevelFilter = LogLevel | 'ALL'

const LEVEL_ORDER: Record<LogLevel, number> = { DEBUG: 0, INFO: 1, WARN: 2, ERROR: 3 }

const LEVEL_COLORS: Record<LogLevel, string> = {
  DEBUG: 'var(--text-muted)',
  INFO: 'var(--accent)',
  WARN: '#f59e0b',
  ERROR: '#ef4444',
}

const LEVEL_BG: Record<LogLevel, string> = {
  DEBUG: 'rgba(113,113,122,0.12)',
  INFO: 'var(--accent-dim)',
  WARN: 'rgba(245,158,11,0.12)',
  ERROR: 'rgba(239,68,68,0.12)',
}

function parseLevel(raw: string): LogLevel {
  const up = raw.toUpperCase()
  if (up === 'DEBUG' || up === 'WARN' || up === 'ERROR') return up
  return 'INFO'
}

function parseLine(raw: string, id: string): LogLine | null {
  try {
    const obj = JSON.parse(raw) as Record<string, unknown>
    const { time, level, msg, ...rest } = obj
    return {
      id,
      time: typeof time === 'string' ? time : '',
      level: parseLevel(typeof level === 'string' ? level : 'INFO'),
      msg: typeof msg === 'string' ? msg : String(msg ?? ''),
      attrs: rest,
    }
  } catch {
    return null
  }
}

function shortTime(iso: string): string {
  if (!iso) return ''
  try {
    const d = new Date(iso)
    return d.toLocaleTimeString('en-US', { hour12: false }) + '.' + String(d.getMilliseconds()).padStart(3, '0')
  } catch {
    return iso.slice(11, 23)
  }
}

function attrsString(attrs: Record<string, unknown>): string {
  return Object.entries(attrs)
    .map(([k, v]) => `${k}=${JSON.stringify(v)}`)
    .join(' ')
}

let lineCounter = 0
function nextId() {
  return `ll-${++lineCounter}`
}

const MAX_LINES = 2000

interface LogViewerProps {
  /** compact=true: fill parent height (used inside the AppShell drawer) */
  compact?: boolean
}

export function LogViewer({ compact = false }: LogViewerProps) {
  const [lines, setLines] = useState<LogLine[]>([])
  const [filter, setFilter] = useState<LevelFilter>('ALL')
  const [autoScroll, setAutoScroll] = useState(true)
  const [status, setStatus] = useState<'connecting' | 'connected' | 'error'>('connecting')
  const bottomRef = useRef<HTMLDivElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  // SSE connection
  useEffect(() => {
    const url = `${BASE_URL}/api/logs/stream`
    let es: EventSource
    let dead = false

    function connect() {
      if (dead) return
      setStatus('connecting')
      es = new EventSource(url)

      es.onopen = () => setStatus('connected')

      es.onmessage = (e: MessageEvent) => {
        if (!e.data) return
        const line = parseLine(e.data as string, nextId())
        if (!line) return
        setLines((prev) => {
          const next = [...prev, line]
          return next.length > MAX_LINES ? next.slice(next.length - MAX_LINES) : next
        })
      }

      es.onerror = () => {
        es.close()
        setStatus('error')
        if (!dead) setTimeout(connect, 3000)
      }
    }

    connect()
    return () => {
      dead = true
      es?.close()
    }
  }, [])

  // Auto-scroll
  useEffect(() => {
    if (autoScroll) {
      bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
    }
  }, [lines, autoScroll])

  // Pause auto-scroll when user scrolls up
  const handleScroll = useCallback(() => {
    const el = containerRef.current
    if (!el) return
    const isAtBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 40
    setAutoScroll(isAtBottom)
  }, [])

  const filtered = filter === 'ALL'
    ? lines
    : lines.filter((l) => LEVEL_ORDER[l.level] >= LEVEL_ORDER[filter])

  // ── Styles ──────────────────────────────────────────────────────────────────

  const wrapStyle: CSSProperties = {
    display: 'flex',
    flexDirection: 'column',
    height: compact ? '100%' : 'calc(100vh - var(--topbar-height) - 48px)',
    gap: compact ? '6px' : '12px',
  }

  const toolbarStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    flexShrink: 0,
  }

  const dotStyle = (s: typeof status): CSSProperties => ({
    width: '7px',
    height: '7px',
    borderRadius: '50%',
    background: s === 'connected' ? '#22c55e' : s === 'connecting' ? '#f59e0b' : '#ef4444',
    flexShrink: 0,
  })

  const statusTextStyle: CSSProperties = {
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    color: 'var(--text-muted)',
    marginRight: 'auto',
  }

  const filterBtnStyle = (active: boolean): CSSProperties => ({
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    fontWeight: 600,
    padding: '3px 8px',
    borderRadius: 'var(--radius-md)',
    border: '1px solid ' + (active ? 'var(--border-default)' : 'var(--border-subtle)'),
    background: active ? 'var(--bg-elevated)' : 'transparent',
    color: active ? 'var(--text-primary)' : 'var(--text-muted)',
    cursor: 'pointer',
  })

  const clearBtnStyle: CSSProperties = {
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    padding: '3px 8px',
    borderRadius: 'var(--radius-md)',
    border: '1px solid var(--border-subtle)',
    background: 'transparent',
    color: 'var(--text-muted)',
    cursor: 'pointer',
  }

  const logAreaStyle: CSSProperties = {
    flex: 1,
    overflow: 'auto',
    background: 'var(--bg-surface)',
    border: '1px solid var(--border-subtle)',
    borderRadius: 'var(--radius-lg)',
    padding: '8px 0',
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    lineHeight: '1.6',
  }

  const lineStyle = (level: LogLevel): CSSProperties => ({
    display: 'flex',
    alignItems: 'baseline',
    gap: '10px',
    padding: '1px 12px',
    borderLeft: '2px solid transparent',
    borderLeftColor: level === 'ERROR' ? '#ef444430' : level === 'WARN' ? '#f59e0b30' : 'transparent',
  })

  const timeStyle: CSSProperties = {
    color: 'var(--text-disabled)',
    flexShrink: 0,
    width: '96px',
  }

  const badgeStyle = (level: LogLevel): CSSProperties => ({
    fontSize: '10px',
    fontWeight: 700,
    padding: '1px 5px',
    borderRadius: 'var(--radius-sm)',
    color: LEVEL_COLORS[level],
    background: LEVEL_BG[level],
    flexShrink: 0,
    width: '40px',
    textAlign: 'center',
  })

  const msgStyle: CSSProperties = {
    color: 'var(--text-primary)',
    flex: 1,
    wordBreak: 'break-all',
  }

  const attrsStyle: CSSProperties = {
    color: 'var(--text-muted)',
    flexShrink: 0,
    maxWidth: '55%',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  }

  const emptyStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    height: '100%',
    color: 'var(--text-disabled)',
    fontFamily: 'var(--font-mono)',
    fontSize: '13px',
  }

  return (
    <div style={wrapStyle}>
      {/* Toolbar */}
      <div style={toolbarStyle}>
        <div style={dotStyle(status)} />
        <span style={statusTextStyle}>
          {status === 'connected' ? `${filtered.length} lines` : status}
        </span>

        {(['ALL', 'DEBUG', 'INFO', 'WARN', 'ERROR'] as LevelFilter[]).map((l) => (
          <button key={l} style={filterBtnStyle(filter === l)} onClick={() => setFilter(l)}>
            {l}
          </button>
        ))}

        <button
          style={filterBtnStyle(autoScroll)}
          onClick={() => setAutoScroll((v) => !v)}
          title="Toggle auto-scroll"
        >
          ↓
        </button>

        <button style={clearBtnStyle} onClick={() => setLines([])}>
          clear
        </button>
      </div>

      {/* Log area */}
      <div ref={containerRef} style={logAreaStyle} onScroll={handleScroll}>
        {filtered.length === 0 ? (
          <div style={emptyStyle}>
            {status === 'connecting' ? 'connecting to log stream…' : 'no log lines yet'}
          </div>
        ) : (
          filtered.map((line) => (
            <div key={line.id} style={lineStyle(line.level)}>
              <span style={timeStyle}>{shortTime(line.time)}</span>
              <span style={badgeStyle(line.level)}>{line.level}</span>
              <span style={msgStyle}>{line.msg}</span>
              {Object.keys(line.attrs).length > 0 && (
                <span style={attrsStyle} title={attrsString(line.attrs)}>
                  {attrsString(line.attrs)}
                </span>
              )}
            </div>
          ))
        )}
        <div ref={bottomRef} />
      </div>
    </div>
  )
}

export default LogViewer
