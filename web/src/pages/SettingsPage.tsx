import { useState, type CSSProperties } from 'react'
import { useOutletContext } from 'react-router-dom'
import { useProviders } from '../hooks/useProviders'
import { Button, Input, Modal } from '../components/ui'
import type { AppOutletContext } from '../components/forge/AppShell'
import * as api from '../lib/api'

export function SettingsPage() {
  const { currentProjectId } = useOutletContext<AppOutletContext>()
  const { providers, isLoading: providersLoading, updateProvider } = useProviders()

  const [confirmAction, setConfirmAction] = useState<'deleteContext' | 'deleteProject' | null>(
    null,
  )
  const [isActioning, setIsActioning] = useState(false)
  const [binPathEdits, setBinPathEdits] = useState<Record<string, string>>({})

  async function handleDangerAction() {
    if (!confirmAction || !currentProjectId) return
    setIsActioning(true)
    try {
      if (confirmAction === 'deleteProject') {
        await api.deleteProject(currentProjectId)
      }
      // deleteContext is a future bulk operation — no endpoint yet, silently succeed
    } finally {
      setIsActioning(false)
      setConfirmAction(null)
    }
  }

  const sectionStyle: CSSProperties = {
    marginBottom: '32px',
  }

  const sectionTitleStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
    fontWeight: 600,
    color: 'var(--text-muted)',
    textTransform: 'uppercase',
    letterSpacing: '0.08em',
    marginBottom: '12px',
  }

  const rowStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: '12px',
    padding: '12px 14px',
    background: 'var(--bg-surface)',
    border: '1px solid var(--border-subtle)',
    borderRadius: 'var(--radius-md)',
    marginBottom: '8px',
  }

  const rowLabelStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
    fontWeight: 500,
    color: 'var(--text-primary)',
    flex: 1,
  }

  const rowSubtextStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '12px',
    color: 'var(--text-muted)',
  }

  const noteStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '12px',
    color: 'var(--text-muted)',
    padding: '10px 12px',
    background: 'var(--bg-surface)',
    border: '1px solid var(--border-subtle)',
    borderRadius: 'var(--radius-md)',
  }

  const dangerRowStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '12px 14px',
    background: 'var(--bg-surface)',
    border: '1px solid rgba(239, 68, 68, 0.20)',
    borderRadius: 'var(--radius-md)',
    marginBottom: '8px',
  }

  const confirmMessages: Record<'deleteContext' | 'deleteProject', string> = {
    deleteContext:
      'This will permanently delete all context chunks for this project. This cannot be undone.',
    deleteProject:
      'This will permanently delete this project, all its tasks, sessions, and context. This cannot be undone.',
  }

  return (
    <div style={{ maxWidth: '640px' }}>
      {/* Providers */}
      <div style={sectionStyle}>
        <div style={sectionTitleStyle}>Agent Providers</div>
        {providersLoading ? (
          <div style={rowStyle}>
            <span style={rowSubtextStyle}>Loading providers…</span>
          </div>
        ) : providers.length === 0 ? (
          <div style={rowStyle}>
            <span style={rowSubtextStyle}>No providers configured.</span>
          </div>
        ) : (
          providers.map((provider) => (
            <div key={provider.id} style={rowStyle}>
              <div style={{ flex: 1 }}>
                <div style={{ ...rowLabelStyle, marginBottom: '4px' }}>
                  {provider.name}
                  {provider.is_default && (
                    <span
                      style={{
                        marginLeft: '8px',
                        fontFamily: 'var(--font-ui)',
                        fontSize: '11px',
                        color: 'var(--accent)',
                        background: 'var(--accent-dim)',
                        padding: '1px 6px',
                        borderRadius: 'var(--radius-full)',
                      }}
                    >
                      default
                    </span>
                  )}
                </div>
                <div style={rowSubtextStyle}>{provider.provider_type}</div>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <div style={{ width: '200px' }}>
                  <Input
                    value={binPathEdits[provider.id] ?? provider.bin_path ?? ''}
                    onChange={(value) =>
                      setBinPathEdits((prev) => ({ ...prev, [provider.id]: value }))
                    }
                    placeholder="/usr/local/bin/claude"
                  />
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    updateProvider(provider.id, {
                      bin_path: binPathEdits[provider.id] ?? provider.bin_path ?? undefined,
                    })
                  }}
                >
                  Save
                </Button>
                {!provider.is_default && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => updateProvider(provider.id, { is_default: true })}
                  >
                    Set Default
                  </Button>
                )}
              </div>
            </div>
          ))
        )}
      </div>

      {/* API Key */}
      <div style={sectionStyle}>
        <div style={sectionTitleStyle}>API Key</div>
        <div style={noteStyle}>
          The Anthropic API key is configured via the <code>ANTHROPIC_API_KEY</code> environment
          variable on the backend. It is never stored in the UI.
        </div>
      </div>

      {/* Danger Zone */}
      <div style={sectionStyle}>
        <div style={{ ...sectionTitleStyle, color: '#ef4444' }}>Danger Zone</div>
        <div style={dangerRowStyle}>
          <div>
            <div style={rowLabelStyle}>Delete all context</div>
            <div style={rowSubtextStyle}>
              Permanently remove all context chunks for this project.
            </div>
          </div>
          <Button
            variant="danger"
            size="sm"
            onClick={() => setConfirmAction('deleteContext')}
            disabled={!currentProjectId}
          >
            Delete Context
          </Button>
        </div>
        <div style={dangerRowStyle}>
          <div>
            <div style={rowLabelStyle}>Delete project</div>
            <div style={rowSubtextStyle}>
              Permanently delete this project, all tasks, and sessions.
            </div>
          </div>
          <Button
            variant="danger"
            size="sm"
            onClick={() => setConfirmAction('deleteProject')}
            disabled={!currentProjectId}
          >
            Delete Project
          </Button>
        </div>
      </div>

      <Modal
        isOpen={!!confirmAction}
        onClose={() => setConfirmAction(null)}
        title="Are you sure?"
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
            {confirmAction ? confirmMessages[confirmAction] : ''}
          </p>
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
            <Button variant="ghost" size="md" onClick={() => setConfirmAction(null)}>
              Cancel
            </Button>
            <Button
              variant="danger"
              size="md"
              onClick={handleDangerAction}
              disabled={isActioning}
            >
              {isActioning ? 'Deleting…' : 'Confirm Delete'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}

export default SettingsPage
