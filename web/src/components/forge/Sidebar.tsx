import type { CSSProperties } from 'react'
import { NavLink } from 'react-router-dom'
import type { Project } from '../../types'
import { ProjectSwitcher } from './ProjectSwitcher'

interface NavItem {
  label: string
  path: string
}

const navItems: NavItem[] = [
  { label: 'Board', path: '/dashboard' },
  { label: 'Sessions', path: '/dashboard/sessions' },
  { label: 'Context', path: '/dashboard/context' },
  { label: 'Logs', path: '/dashboard/logs' },
  { label: 'Settings', path: '/dashboard/settings' },
]

interface SidebarProps {
  projects: Project[]
  currentProjectId: string | null
  onSelectProject: (id: string) => void
}

export function Sidebar({ projects, currentProjectId, onSelectProject }: SidebarProps) {
  const sidebarStyle: CSSProperties = {
    position: 'fixed',
    top: 0,
    left: 0,
    width: 'var(--sidebar-width)',
    height: '100vh',
    background: 'var(--bg-base)',
    borderRight: '1px solid var(--border-subtle)',
    display: 'flex',
    flexDirection: 'column',
    zIndex: 50,
    overflowY: 'auto',
  }

  const logoStyle: CSSProperties = {
    fontFamily: 'var(--font-mono)',
    fontSize: '14px',
    fontWeight: 700,
    color: 'var(--accent)',
    letterSpacing: '0.08em',
    padding: '16px 16px 0',
    display: 'block',
    lineHeight: 1,
  }

  const switcherWrapStyle: CSSProperties = {
    margin: '12px 8px',
  }

  const navSectionStyle: CSSProperties = {
    flex: 1,
    padding: '4px 8px',
    display: 'flex',
    flexDirection: 'column',
    gap: '2px',
  }

  const footerStyle: CSSProperties = {
    padding: '12px 8px',
    borderTop: '1px solid var(--border-subtle)',
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
    cursor: 'pointer',
  }

  const avatarStyle: CSSProperties = {
    width: '28px',
    height: '28px',
    borderRadius: 'var(--radius-full)',
    background: 'var(--accent-dim)',
    border: '1px solid rgba(167, 139, 250, 0.30)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: '11px',
    fontWeight: 600,
    color: 'var(--accent)',
    fontFamily: 'var(--font-ui)',
    flexShrink: 0,
  }

  const userNameStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
    fontWeight: 500,
    color: 'var(--text-secondary)',
    lineHeight: 1,
  }

  return (
    <aside style={sidebarStyle}>
      <span style={logoStyle}>FORGE</span>

      <div style={switcherWrapStyle}>
        <ProjectSwitcher
          projects={projects}
          currentProjectId={currentProjectId}
          onSelect={onSelectProject}
        />
      </div>

      <nav style={navSectionStyle}>
        {navItems.map((item) => (
          <NavLink
            key={item.path}
            to={item.path}
            end={item.path === '/dashboard'}
            style={({ isActive }) => ({
              display: 'flex',
              alignItems: 'center',
              height: '32px',
              padding: '0 12px',
              borderRadius: 'var(--radius-md)',
              fontFamily: 'var(--font-ui)',
              fontSize: '13px',
              fontWeight: 500,
              textDecoration: 'none',
              color: isActive ? 'var(--accent)' : 'var(--text-secondary)',
              background: isActive ? 'var(--accent-dim)' : 'transparent',
              transition: 'all var(--duration-default) var(--ease-default)',
            })}
            onMouseEnter={(e) => {
              const el = e.currentTarget
              const isActive = el.getAttribute('aria-current') === 'page'
              if (!isActive) {
                el.style.background = 'var(--bg-elevated)'
                el.style.color = 'var(--text-primary)'
              }
            }}
            onMouseLeave={(e) => {
              const el = e.currentTarget
              const isActive = el.getAttribute('aria-current') === 'page'
              if (!isActive) {
                el.style.background = 'transparent'
                el.style.color = 'var(--text-secondary)'
              }
            }}
          >
            {item.label}
          </NavLink>
        ))}
      </nav>

      <div style={footerStyle}>
        <div style={avatarStyle}>A</div>
        <span style={userNameStyle}>Aki</span>
      </div>
    </aside>
  )
}

export default Sidebar
