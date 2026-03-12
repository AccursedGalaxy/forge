import type { CSSProperties, ButtonHTMLAttributes, ReactNode } from 'react'

export type ButtonVariant = 'primary' | 'ghost' | 'danger'
export type ButtonSize = 'sm' | 'md' | 'lg'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  size?: ButtonSize
  children: ReactNode
  className?: string
}

const heightMap: Record<ButtonSize, string> = {
  sm: '28px',
  md: '34px',
  lg: '40px',
}

const paddingMap: Record<ButtonSize, string> = {
  sm: '0 10px',
  md: '0 14px',
  lg: '0 18px',
}

function getVariantStyles(variant: ButtonVariant, disabled: boolean): CSSProperties {
  if (disabled) {
    return {
      background: 'transparent',
      color: 'var(--text-disabled)',
      border: '1px solid var(--border-subtle)',
      cursor: 'not-allowed',
    }
  }

  switch (variant) {
    case 'primary':
      return {
        background: 'var(--accent-dim)',
        color: 'var(--accent)',
        border: '1px solid rgba(167, 139, 250, 0.30)',
      }
    case 'ghost':
      return {
        background: 'transparent',
        color: 'var(--text-secondary)',
        border: '1px solid transparent',
      }
    case 'danger':
      return {
        background: 'rgba(239, 68, 68, 0.10)',
        color: '#f87171',
        border: '1px solid rgba(239, 68, 68, 0.20)',
      }
  }
}

export function Button({
  variant = 'primary',
  size = 'md',
  children,
  className,
  disabled = false,
  style,
  onMouseEnter,
  onMouseLeave,
  ...props
}: ButtonProps) {
  const variantStyles = getVariantStyles(variant, disabled)

  const baseStyle: CSSProperties = {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    gap: '6px',
    height: heightMap[size],
    padding: paddingMap[size],
    borderRadius: 'var(--radius-md)',
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
    fontWeight: 500,
    lineHeight: 1,
    transition: 'all var(--duration-default) var(--ease-default)',
    whiteSpace: 'nowrap',
    userSelect: 'none',
    ...variantStyles,
    ...style,
  }

  function handleMouseEnter(e: React.MouseEvent<HTMLButtonElement>) {
    if (!disabled) {
      const target = e.currentTarget
      switch (variant) {
        case 'primary':
          target.style.background = 'rgba(167, 139, 250, 0.22)'
          target.style.color = 'var(--accent-hover)'
          break
        case 'ghost':
          target.style.background = 'var(--bg-elevated)'
          target.style.color = 'var(--text-primary)'
          break
        case 'danger':
          target.style.background = 'rgba(239, 68, 68, 0.15)'
          break
      }
    }
    onMouseEnter?.(e)
  }

  function handleMouseLeave(e: React.MouseEvent<HTMLButtonElement>) {
    if (!disabled) {
      const target = e.currentTarget
      const styles = getVariantStyles(variant, disabled)
      target.style.background = styles.background as string
      target.style.color = styles.color as string
    }
    onMouseLeave?.(e)
  }

  return (
    <button
      {...props}
      disabled={disabled}
      style={baseStyle}
      className={className}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      {children}
    </button>
  )
}

export default Button
