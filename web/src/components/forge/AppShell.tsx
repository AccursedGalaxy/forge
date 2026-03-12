import { useState, useEffect, useCallback, type CSSProperties } from 'react'
import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { TopBar } from './TopBar'
import { LogViewer } from './LogViewer'
import { useProjects } from '../../hooks/useProjects'

export interface AppOutletContext {
  currentProjectId: string | null
  setCurrentProjectId: (id: string | null) => void
  isNewTaskOpen: boolean
  setIsNewTaskOpen: (open: boolean) => void
}

const LOG_DRAWER_HEIGHT = 340

export function AppShell() {
  const { projects } = useProjects()
  const [currentProjectId, setCurrentProjectId] = useState<string | null>(null)
  const [isNewTaskOpen, setIsNewTaskOpen] = useState(false)
  const [isLogOpen, setIsLogOpen] = useState(false)

  // Auto-select first project when projects load
  useEffect(() => {
    if (projects.length > 0 && !currentProjectId) {
      setCurrentProjectId(projects[0].id)
    }
  }, [projects, currentProjectId])

  // Ctrl+Shift+L toggles the log drawer from anywhere in the app
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (e.ctrlKey && e.shiftKey && e.key === 'L') {
      e.preventDefault()
      setIsLogOpen((v) => !v)
    }
  }, [])

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  const mainStyle: CSSProperties = {
    marginLeft: 'var(--sidebar-width)',
    minHeight: '100vh',
    display: 'flex',
    flexDirection: 'column',
    background: 'var(--bg-base)',
    paddingBottom: isLogOpen ? LOG_DRAWER_HEIGHT : 0,
    transition: 'padding-bottom var(--duration-default) var(--ease-default)',
  }

  const contentStyle: CSSProperties = {
    flex: 1,
    padding: '24px',
    maxWidth: 'var(--content-max-width)',
    width: '100%',
  }

  const drawerStyle: CSSProperties = {
    position: 'fixed',
    bottom: 0,
    left: 'var(--sidebar-width)',
    right: 0,
    height: LOG_DRAWER_HEIGHT,
    background: 'var(--bg-base)',
    borderTop: '1px solid var(--border-default)',
    display: 'flex',
    flexDirection: 'column',
    zIndex: 40,
    transform: isLogOpen ? 'translateY(0)' : `translateY(${LOG_DRAWER_HEIGHT}px)`,
    transition: 'transform var(--duration-default) var(--ease-default)',
  }

  const drawerHeaderStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '6px 16px',
    borderBottom: '1px solid var(--border-subtle)',
    flexShrink: 0,
  }

  const drawerTitleStyle: CSSProperties = {
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    fontWeight: 600,
    color: 'var(--text-muted)',
    letterSpacing: '0.06em',
    textTransform: 'uppercase',
  }

  const closeBtnStyle: CSSProperties = {
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    color: 'var(--text-muted)',
    background: 'transparent',
    border: 'none',
    cursor: 'pointer',
    padding: '2px 6px',
    borderRadius: 'var(--radius-sm)',
  }

  const drawerContentStyle: CSSProperties = {
    flex: 1,
    overflow: 'hidden',
    padding: '8px 16px 12px',
  }

  const outletContext: AppOutletContext = {
    currentProjectId,
    setCurrentProjectId,
    isNewTaskOpen,
    setIsNewTaskOpen,
  }

  return (
    <>
      <Sidebar
        projects={projects}
        currentProjectId={currentProjectId}
        onSelectProject={setCurrentProjectId}
      />
      <main style={mainStyle}>
        <TopBar onNewTask={() => setIsNewTaskOpen(true)} />
        <div style={contentStyle}>
          <Outlet context={outletContext} />
        </div>
      </main>

      {/* Log drawer — always mounted so the SSE connection persists; slide in/out */}
      <div style={drawerStyle} aria-hidden={!isLogOpen}>
        <div style={drawerHeaderStyle}>
          <span style={drawerTitleStyle}>Logs</span>
          <button
            style={closeBtnStyle}
            onClick={() => setIsLogOpen(false)}
            title="Close (Ctrl+Shift+L)"
          >
            ✕
          </button>
        </div>
        <div style={drawerContentStyle}>
          {/* LogViewer is always mounted — connection stays alive, history accumulates */}
          <LogViewer compact />
        </div>
      </div>
    </>
  )
}

export default AppShell
