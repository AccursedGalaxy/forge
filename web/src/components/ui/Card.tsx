import { type CSSProperties, type ReactNode, useState } from 'react'

interface CardProps {
  children: ReactNode
  hoverable?: boolean
  className?: string
  onClick?: () => void
  style?: CSSProperties
}

export function Card({ children, hoverable = false, className, onClick, style }: CardProps) {
  const [hovered, setHovered] = useState(false)

  const cardStyle: CSSProperties = {
    background: 'var(--bg-surface)',
    border: hovered && hoverable
      ? '1px solid var(--border-default)'
      : '1px solid var(--border-subtle)',
    borderRadius: 'var(--radius-md)',
    padding: '16px',
    boxShadow: hovered && hoverable ? 'var(--shadow-md)' : 'var(--shadow-sm)',
    transition: 'all var(--duration-default) var(--ease-default)',
    cursor: onClick ? 'pointer' : 'default',
    ...style,
  }

  return (
    <div
      style={cardStyle}
      className={className}
      onClick={onClick}
      onMouseEnter={() => hoverable && setHovered(true)}
      onMouseLeave={() => hoverable && setHovered(false)}
    >
      {children}
    </div>
  )
}

export default Card
