import type { CSSProperties, ReactNode } from 'react'

export type BadgeVariant = 'status' | 'autonomy' | 'count'
export type BadgeStatus = 'backlog' | 'planned' | 'in_progress' | 'review' | 'done'
export type BadgeAutonomy = 'supervised' | 'checkpoint' | 'autonomous'

interface BadgeProps {
  variant?: BadgeVariant
  status?: BadgeStatus
  autonomy?: BadgeAutonomy
  count?: number
  children?: ReactNode
  className?: string
}

const statusStyles: Record<BadgeStatus, CSSProperties> = {
  backlog: {
    background: 'var(--status-backlog)',
    color: 'var(--status-backlog-text)',
  },
  planned: {
    background: 'var(--status-planned)',
    color: 'var(--status-planned-text)',
  },
  in_progress: {
    background: 'var(--status-in-progress)',
    color: 'var(--status-in-progress-text)',
  },
  review: {
    background: 'var(--status-review)',
    color: 'var(--status-review-text)',
  },
  done: {
    background: 'var(--status-done)',
    color: 'var(--status-done-text)',
  },
}

const statusLabels: Record<BadgeStatus, string> = {
  backlog: 'Backlog',
  planned: 'Planned',
  in_progress: 'In Progress',
  review: 'Review',
  done: 'Done',
}

const autonomyStyles: Record<BadgeAutonomy, CSSProperties> = {
  supervised: {
    background: 'var(--autonomy-supervised)',
    color: 'var(--autonomy-supervised-text)',
  },
  checkpoint: {
    background: 'var(--autonomy-checkpoint)',
    color: 'var(--autonomy-checkpoint-text)',
  },
  autonomous: {
    background: 'var(--autonomy-autonomous)',
    color: 'var(--autonomy-autonomous-text)',
  },
}

const autonomyLabels: Record<BadgeAutonomy, string> = {
  supervised: 'Supervised',
  checkpoint: 'Checkpoint',
  autonomous: 'Autonomous',
}

export function Badge({
  variant = 'status',
  status,
  autonomy,
  count,
  children,
  className,
}: BadgeProps) {
  if (variant === 'count') {
    const style: CSSProperties = {
      display: 'inline-flex',
      alignItems: 'center',
      justifyContent: 'center',
      width: '16px',
      height: '16px',
      borderRadius: 'var(--radius-full)',
      background: 'var(--accent-dim)',
      color: 'var(--accent)',
      fontFamily: 'var(--font-ui)',
      fontSize: '10px',
      fontWeight: 600,
      lineHeight: 1,
    }
    return (
      <span style={style} className={className}>
        {count ?? children}
      </span>
    )
  }

  if (variant === 'autonomy' && autonomy) {
    const style: CSSProperties = {
      display: 'inline-flex',
      alignItems: 'center',
      padding: '2px 8px',
      borderRadius: 'var(--radius-full)',
      fontFamily: 'var(--font-ui)',
      fontSize: '11px',
      fontWeight: 500,
      letterSpacing: '0.05em',
      textTransform: 'uppercase',
      lineHeight: 1.4,
      ...autonomyStyles[autonomy],
    }
    return (
      <span style={style} className={className}>
        {children ?? autonomyLabels[autonomy]}
      </span>
    )
  }

  // Default: status badge
  if (status) {
    const style: CSSProperties = {
      display: 'inline-flex',
      alignItems: 'center',
      padding: '2px 8px',
      borderRadius: 'var(--radius-full)',
      fontFamily: 'var(--font-ui)',
      fontSize: '11px',
      fontWeight: 500,
      letterSpacing: '0.05em',
      textTransform: 'uppercase',
      lineHeight: 1.4,
      ...statusStyles[status],
    }
    return (
      <span style={style} className={className}>
        {children ?? statusLabels[status]}
      </span>
    )
  }

  return (
    <span className={className}>
      {children}
    </span>
  )
}

export default Badge
