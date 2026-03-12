import { useEffect, useRef, type CSSProperties } from 'react'
import type { StreamLine, SessionStatus, PlanStep } from '../../types'
import { Button } from '../ui'
import { PlanApprovalPanel } from './PlanApprovalPanel'

interface SessionStreamProps {
  sessionId: string | null
  isOpen: boolean
  onClose: () => void
  taskTitle: string
  lines: StreamLine[]
  isRunning: boolean
  status: SessionStatus | null
  phase: string | null
  planSteps: PlanStep[]
  notes: string | null
  isApproving: boolean
  onInterrupt: () => void
  onApprove: () => void
  onReject: () => void
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

const phaseLabels: Record<string, string> = {
  plan: 'Planning',
  execute: 'Executing',
  resume: 'Resuming',
}

export function SessionStream({
  sessionId,
  isOpen,
  onClose,
  taskTitle,
  lines,
  isRunning,
  status,
  phase,
  planSteps,
  notes: _notes,
  isApproving,
  onInterrupt,
  onApprove,
  onReject,
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

  const phaseChipStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '10px',
    fontWeight: 600,
    color: isRunning ? 'var(--accent)' : 'var(--text-muted)',
    background: isRunning ? 'var(--accent-dim)' : 'var(--bg-surface)',
    border: `1px solid ${isRunning ? 'rgba(167,139,250,0.3)' : 'var(--border-subtle)'}`,
    borderRadius: '999px',
    padding: '2px 8px',
    textTransform: 'uppercase',
    letterSpacing: '0.06em',
    marginRight: '8px',
    flexShrink: 0,
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

  const showApproval = status === 'awaiting_approval' && planSteps.length > 0
  const phaseLabel = phase ? (phaseLabels[phase] ?? phase) : (status ?? '')

  return (
    <div style={panelStyle} aria-hidden={!isOpen}>
      <div style={headerStyle}>
        <span style={titleStyle}>{taskTitle || 'Session'}</span>
        {phaseLabel && <span style={phaseChipStyle}>{phaseLabel}</span>}
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

      {showApproval && (
        <PlanApprovalPanel
          planSteps={planSteps}
          isApproving={isApproving}
          onApprove={onApprove}
          onReject={onReject}
        />
      )}

      {isRunning && !showApproval && (
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
