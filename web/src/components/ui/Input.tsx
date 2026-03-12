import { type CSSProperties, type InputHTMLAttributes, useState } from 'react'

interface InputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'onChange'> {
  label?: string
  error?: string
  onChange?: (value: string) => void
}

export function Input({
  label,
  error,
  id,
  name,
  type = 'text',
  placeholder,
  value,
  onChange,
  disabled = false,
  className,
  ...props
}: InputProps) {
  const [focused, setFocused] = useState(false)

  const wrapperStyle: CSSProperties = {
    display: 'flex',
    flexDirection: 'column',
    gap: '6px',
    width: '100%',
  }

  const labelStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '12px',
    fontWeight: 500,
    color: 'var(--text-secondary)',
    lineHeight: 1,
  }

  const inputStyle: CSSProperties = {
    height: '34px',
    padding: '0 12px',
    background: disabled ? 'var(--bg-elevated)' : 'var(--bg-surface)',
    border: error
      ? '1px solid rgba(239, 68, 68, 0.50)'
      : focused
        ? '1px solid var(--accent)'
        : '1px solid var(--border-default)',
    borderRadius: 'var(--radius-md)',
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
    color: disabled ? 'var(--text-disabled)' : 'var(--text-primary)',
    outline: 'none',
    boxShadow: focused && !error ? '0 0 0 3px var(--accent-dim)' : 'none',
    transition: 'all var(--duration-default) var(--ease-default)',
    width: '100%',
    cursor: disabled ? 'not-allowed' : 'text',
  }

  const errorStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '12px',
    color: '#f87171',
    lineHeight: 1,
  }

  // Inject placeholder style via a style tag approach — we handle it via CSS variable
  const placeholderColor = disabled ? 'var(--text-disabled)' : 'var(--text-disabled)'

  return (
    <div style={wrapperStyle} className={className}>
      {label && (
        <label htmlFor={id} style={labelStyle}>
          {label}
        </label>
      )}
      <input
        {...props}
        id={id}
        name={name}
        type={type}
        placeholder={placeholder}
        value={value}
        disabled={disabled}
        style={{
          ...inputStyle,
          // CSS custom property for placeholder color is handled in globals.css
        }}
        onChange={(e) => onChange?.(e.target.value)}
        onFocus={() => setFocused(true)}
        onBlur={() => setFocused(false)}
        // Inline placeholder color via CSS workaround
        data-placeholder-color={placeholderColor}
      />
      {error && <span style={errorStyle}>{error}</span>}
    </div>
  )
}

export default Input
