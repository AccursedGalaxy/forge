import type { CSSProperties } from 'react'
import type { Task, TaskStatus } from '../../types'
import { Skeleton } from '../ui'
import KanbanColumn from './KanbanColumn'

const ALL_STATUSES: TaskStatus[] = ['backlog', 'planned', 'in_progress', 'review', 'done']

interface KanbanBoardProps {
  tasksByStatus: Record<TaskStatus, Task[]>
  activeTaskId: string | null
  onTaskClick: (task: Task) => void
  isLoading: boolean
}

function SkeletonColumn() {
  const columnStyle: CSSProperties = {
    width: '260px',
    minWidth: '260px',
    display: 'flex',
    flexDirection: 'column',
    gap: '8px',
  }

  const headerStyle: CSSProperties = {
    marginBottom: '2px',
  }

  return (
    <div style={columnStyle}>
      <div style={headerStyle}>
        <Skeleton width="100px" height="20px" />
      </div>
      <Skeleton width="100%" height="72px" />
      <Skeleton width="100%" height="72px" />
      <Skeleton width="100%" height="72px" />
    </div>
  )
}

export function KanbanBoard({ tasksByStatus, activeTaskId, onTaskClick, isLoading }: KanbanBoardProps) {
  const boardStyle: CSSProperties = {
    display: 'flex',
    gap: '16px',
    overflowX: 'auto',
    padding: '4px 2px 16px',
    alignItems: 'flex-start',
  }

  if (isLoading) {
    return (
      <div style={boardStyle}>
        {ALL_STATUSES.map((s) => (
          <SkeletonColumn key={s} />
        ))}
      </div>
    )
  }

  return (
    <div style={boardStyle}>
      {ALL_STATUSES.map((status) => (
        <KanbanColumn
          key={status}
          status={status}
          tasks={tasksByStatus[status]}
          activeTaskId={activeTaskId}
          onTaskClick={onTaskClick}
        />
      ))}
    </div>
  )
}

export default KanbanBoard
