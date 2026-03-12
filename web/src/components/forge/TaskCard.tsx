import { useState, type CSSProperties } from 'react'
import type { Task } from '../../types'
import { Badge } from '../ui'

interface TaskCardProps {
  task: Task
  isActive: boolean
  onClick: () => void
}

const pulseKeyframes = `
@keyframes forge-pulse {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.4); opacity: 0.6; }
}
`

// Inject keyframes once
if (typeof document !== 'undefined' && !document.getElementById('forge-task-card-styles')) {
  const style = document.createElement('style')
  style.id = 'forge-task-card-styles'
  style.textContent = pulseKeyframes
  document.head.appendChild(style)
}

export function TaskCard({ task, isActive, onClick }: TaskCardProps) {
  const [isHovered, setIsHovered] = useState(false)

  const cardStyle: CSSProperties = {
    background: 'var(--bg-surface)',
    border: `1px solid ${isHovered || isActive ? 'var(--border-default)' : 'var(--border-subtle)'}`,
    borderRadius: 'var(--radius-md)',
    padding: '10px 12px',
    cursor: 'pointer',
    boxShadow: isHovered ? 'var(--shadow-md)' : 'var(--shadow-sm)',
    transition: 'all var(--duration-default) var(--ease-default)',
  }

  const titleStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
    fontWeight: 500,
    color: 'var(--text-primary)',
    lineHeight: 1.4,
    marginBottom: task.description ? '4px' : '8px',
  }

  const descriptionStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '12px',
    color: 'var(--text-muted)',
    lineHeight: 1.5,
    marginBottom: '8px',
    overflow: 'hidden',
    display: '-webkit-box',
    WebkitLineClamp: 2,
    WebkitBoxOrient: 'vertical',
  }

  const footerStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
  }

  const activeDotStyle: CSSProperties = {
    width: '6px',
    height: '6px',
    borderRadius: '50%',
    background: 'var(--accent)',
    animation: 'forge-pulse 1.4s ease-in-out infinite',
    flexShrink: 0,
  }

  return (
    <div
      style={cardStyle}
      onClick={onClick}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <div style={titleStyle}>{task.title}</div>
      {task.description && <div style={descriptionStyle}>{task.description}</div>}
      <div style={footerStyle}>
        <Badge variant="autonomy" autonomy={task.autonomy_level} />
        {isActive && <div style={activeDotStyle} />}
      </div>
    </div>
  )
}

export default TaskCard
