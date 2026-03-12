import { useEffect, useRef, type CSSProperties } from 'react'
import type { StreamLine } from '../../types'
import { Button } from '../ui'

interface SessionStreamProps {
  sessionId: string | null
  isOpen: boolean
  onClose: () => void
  taskTitle: string
  lines: StreamLine[]
  isRunning: boolean
  onInterrupt: () => void
}

const slideKeyframes = `
@keyframes forge-slide-in {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
}
`

if (typeof document !== 'undefined' && !document.getElementById('forge-session-stream-styles')) {
  const style = document.createElement('style')
  style.id = 'forge-session-stream-styles'
  style.textContent = slideKeyframes
  document.head.appendChild(style)
}

const lineColors: Record<StreamLine['type'], string> = {
  default: 'var(--text-secondary)',
  tool: 'var(--accent)',
  thinking: 'var(--text-muted)',
  error: '#ef4444',
  success: '#22c55e',
}

export function SessionStream({
  sessionId,
  isOpen,
  onClose,
  taskTitle,
  lines,
  isRunning,
  onInterrupt,
}: SessionStreamProps) {
  const streamRef = useRef<HTMLDivElement>(null)

  // Auto-scroll to bottom when new lines arrive
  useEffect(() => {
    if (streamRef.current) {
      streamRef.current.scrollTop = streamRef.current.scrollHeight
    }
  }, [lines])

  const panelStyle: CSSProperties = {
    position: 'fixed',
    right: 0,
    top: 0,
    height: '100vh',
    width: 'var(--session-panel-width)',
    background: 'var(--bg-elevated)',
    borderLeft: '1px solid var(--border-default)',
    display: 'flex',
    flexDirection: 'column',
    zIndex: 100,
    transform: isOpen ? 'translateX(0)' : 'translateX(100%)',
    transition: `transform var(--duration-slow) var(--ease-default)`,
  }

  const headerStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '12px 16px',
    borderBottom: '1px solid var(--border-subtle)',
    flexShrink: 0,
  }

  const titleStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
    fontWeight: 500,
    color: 'var(--text-primary)',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
    flex: 1,
    marginRight: '8px',
  }

  const closeButtonStyle: CSSProperties = {
    background: 'none',
    border: 'none',
    cursor: 'pointer',
    color: 'var(--text-muted)',
    fontSize: '16px',
    lineHeight: 1,
    padding: '4px',
    borderRadius: 'var(--radius-sm)',
    flexShrink: 0,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    transition: 'color var(--duration-fast)',
  }

  const streamAreaStyle: CSSProperties = {
    flex: 1,
    overflowY: 'auto',
    background: '#0D0D10',
    padding: '12px 14px',
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    lineHeight: 1.6,
  }

  const emptyStateStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    height: '100%',
    color: 'var(--text-muted)',
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
  }

  const footerStyle: CSSProperties = {
    padding: '12px 16px',
    borderTop: '1px solid var(--border-subtle)',
    flexShrink: 0,
  }

  return (
    <div style={panelStyle} aria-hidden={!isOpen}>
      <div style={headerStyle}>
        <span style={titleStyle}>{taskTitle || 'Session'}</span>
        <button style={closeButtonStyle} onClick={onClose} aria-label="Close session panel">
          ✕
        </button>
      </div>

      <div ref={streamRef} style={streamAreaStyle}>
        {!sessionId ? (
          <div style={emptyStateStyle}>No active session</div>
        ) : lines.length === 0 ? (
          <div style={{ color: 'var(--text-muted)' }}>Waiting for output…</div>
        ) : (
          lines.map((line) => (
            <div
              key={line.id}
              style={{
                color: lineColors[line.type],
                fontStyle: line.type === 'thinking' ? 'italic' : 'normal',
                marginBottom: '2px',
                wordBreak: 'break-word',
              }}
            >
              {line.text}
            </div>
          ))
        )}
      </div>

      {isRunning && (
        <div style={footerStyle}>
          <Button variant="danger" size="sm" onClick={onInterrupt}>
            Interrupt
          </Button>
        </div>
      )}
    </div>
  )
}

export default SessionStream
