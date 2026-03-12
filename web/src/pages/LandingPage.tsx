import type { CSSProperties } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '../components/ui'

const steps = [
  {
    number: '01',
    title: 'Define your task',
    description: 'Describe what needs to be built. FORGE breaks it into structured work items and assigns them to specialized agents.',
  },
  {
    number: '02',
    title: 'Orchestrate agents',
    description: 'Multiple AI agents work in parallel across your codebase — each with a clear context boundary and defined responsibility.',
  },
  {
    number: '03',
    title: 'Review and ship',
    description: 'Inspect agent output, review diffs, approve checkpoints, and merge when confident. You stay in control.',
  },
]

const features = [
  {
    title: 'Multi-Agent Orchestration',
    description: 'Coordinate specialized agents across tasks, sessions, and codebases without micromanagement.',
  },
  {
    title: 'Autonomy Control',
    description: 'Set supervised, checkpoint, or fully autonomous modes per task. Your codebase, your rules.',
  },
  {
    title: 'Context Engine',
    description: 'Persistent memory across sessions. Agents remember architecture decisions, patterns, and constraints.',
  },
  {
    title: 'Kanban Board',
    description: 'Visual task management built for AI-augmented development workflows. Not Jira.',
  },
]

const openSourceFeatures = [
  'Self-hosted deployment',
  'Unlimited local agents',
  'Full source access',
  'Community support',
  'Basic orchestration',
]

const cloudFeatures = [
  'Managed infrastructure',
  'Hosted agent pool',
  'Priority model access',
  'Team collaboration',
  'Advanced orchestration',
  'Audit logs & SSO',
]

export function LandingPage() {
  const navigate = useNavigate()

  const pageStyle: CSSProperties = {
    background: 'var(--bg-base)',
    minHeight: '100vh',
    color: 'var(--text-primary)',
  }

  // Hero
  const heroStyle: CSSProperties = {
    maxWidth: '860px',
    margin: '0 auto',
    padding: '120px 24px 96px',
    textAlign: 'center',
  }

  const heroHeadlineStyle: CSSProperties = {
    fontFamily: 'var(--font-display)',
    fontStyle: 'italic',
    fontSize: 'clamp(52px, 8vw, 88px)',
    fontWeight: 400,
    color: 'var(--text-primary)',
    lineHeight: 1.05,
    letterSpacing: '-0.02em',
    marginBottom: '24px',
  }

  const heroSubtitleStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '17px',
    color: 'var(--text-secondary)',
    lineHeight: 1.6,
    marginBottom: '40px',
    maxWidth: '520px',
    margin: '0 auto 40px',
  }

  // Section wrapper
  const sectionStyle: CSSProperties = {
    maxWidth: 'var(--content-max-width)',
    margin: '0 auto',
    padding: '80px 24px',
  }

  const sectionDivider: CSSProperties = {
    borderTop: '1px solid var(--border-subtle)',
  }

  const sectionLabelStyle: CSSProperties = {
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    fontWeight: 500,
    color: 'var(--accent)',
    letterSpacing: '0.12em',
    textTransform: 'uppercase',
    marginBottom: '12px',
  }

  const sectionTitleStyle: CSSProperties = {
    fontFamily: 'var(--font-display)',
    fontSize: 'clamp(28px, 4vw, 40px)',
    fontWeight: 400,
    color: 'var(--text-primary)',
    lineHeight: 1.2,
    marginBottom: '48px',
  }

  // Steps
  const stepsGridStyle: CSSProperties = {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))',
    gap: '24px',
  }

  const stepCardStyle: CSSProperties = {
    background: 'var(--bg-surface)',
    border: '1px solid var(--border-subtle)',
    borderRadius: 'var(--radius-lg)',
    padding: '28px',
  }

  const stepNumberStyle: CSSProperties = {
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    fontWeight: 700,
    color: 'var(--accent)',
    letterSpacing: '0.08em',
    marginBottom: '16px',
    display: 'block',
  }

  const stepTitleStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '15px',
    fontWeight: 600,
    color: 'var(--text-primary)',
    marginBottom: '10px',
    lineHeight: 1.3,
  }

  const stepDescStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
    color: 'var(--text-secondary)',
    lineHeight: 1.6,
  }

  // Features
  const featuresGridStyle: CSSProperties = {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, 1fr)',
    gap: '16px',
  }

  const featureBoxStyle: CSSProperties = {
    background: 'var(--bg-surface)',
    border: '1px solid var(--border-subtle)',
    borderRadius: 'var(--radius-lg)',
    padding: '28px',
    position: 'relative',
    overflow: 'hidden',
  }

  const featureAccentBarStyle: CSSProperties = {
    position: 'absolute',
    top: 0,
    left: 0,
    width: '3px',
    height: '100%',
    background: 'var(--accent-dim)',
    borderRadius: '0 0 0 var(--radius-lg)',
  }

  const featureTitleStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '14px',
    fontWeight: 600,
    color: 'var(--text-primary)',
    marginBottom: '8px',
  }

  const featureDescStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
    color: 'var(--text-secondary)',
    lineHeight: 1.6,
  }

  // Pricing
  const pricingGridStyle: CSSProperties = {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, 1fr)',
    gap: '16px',
    maxWidth: '720px',
  }

  const pricingCardStyle = (highlighted: boolean): CSSProperties => ({
    background: highlighted ? 'var(--bg-elevated)' : 'var(--bg-surface)',
    border: highlighted ? '1px solid rgba(167, 139, 250, 0.25)' : '1px solid var(--border-subtle)',
    borderRadius: 'var(--radius-lg)',
    padding: '28px',
  })

  const pricingTierStyle: CSSProperties = {
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    fontWeight: 700,
    color: 'var(--accent)',
    letterSpacing: '0.10em',
    textTransform: 'uppercase',
    marginBottom: '8px',
    display: 'block',
  }

  const pricingTitleStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '20px',
    fontWeight: 600,
    color: 'var(--text-primary)',
    marginBottom: '4px',
  }

  const pricingSubtitleStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
    color: 'var(--text-muted)',
    marginBottom: '24px',
    lineHeight: 1.4,
  }

  const featureListStyle: CSSProperties = {
    listStyle: 'none',
    padding: 0,
    margin: 0,
    display: 'flex',
    flexDirection: 'column',
    gap: '10px',
  }

  const featureListItemStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
    fontFamily: 'var(--font-ui)',
    fontSize: '13px',
    color: 'var(--text-secondary)',
  }

  const checkStyle: CSSProperties = {
    color: 'var(--status-done-text)',
    fontSize: '12px',
    flexShrink: 0,
    fontFamily: 'var(--font-mono)',
  }

  // Footer
  const footerStyle: CSSProperties = {
    borderTop: '1px solid var(--border-subtle)',
    padding: '32px 24px',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    maxWidth: 'var(--content-max-width)',
    margin: '0 auto',
  }

  const footerLogoStyle: CSSProperties = {
    fontFamily: 'var(--font-mono)',
    fontSize: '13px',
    fontWeight: 700,
    color: 'var(--accent)',
    letterSpacing: '0.08em',
  }

  const footerTaglineStyle: CSSProperties = {
    fontFamily: 'var(--font-ui)',
    fontSize: '12px',
    color: 'var(--text-muted)',
  }

  return (
    <div style={pageStyle}>
      {/* Hero */}
      <section style={heroStyle}>
        <h1 style={heroHeadlineStyle}>Ship faster.<br />Think less.</h1>
        <p style={heroSubtitleStyle}>
          FORGE orchestrates multiple AI coding agents so your team can focus on
          decisions, not busywork.
        </p>
        <div style={{ display: 'flex', justifyContent: 'center', gap: '12px' }}>
          <Button variant="primary" size="lg" onClick={() => navigate('/dashboard')}>
            Get Started
          </Button>
          <Button variant="ghost" size="lg">
            View on GitHub
          </Button>
        </div>
      </section>

      {/* How it works */}
      <div style={sectionDivider}>
        <section style={sectionStyle}>
          <p style={sectionLabelStyle}>How it works</p>
          <h2 style={sectionTitleStyle}>From idea to merged PR,<br />agents handle the gaps.</h2>
          <div style={stepsGridStyle}>
            {steps.map((step) => (
              <div key={step.number} style={stepCardStyle}>
                <span style={stepNumberStyle}>{step.number}</span>
                <h3 style={stepTitleStyle}>{step.title}</h3>
                <p style={stepDescStyle}>{step.description}</p>
              </div>
            ))}
          </div>
        </section>
      </div>

      {/* Features */}
      <div style={sectionDivider}>
        <section style={sectionStyle}>
          <p style={sectionLabelStyle}>Features</p>
          <h2 style={sectionTitleStyle}>Everything your AI workflow needs.</h2>
          <div style={featuresGridStyle}>
            {features.map((feature) => (
              <div key={feature.title} style={featureBoxStyle}>
                <div style={featureAccentBarStyle} />
                <h3 style={featureTitleStyle}>{feature.title}</h3>
                <p style={featureDescStyle}>{feature.description}</p>
              </div>
            ))}
          </div>
        </section>
      </div>

      {/* Pricing */}
      <div style={sectionDivider}>
        <section style={sectionStyle}>
          <p style={sectionLabelStyle}>Pricing</p>
          <h2 style={sectionTitleStyle}>Start free, scale when ready.</h2>
          <div style={pricingGridStyle}>
            {/* Open Source */}
            <div style={pricingCardStyle(false)}>
              <span style={pricingTierStyle}>Free</span>
              <h3 style={pricingTitleStyle}>Open Source</h3>
              <p style={pricingSubtitleStyle}>Self-host on your own infrastructure. No usage limits.</p>
              <ul style={featureListStyle}>
                {openSourceFeatures.map((f) => (
                  <li key={f} style={featureListItemStyle}>
                    <span style={checkStyle}>✓</span>
                    {f}
                  </li>
                ))}
              </ul>
            </div>
            {/* Cloud */}
            <div style={pricingCardStyle(true)}>
              <span style={pricingTierStyle}>Cloud</span>
              <h3 style={pricingTitleStyle}>FORGE Cloud</h3>
              <p style={pricingSubtitleStyle}>Fully managed. Deploy in seconds, scale without limits.</p>
              <ul style={featureListStyle}>
                {cloudFeatures.map((f) => (
                  <li key={f} style={featureListItemStyle}>
                    <span style={checkStyle}>✓</span>
                    {f}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </section>
      </div>

      {/* Footer */}
      <div style={sectionDivider}>
        <footer style={{ ...footerStyle, padding: '40px 24px' }}>
          <div>
            <span style={footerLogoStyle}>FORGE</span>
            <p style={{ ...footerTaglineStyle, marginTop: '4px' }}>
              Multi-agent AI coding orchestration.
            </p>
          </div>
          <p style={footerTaglineStyle}>
            &copy; {new Date().getFullYear()} FORGE. Open source.
          </p>
        </footer>
      </div>
    </div>
  )
}

export default LandingPage
