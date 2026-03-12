import { useState, type CSSProperties } from 'react'
import { useOutletContext } from 'react-router-dom'
import type { Task, AutonomyLevel } from '../types'
import { KanbanBoard, SessionStream } from '../components/forge'
import { Modal, Input, Button } from '../components/ui'
import { useTasks } from '../hooks/useTasks'
import { useStream } from '../hooks/useStream'
import type { AppOutletContext } from '../components/forge/AppShell'

const autonomyOptions: { value: AutonomyLevel; label: string }[] = [
  { value: 'supervised', label: 'Supervised' },
  { value: 'checkpoint', label: 'Checkpoint' },
  { value: 'autonomous', label: 'Autonomous' },
]

export function DashboardPage() {
  const { currentProjectId, isNewTaskOpen, setIsNewTaskOpen } =
    useOutletContext<AppOutletContext>()

  const { tasksByStatus, isLoading, createTask } = useTasks(currentProjectId)

  const [activeTask, setActiveTask] = useState<Task | null>(null)
  const [activeSessionId] = useState<string | null>(null)
  const [isPanelOpen, setIsPanelOpen] = useState(false)

  const [newTaskTitle, setNewTaskTitle] = useState('')
  const [newTaskAutonomy, setNewTaskAutonomy] = useState<AutonomyLevel>('supervised')
  const [isCreating, setIsCreating] = useState(false)

  const { lines, isRunning } = useStream(activeSessionId)

  function handleTaskClick(task: Task) {
    setActiveTask(task)
    setIsPanelOpen(true)
  }

  function handleCloseModal() {
    setIsNewTaskOpen(false)
    setNewTaskTitle('')
    setNewTaskAutonomy('supervised')
  }

  async function handleCreateTask() {
    if (!newTaskTitle.trim()) return
    setIsCreating(true)
    await createTask({ title: newTaskTitle.trim(), autonomy_level: newTaskAutonomy })
    setIsCreating(false)
    handleCloseModal()
  }

  const autonomySelectorStyle: CSSProperties = {
    display: 'flex',
    gap: '8px',
    marginTop: '4px',
  }

  const autonomyOptionStyle = (selected: boolean): CSSProperties => ({
    flex: 1,
    padding: '6px 0',
    borderRadius: 'var(--radius-md)',
    fontFamily: 'var(--font-ui)',
    fontSize: '12px',
    fontWeight: 500,
    textAlign: 'center',
    cursor: 'pointer',
    border: selected ? '1px solid rgba(167, 139, 250, 0.40)' : '1px solid var(--border-subtle)',
    background: selected ? 'var(--accent-dim)' : 'var(--bg-surface)',
    color: selected ? 'var(--accent)' : 'var(--text-secondary)',
    transition: 'all var(--duration-default) var(--ease-default)',
  })

  const labelStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '12px',
    color: 'var(--text-muted)',
    marginBottom: '6px',
    display: 'block',
  }

  return (
    <>
      <KanbanBoard
        tasksByStatus={tasksByStatus}
        activeTaskId={activeTask?.id ?? null}
        onTaskClick={handleTaskClick}
        isLoading={isLoading}
      />

      <SessionStream
        sessionId={activeSessionId}
        isOpen={isPanelOpen}
        onClose={() => setIsPanelOpen(false)}
        taskTitle={activeTask?.title ?? ''}
        lines={lines}
        isRunning={isRunning}
        onInterrupt={() => {}}
      />

      <Modal isOpen={isNewTaskOpen} onClose={handleCloseModal} title="New Task">
        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          <Input
            label="Task title"
            value={newTaskTitle}
            onChange={setNewTaskTitle}
            placeholder="Describe what needs to be done…"
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleCreateTask()
            }}
          />
          <div>
            <span style={labelStyle}>Autonomy level</span>
            <div style={autonomySelectorStyle}>
              {autonomyOptions.map((opt) => (
                <div
                  key={opt.value}
                  style={autonomyOptionStyle(opt.value === newTaskAutonomy)}
                  onClick={() => setNewTaskAutonomy(opt.value)}
                >
                  {opt.label}
                </div>
              ))}
            </div>
          </div>
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
            <Button variant="ghost" size="md" onClick={handleCloseModal}>
              Cancel
            </Button>
            <Button
              variant="primary"
              size="md"
              onClick={handleCreateTask}
              disabled={isCreating || !newTaskTitle.trim()}
            >
              {isCreating ? 'Creating…' : 'Create Task'}
            </Button>
          </div>
        </div>
      </Modal>
    </>
  )
}

export default DashboardPage
