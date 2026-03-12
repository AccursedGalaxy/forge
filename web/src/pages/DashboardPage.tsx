import { useState, useCallback, type CSSProperties } from 'react'
import { useOutletContext } from 'react-router-dom'
import type { Task, AutonomyLevel, Session } from '../types'
import { KanbanBoard, SessionStream } from '../components/forge'
import { Modal, Input, Button } from '../components/ui'
import { useTasks } from '../hooks/useTasks'
import { useStream } from '../hooks/useStream'
import * as api from '../lib/api'
import log from '../lib/logger'
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
  const [activeSession, setActiveSession] = useState<Session | null>(null)
  const [isPanelOpen, setIsPanelOpen] = useState(false)
  const [isApproving, setIsApproving] = useState(false)

  const [newTaskTitle, setNewTaskTitle] = useState('')
  const [newTaskAutonomy, setNewTaskAutonomy] = useState<AutonomyLevel>('supervised')
  const [isCreating, setIsCreating] = useState(false)

  const activeSessionId = activeSession?.id ?? null
  const { lines, isRunning, status, phase, planSteps, notes } = useStream(activeSessionId)

  async function handleTaskClick(task: Task) {
    setActiveTask(task)
    setIsPanelOpen(true)

    // Load latest session for this task, or start a new plan session
    try {
      const sessions = await api.getSessions(task.id)
      if (sessions.length > 0) {
        // Use the most recent session
        setActiveSession(sessions[0])
      } else if (currentProjectId) {
        // Start a new planning session
        const session = await api.createSession({
          task_id: task.id,
          project_id: currentProjectId,
          session_type: 'plan',
        })
        setActiveSession(session)
      }
    } catch (err) {
      log.error('DashboardPage', 'Failed to load/create session', err)
    }
  }

  const handleApprove = useCallback(async () => {
    if (!activeSessionId) return
    setIsApproving(true)
    try {
      const updated = await api.approveSession(activeSessionId)
      setActiveSession(updated)
    } catch (err) {
      log.error('DashboardPage', 'Failed to approve session', err)
    } finally {
      setIsApproving(false)
    }
  }, [activeSessionId])

  const handleReject = useCallback(() => {
    // Close the panel — user can create a new session later
    setIsPanelOpen(false)
    setActiveSession(null)
  }, [])

  const handleInterrupt = useCallback(async () => {
    if (!activeSessionId) return
    try {
      await api.interruptSession(activeSessionId)
    } catch (err) {
      log.error('DashboardPage', 'Failed to interrupt session', err)
    }
  }, [activeSessionId])

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

  // Derive the effective status from SSE updates or fall back to session DB status
  const effectiveStatus = status ?? activeSession?.status ?? null
  // Use planSteps from SSE stream, or fall back to session's stored plan_steps
  const effectivePlanSteps =
    planSteps.length > 0
      ? planSteps
      : (activeSession?.plan_steps ?? [])

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
        onClose={() => {
          setIsPanelOpen(false)
        }}
        taskTitle={activeTask?.title ?? ''}
        lines={lines}
        isRunning={isRunning}
        status={effectiveStatus}
        phase={phase}
        planSteps={effectivePlanSteps}
        notes={notes}
        isApproving={isApproving}
        onInterrupt={handleInterrupt}
        onApprove={handleApprove}
        onReject={handleReject}
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
