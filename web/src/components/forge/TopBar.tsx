import type { CSSProperties } from 'react'
import { useLocation } from 'react-router-dom'
import { Button } from '../ui'

const breadcrumbMap: Record<string, string> = {
  '/dashboard': 'Board',
  '/dashboard/sessions': 'Sessions',
  '/dashboard/context': 'Context',
  '/dashboard/settings': 'Settings',
}

interface TopBarProps {
  onNewTask: () => void
}

export function TopBar({ onNewTask }: TopBarProps) {
  const location = useLocation()
  const currentPage = breadcrumbMap[location.pathname] ?? 'Board'

  const topbarStyle: CSSProperties = {
    position: 'sticky',
    top: 0,
    height: 'var(--topbar-height)',
    background: 'var(--bg-base)',
    borderBottom: '1px solid var(--border-subtle)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '0 24px',
    zIndex: 40,
  }

  const breadcrumbStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: '6px',
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
    color: 'var(--text-muted)',
  }

  const breadcrumbActiveStyle: CSSProperties = {
    color: 'var(--text-primary)',
    fontWeight: 500,
  }

  const actionsStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
  }

  return (
    <header style={topbarStyle}>
      <div style={breadcrumbStyle}>
        <span>FORGE</span>
        <span style={{ color: 'var(--border-strong)' }}>/</span>
        <span style={breadcrumbActiveStyle}>{currentPage}</span>
      </div>

      <div style={actionsStyle}>
        <Button variant="ghost" size="md">
          Run Agent
        </Button>
        <Button variant="primary" size="md" onClick={onNewTask}>
          New Task
        </Button>
      </div>
    </header>
  )
}

export default TopBar
