import { type CSSProperties } from 'react'
import type { PlanStep } from '../../types'
import { Button } from '../ui'

interface PlanApprovalPanelProps {
  planSteps: PlanStep[]
  isApproving: boolean
  onApprove: () => void
  onReject: () => void
}

export function PlanApprovalPanel({
  planSteps,
  isApproving,
  onApprove,
  onReject,
}: PlanApprovalPanelProps) {
  const containerStyle: CSSProperties = {
    padding: '16px',
    borderTop: '1px solid var(--border-subtle)',
    background: 'var(--bg-surface)',
    flexShrink: 0,
  }

  const titleStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '12px',
    fontWeight: 600,
    color: 'var(--text-muted)',
    textTransform: 'uppercase',
    letterSpacing: '0.05em',
    marginBottom: '12px',
  }

  const stepsListStyle: CSSProperties = {
    listStyle: 'none',
    padding: 0,
    margin: '0 0 14px 0',
    display: 'flex',
    flexDirection: 'column',
    gap: '8px',
  }

  const stepStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'flex-start',
    gap: '10px',
    fontFamily: 'var(--font-ui)',
    fontSize: '12px',
    color: 'var(--text-secondary)',
    lineHeight: 1.5,
  }

  const stepNumStyle: CSSProperties = {
    flexShrink: 0,
    width: '18px',
    height: '18px',
    borderRadius: '50%',
    background: 'var(--accent-dim)',
    border: '1px solid rgba(167, 139, 250, 0.3)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: '10px',
    fontWeight: 600,
    color: 'var(--accent)',
    marginTop: '1px',
  }

  const actionsStyle: CSSProperties = {
    display: 'flex',
    gap: '8px',
  }

  if (planSteps.length === 0) return null

  return (
    <div style={containerStyle}>
      <div style={titleStyle}>Proposed Plan</div>
      <ol style={stepsListStyle}>
        {planSteps.map((step) => (
          <li key={step.index} style={stepStyle}>
            <span style={stepNumStyle}>{step.index + 1}</span>
            <span>{step.description}</span>
          </li>
        ))}
      </ol>
      <div style={actionsStyle}>
        <Button
          variant="primary"
          size="sm"
          onClick={onApprove}
          disabled={isApproving}
        >
          {isApproving ? 'Approving…' : 'Approve & Execute'}
        </Button>
        <Button variant="ghost" size="sm" onClick={onReject} disabled={isApproving}>
          Reject
        </Button>
      </div>
    </div>
  )
}

export default PlanApprovalPanel
