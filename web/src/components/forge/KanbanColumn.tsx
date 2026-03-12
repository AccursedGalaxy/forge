import type { CSSProperties } from 'react'
import type { Task, TaskStatus } from '../../types'
import { Badge } from '../ui'
import TaskCard from './TaskCard'

interface KanbanColumnProps {
  status: TaskStatus
  tasks: Task[]
  activeTaskId: string | null
  onTaskClick: (task: Task) => void
}

const columnLabels: Record<TaskStatus, string> = {
  backlog: 'Backlog',
  planned: 'Planned',
  in_progress: 'In Progress',
  review: 'Review',
  done: 'Done',
}

export function KanbanColumn({ status, tasks, activeTaskId, onTaskClick }: KanbanColumnProps) {
  const columnStyle: CSSProperties = {
    width: '260px',
    minWidth: '260px',
    display: 'flex',
    flexDirection: 'column',
    gap: '0',
  }

  const headerStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    marginBottom: '10px',
  }

  const labelStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '12px',
    fontWeight: 600,
    color: 'var(--text-secondary)',
    textTransform: 'uppercase',
    letterSpacing: '0.08em',
  }

  const countStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '12px',
    color: 'var(--text-muted)',
  }

  const cardsStyle: CSSProperties = {
    display: 'flex',
    flexDirection: 'column',
    gap: '8px',
    flex: 1,
  }

  const emptyStyle: CSSProperties = {
    border: '1px dashed var(--border-subtle)',
    borderRadius: 'var(--radius-md)',
    padding: '20px 12px',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontFamily: 'var(--font-ui)',
    fontSize: '12px',
    color: 'var(--text-muted)',
  }

  return (
    <div style={columnStyle}>
      <div style={headerStyle}>
        <Badge variant="status" status={status} />
        <span style={labelStyle}>{columnLabels[status]}</span>
        <span style={countStyle}>{tasks.length}</span>
      </div>
      <div style={cardsStyle}>
        {tasks.length === 0 ? (
          <div style={emptyStyle}>No tasks</div>
        ) : (
          tasks.map((task) => (
            <TaskCard
              key={task.id}
              task={task}
              isActive={task.id === activeTaskId}
              onClick={() => onTaskClick(task)}
            />
          ))
        )}
      </div>
    </div>
  )
}

export default KanbanColumn
