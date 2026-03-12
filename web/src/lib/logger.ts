type LogLevel = 'debug' | 'info' | 'warn' | 'error'

interface LogEntry {
  timestamp: string
  level: LogLevel
  component: string
  message: string
  meta?: unknown
}

const BUFFER_MAX = 50
const FLUSH_INTERVAL_MS = 5000
const BASE_URL = (import.meta.env.VITE_API_URL as string | undefined) ?? 'http://localhost:8080'

const levelColors: Record<LogLevel, string> = {
  debug: '#6b7280',
  info: '#a78bfa',
  warn: '#f59e0b',
  error: '#ef4444',
}

class Logger {
  private buffer: LogEntry[] = []
  private timer: ReturnType<typeof setInterval> | null = null

  constructor() {
    if (!import.meta.env.DEV) {
      this.timer = setInterval(() => this.flush(), FLUSH_INTERVAL_MS)
    }
  }

  private write(level: LogLevel, component: string, message: string, meta?: unknown) {
    const entry: LogEntry = {
      timestamp: new Date().toISOString(),
      level,
      component,
      message,
      meta,
    }

    if (import.meta.env.DEV) {
      const color = levelColors[level]
      const tag = `%c[${component}]%c`
      const style = `color: ${color}; font-weight: 600`
      if (meta !== undefined) {
        console[level === 'debug' ? 'log' : level](tag, style, 'color: inherit', message, meta)
      } else {
        console[level === 'debug' ? 'log' : level](tag, style, 'color: inherit', message)
      }
    } else {
      this.buffer.push(entry)
      if (this.buffer.length >= BUFFER_MAX || level === 'error') {
        this.flush()
      }
    }
  }

  private flush() {
    if (this.buffer.length === 0) return
    const entries = this.buffer.splice(0)
    fetch(`${BASE_URL}/api/logs`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(entries),
    }).catch(() => {
      // Silently discard — logging should never break the app
    })
  }

  debug(component: string, message: string, meta?: unknown) {
    this.write('debug', component, message, meta)
  }

  info(component: string, message: string, meta?: unknown) {
    this.write('info', component, message, meta)
  }

  warn(component: string, message: string, meta?: unknown) {
    this.write('warn', component, message, meta)
  }

  error(component: string, message: string, meta?: unknown) {
    this.write('error', component, message, meta)
  }

  destroy() {
    if (this.timer) clearInterval(this.timer)
  }
}

const log = new Logger()
export default log
