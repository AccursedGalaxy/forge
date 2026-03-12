import { type CSSProperties, type ReactNode, useEffect } from 'react'
import { createPortal } from 'react-dom'

interface ModalProps {
  isOpen: boolean
  onClose: () => void
  title?: string
  children: ReactNode
  maxWidth?: string
}

export function Modal({ isOpen, onClose, title, children, maxWidth = '480px' }: ModalProps) {
  // Close on Escape key
  useEffect(() => {
    if (!isOpen) return

    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') {
        onClose()
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, onClose])

  // Prevent body scroll when modal is open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
    return () => {
      document.body.style.overflow = ''
    }
  }, [isOpen])

  if (!isOpen) return null

  const overlayStyle: CSSProperties = {
    position: 'fixed',
    inset: 0,
    background: 'rgba(0, 0, 0, 0.6)',
    backdropFilter: 'blur(4px)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    zIndex: 100,
    padding: '24px',
  }

  const panelStyle: CSSProperties = {
    background: 'var(--bg-elevated)',
    border: '1px solid var(--border-default)',
    borderRadius: 'var(--radius-lg)',
    boxShadow: 'var(--shadow-xl)',
    padding: '24px',
    width: '100%',
    maxWidth,
    maxHeight: 'calc(100vh - 48px)',
    overflowY: 'auto',
    position: 'relative',
  }

  const headerStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: title ? '20px' : 0,
  }

  const titleStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '15px',
    fontWeight: 600,
    color: 'var(--text-primary)',
    lineHeight: 1.3,
  }

  const closeBtnStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    width: '28px',
    height: '28px',
    background: 'transparent',
    border: '1px solid transparent',
    borderRadius: 'var(--radius-md)',
    color: 'var(--text-muted)',
    cursor: 'pointer',
    fontSize: '16px',
    lineHeight: 1,
    transition: 'all var(--duration-default) var(--ease-default)',
    flexShrink: 0,
  }

  function handleOverlayClick(e: React.MouseEvent<HTMLDivElement>) {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  return createPortal(
    <div style={overlayStyle} onClick={handleOverlayClick} role="dialog" aria-modal="true">
      <div style={panelStyle}>
        {(title) && (
          <div style={headerStyle}>
            {title && <h2 style={titleStyle}>{title}</h2>}
            <button
              style={closeBtnStyle}
              onClick={onClose}
              aria-label="Close modal"
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'var(--bg-overlay)'
                e.currentTarget.style.color = 'var(--text-primary)'
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'transparent'
                e.currentTarget.style.color = 'var(--text-muted)'
              }}
            >
              ×
            </button>
          </div>
        )}
        {!title && (
          <button
            style={{ ...closeBtnStyle, position: 'absolute', top: '16px', right: '16px' }}
            onClick={onClose}
            aria-label="Close modal"
            onMouseEnter={(e) => {
              e.currentTarget.style.background = 'var(--bg-overlay)'
              e.currentTarget.style.color = 'var(--text-primary)'
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = 'transparent'
              e.currentTarget.style.color = 'var(--text-muted)'
            }}
          >
            ×
          </button>
        )}
        {children}
      </div>
    </div>,
    document.body
  )
}

export default Modal
