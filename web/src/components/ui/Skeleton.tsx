import type { CSSProperties } from 'react'

interface SkeletonProps {
  width?: string | number
  height?: string | number
  className?: string
  style?: CSSProperties
}

export function Skeleton({ width = '100%', height = '16px', className, style }: SkeletonProps) {
  const skeletonStyle: CSSProperties = {
    display: 'block',
    width: typeof width === 'number' ? `${width}px` : width,
    height: typeof height === 'number' ? `${height}px` : height,
    borderRadius: 'var(--radius-md)',
    background: `linear-gradient(
      90deg,
      var(--bg-elevated) 0px,
      var(--bg-overlay) 40px,
      var(--bg-elevated) 80px
    )`,
    backgroundSize: '800px 100%',
    animation: 'shimmer 1.6s infinite linear',
    ...style,
  }

  return <span style={skeletonStyle} className={className} aria-hidden="true" />
}

export default Skeleton
