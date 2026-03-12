import { useState, useRef, useEffect, type CSSProperties } from 'react'
import type { Project } from '../../types'

interface ProjectSwitcherProps {
  projects: Project[]
  currentProjectId: string | null
  onSelect: (id: string) => void
}

export function ProjectSwitcher({ projects, currentProjectId, onSelect }: ProjectSwitcherProps) {
  const [isOpen, setIsOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  const currentProject = projects.find((p) => p.id === currentProjectId)

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const triggerStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '8px 10px',
    background: 'var(--bg-surface)',
    border: '1px solid var(--border-subtle)',
    borderRadius: 'var(--radius-md)',
    cursor: 'pointer',
    transition: 'all var(--duration-default) var(--ease-default)',
    width: '100%',
  }

  const nameStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
    fontWeight: 500,
    color: 'var(--text-primary)',
    lineHeight: 1,
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  }

  const chevronStyle: CSSProperties = {
    color: 'var(--text-muted)',
    fontSize: '10px',
    lineHeight: 1,
    flexShrink: 0,
    marginLeft: '4px',
    transform: isOpen ? 'rotate(180deg)' : 'rotate(0deg)',
    transition: `transform var(--duration-fast)`,
  }

  const dropdownStyle: CSSProperties = {
    position: 'absolute',
    top: 'calc(100% + 4px)',
    left: 0,
    right: 0,
    background: 'var(--bg-elevated)',
    border: '1px solid var(--border-default)',
    borderRadius: 'var(--radius-md)',
    zIndex: 200,
    overflow: 'hidden',
    boxShadow: 'var(--shadow-lg)',
  }

  const itemStyle = (isSelected: boolean): CSSProperties => ({
    display: 'block',
    width: '100%',
    padding: '8px 10px',
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
    fontWeight: isSelected ? 500 : 400,
    color: isSelected ? 'var(--accent)' : 'var(--text-secondary)',
    background: isSelected ? 'var(--accent-dim)' : 'transparent',
    cursor: 'pointer',
    border: 'none',
    textAlign: 'left',
    transition: 'background var(--duration-fast)',
  })

  const emptyStyle: CSSProperties = {
    padding: '10px',
    fontFamily: 'var(--font-ui)',
    fontSize: '12px',
    color: 'var(--text-muted)',
  }

  return (
    <div ref={containerRef} style={{ position: 'relative' }}>
      <div style={triggerStyle} onClick={() => setIsOpen((o) => !o)}>
        <span style={nameStyle}>{currentProject?.name ?? 'No projects'}</span>
        <span style={chevronStyle}>▾</span>
      </div>

      {isOpen && (
        <div style={dropdownStyle}>
          {projects.length === 0 ? (
            <div style={emptyStyle}>No projects</div>
          ) : (
            projects.map((project) => (
              <button
                key={project.id}
                style={itemStyle(project.id === currentProjectId)}
                onClick={() => {
                  onSelect(project.id)
                  setIsOpen(false)
                }}
                onMouseEnter={(e) => {
                  if (project.id !== currentProjectId) {
                    ;(e.currentTarget as HTMLButtonElement).style.background = 'var(--bg-surface)'
                  }
                }}
                onMouseLeave={(e) => {
                  ;(e.currentTarget as HTMLButtonElement).style.background =
                    project.id === currentProjectId ? 'var(--accent-dim)' : 'transparent'
                }}
              >
                {project.name}
              </button>
            ))
          )}
        </div>
      )}
    </div>
  )
}

export default ProjectSwitcher
