import { useState, type CSSProperties } from 'react'
import { useOutletContext } from 'react-router-dom'
import type { ContextChunk } from '../types'
import { useContext } from '../hooks/useContext'
import { Skeleton, Button, Modal } from '../components/ui'
import type { AppOutletContext } from '../components/forge/AppShell'

const chunkTypeLabels: Record<ContextChunk['chunk_type'], string> = {
  session_output: 'Session Output',
  code_diff: 'Code Diff',
  task_note: 'Task Note',
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

export function ContextPage() {
  const { currentProjectId } = useOutletContext<AppOutletContext>()
  const { chunks, isLoading, deleteChunk } = useContext(currentProjectId)

  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)

  async function handleConfirmDelete() {
    if (!confirmDeleteId) return
    setIsDeleting(true)
    await deleteChunk(confirmDeleteId)
    setIsDeleting(false)
    setConfirmDeleteId(null)
  }

  const headerStyle: CSSProperties = {
    marginBottom: '20px',
  }

  const titleStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '16px',
    fontWeight: 600,
    color: 'var(--text-primary)',
    marginBottom: '4px',
  }

  const subtitleStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
    color: 'var(--text-muted)',
  }

  const listStyle: CSSProperties = {
    display: 'flex',
    flexDirection: 'column',
    gap: '8px',
  }

  const rowStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'flex-start',
    gap: '12px',
    padding: '12px 14px',
    background: 'var(--bg-surface)',
    border: '1px solid var(--border-subtle)',
    borderRadius: 'var(--radius-md)',
  }

  const rowInfoStyle: CSSProperties = {
    flex: 1,
    minWidth: 0,
  }

  const rowHeaderStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    marginBottom: '4px',
  }

  const previewStyle: CSSProperties = {
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    color: 'var(--text-secondary)',
    lineHeight: 1.5,
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    display: '-webkit-box',
    WebkitLineClamp: 2,
    WebkitBoxOrient: 'vertical',
  }

  const dateStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '11px',
    color: 'var(--text-muted)',
  }

  const emptyStyle: CSSProperties = {
    textAlign: 'center',
    padding: '48px 24px',
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
    color: 'var(--text-muted)',
    border: '1px dashed var(--border-subtle)',
    borderRadius: 'var(--radius-lg)',
  }

  return (
    <div>
      <div style={headerStyle}>
        <div style={titleStyle}>Context Memory</div>
        <div style={subtitleStyle}>
          Chunks of context retrieved and stored from agent sessions.
        </div>
      </div>

      {isLoading ? (
        <div style={listStyle}>
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} width="100%" height="72px" />
          ))}
        </div>
      ) : chunks.length === 0 ? (
        <div style={emptyStyle}>
          No context chunks yet. Run a session to start building memory.
        </div>
      ) : (
        <div style={listStyle}>
          {chunks.map((chunk) => (
            <div key={chunk.id} style={rowStyle}>
              <div style={rowInfoStyle}>
                <div style={rowHeaderStyle}>
                  <span
                    style={{
                      fontFamily: 'var(--font-ui)',
                      fontSize: '11px',
                      fontWeight: 500,
                      padding: '2px 8px',
                      borderRadius: 'var(--radius-full)',
                      background: 'var(--bg-elevated)',
                      color: 'var(--text-secondary)',
                      border: '1px solid var(--border-subtle)',
                    }}
                  >
                    {chunkTypeLabels[chunk.chunk_type]}
                  </span>
                  <span style={dateStyle}>{formatDate(chunk.created_at)}</span>
                </div>
                <div style={previewStyle}>
                  {chunk.content.slice(0, 120)}
                  {chunk.content.length > 120 ? '…' : ''}
                </div>
              </div>
              <Button
                variant="danger"
                size="sm"
                onClick={() => setConfirmDeleteId(chunk.id)}
                style={{ flexShrink: 0 }}
              >
                Delete
              </Button>
            </div>
          ))}
        </div>
      )}

      <Modal
        isOpen={!!confirmDeleteId}
        onClose={() => setConfirmDeleteId(null)}
        title="Delete Context Chunk"
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          <p
            style={{
              fontFamily: 'var(--font-ui)',
              fontSize: '13px',
              color: 'var(--text-secondary)',
              margin: 0,
            }}
          >
            Are you sure you want to delete this context chunk? This cannot be undone.
          </p>
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
            <Button variant="ghost" size="md" onClick={() => setConfirmDeleteId(null)}>
              Cancel
            </Button>
            <Button
              variant="danger"
              size="md"
              onClick={handleConfirmDelete}
              disabled={isDeleting}
            >
              {isDeleting ? 'Deleting…' : 'Delete'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}

export default ContextPage
