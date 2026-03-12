import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './index.html',
    './src/**/*.{ts,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        'bg-base':     'var(--bg-base)',
        'bg-surface':  'var(--bg-surface)',
        'bg-elevated': 'var(--bg-elevated)',
        'bg-overlay':  'var(--bg-overlay)',
        'border-subtle':  'var(--border-subtle)',
        'border-default': 'var(--border-default)',
        'border-strong':  'var(--border-strong)',
        'text-primary':   'var(--text-primary)',
        'text-secondary': 'var(--text-secondary)',
        'text-muted':     'var(--text-muted)',
        'text-disabled':  'var(--text-disabled)',
        'accent':         'var(--accent)',
        'accent-dim':     'var(--accent-dim)',
        'accent-hover':   'var(--accent-hover)',
      },
      fontFamily: {
        display: ['var(--font-display)'],
        ui:      ['var(--font-ui)'],
        mono:    ['var(--font-mono)'],
      },
      borderRadius: {
        sm:   'var(--radius-sm)',
        md:   'var(--radius-md)',
        lg:   'var(--radius-lg)',
        full: 'var(--radius-full)',
      },
      boxShadow: {
        sm: 'var(--shadow-sm)',
        md: 'var(--shadow-md)',
        lg: 'var(--shadow-lg)',
        xl: 'var(--shadow-xl)',
      },
      transitionDuration: {
        fast:    'var(--duration-fast)',
        default: 'var(--duration-default)',
        slow:    'var(--duration-slow)',
        enter:   'var(--duration-enter)',
      },
      spacing: {
        'sidebar':       'var(--sidebar-width)',
        'sidebar-col':   'var(--sidebar-collapsed)',
        'topbar':        'var(--topbar-height)',
        'session-panel': 'var(--session-panel-width)',
      },
    },
  },
  plugins: [],
}

export default config
