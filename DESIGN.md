# FORGE Design System — “OBSIDIAN”

> Single source of truth for all UI decisions. Every component, page, and layout must reference this document.
> When in doubt: less is more. Restraint is the aesthetic.

-----

## Core Aesthetic

**Identity:** Premium developer tooling. Calm, confident, expensive. The kind of interface that makes you trust the product before you’ve used it.

**Tone:** Refined without being sterile. Dark without being oppressive. Dense without being cluttered.

**The one rule:** If it feels like a generic SaaS dashboard, it’s wrong.

-----

## Color System

All colors are CSS variables defined in `styles/tokens.css`. Never hardcode hex values in components.

```css
:root {
  /* Backgrounds — layered depth */
  --bg-base:     #111114;   /* page background */
  --bg-surface:  #18181C;   /* cards, panels */
  --bg-elevated: #1E1E24;   /* dropdowns, modals, hover */
  --bg-overlay:  #24242C;   /* tooltips, popovers */

  /* Borders */
  --border-subtle:  rgba(255, 255, 255, 0.06);
  --border-default: rgba(255, 255, 255, 0.10);
  --border-strong:  rgba(255, 255, 255, 0.18);

  /* Text */
  --text-primary:  #F2F2F0;
  --text-secondary: #A1A1AA;
  --text-muted:    #71717A;
  --text-disabled: #3F3F46;

  /* Accent — muted violet */
  --accent:         #A78BFA;
  --accent-dim:     rgba(167, 139, 250, 0.15);
  --accent-hover:   #BBA4FB;

  /* Semantic status colors */
  --status-backlog:     #3F3F46;   /* zinc-700 */
  --status-planned:     #1D4ED8;   /* blue-700 */
  --status-in-progress: #B45309;   /* amber-700 */
  --status-review:      #6D28D9;   /* violet-700 */
  --status-done:        #047857;   /* emerald-700 */

  --status-backlog-text:     #A1A1AA;
  --status-planned-text:     #93C5FD;
  --status-in-progress-text: #FCD34D;
  --status-review-text:      #C4B5FD;
  --status-done-text:        #6EE7B7;

  /* Autonomy levels */
  --autonomy-supervised:  #1E3A5F;
  --autonomy-supervised-text: #93C5FD;
  --autonomy-checkpoint:  #3B1F5E;
  --autonomy-checkpoint-text: #C4B5FD;
  --autonomy-autonomous:  #1A3D2E;
  --autonomy-autonomous-text: #6EE7B7;

  /* Shadows */
  --shadow-sm: 0 2px 8px rgba(0, 0, 0, 0.3);
  --shadow-md: 0 4px 16px rgba(0, 0, 0, 0.4);
  --shadow-lg: 0 8px 32px rgba(0, 0, 0, 0.5);
  --shadow-xl: 0 16px 48px rgba(0, 0, 0, 0.6);
}
```

-----

## Typography

```css
:root {
  /* Display / hero text */
  --font-display: 'Instrument Serif', Georgia, serif;

  /* All UI chrome, labels, body */
  --font-ui: 'Geist', system-ui, sans-serif;

  /* Code, terminal output, session streams, monospace labels */
  --font-mono: 'Geist Mono', 'JetBrains Mono', monospace;
}
```

**Import in `app/layout.tsx`:**

```tsx
import { Geist, Geist_Mono, Instrument_Serif } from 'next/font/google';
```

### Type Scale

|Token           |Size   |Weight     |Usage                                   |
|----------------|-------|-----------|----------------------------------------|
|`text-display`  |48–72px|400 (serif)|Hero headlines — use italic for emphasis|
|`text-heading-1`|28px   |600        |Page titles                             |
|`text-heading-2`|20px   |600        |Section headers, panel titles           |
|`text-heading-3`|15px   |600        |Card titles, group labels               |
|`text-body`     |14px   |400        |Default body copy                       |
|`text-small`    |12px   |400        |Secondary labels, timestamps            |
|`text-micro`    |11px   |500        |Badges, status tags, caps labels        |
|`text-code`     |13px   |400        |Terminal output, code snippets          |

### Typography Rules

- Display font (Instrument Serif) is **only** for hero/marketing text. Never in app chrome.
- Italic Instrument Serif for emotional emphasis in headlines: *“Ship faster.”*
- All caps + letter-spacing for category labels and section headers in UI: `ACTIVE SESSIONS`
- Monospace for anything code-adjacent: session IDs, git refs, timestamps in logs

-----

## Spacing

Base unit: `4px`. All spacing is multiples of 4.

```
4px   — xs: tight internal padding
8px   — sm: component internal spacing
12px  — md: default padding inside cards
16px  — lg: standard gap between elements
24px  — xl: section spacing within panels
32px  — 2xl: major section breaks
48px  — 3xl: page-level padding
64px  — 4xl: hero sections
```

**Layout constants:**

```css
--sidebar-width: 240px;
--sidebar-collapsed: 56px;
--topbar-height: 52px;
--session-panel-width: 420px;
--content-max-width: 1400px;
```

-----

## Border Radius

```css
--radius-sm:   4px;   /* inputs, small badges */
--radius-md:   6px;   /* cards, buttons, panels — DEFAULT */
--radius-lg:   10px;  /* modals, large panels */
--radius-full: 9999px; /* pill badges only */
```

**Rule:** Default to `--radius-md` (6px). Never go above `--radius-lg`. Pills only for autonomy/status badges.

-----

## Motion

```css
--duration-fast:   100ms;
--duration-default: 150ms;
--duration-slow:   250ms;
--duration-enter:  200ms;

--ease-default: cubic-bezier(0.16, 1, 0.3, 1);
--ease-out:     cubic-bezier(0, 0, 0.2, 1);
```

**Rules:**

- All interactive elements: `transition: all var(--duration-default) var(--ease-default)`
- Page/panel mounts: fade-in + subtle translateY(4px) → translateY(0)
- No bouncy spring animations. No dramatic slides. Subtle always wins.
- Skeleton loaders on all async data — never raw spinners

-----

## Component Primitives

### Button

Three variants. No others.

```
primary   — bg: accent-dim, text: accent, border: accent/30, hover: accent-dim stronger
ghost     — bg: transparent, text: secondary, border: transparent, hover: bg-elevated
danger    — bg: red/10, text: red-400, border: red/20, hover: red/15
```

Sizes: `sm` (28px height), `md` (34px height — default), `lg` (40px height)

All buttons: `--radius-md`, `--font-ui`, `font-size: 13px`, `font-weight: 500`

### Badge / Tag

```
status badge  — pill shape (--radius-full), 11px, uppercase, letter-spacing 0.05em
autonomy badge — pill shape, color from autonomy tokens
count badge   — circular, 16px, accent color, for notification counts
```

### Input / Textarea

```
bg: --bg-surface
border: --border-default
border-radius: --radius-md
height: 34px (single line)
padding: 0 12px
font: --font-ui, 13px
focus: border-color: --accent, box-shadow: 0 0 0 3px var(--accent-dim)
placeholder: --text-disabled
```

### Card

```
bg: --bg-surface
border: 1px solid --border-subtle
border-radius: --radius-md
padding: 16px
shadow: --shadow-sm
hover (interactive cards): border-color: --border-default, shadow: --shadow-md
transition: --duration-default
```

### Sidebar

```
width: --sidebar-width (240px)
bg: --bg-base
border-right: 1px solid --border-subtle
height: 100vh, position: fixed

Nav items:
  height: 32px
  padding: 0 12px
  border-radius: --radius-md
  font: 13px, weight 500
  color: --text-secondary
  hover: bg: --bg-elevated, color: --text-primary
  active: bg: --accent-dim, color: --accent
```

### TopBar

```
height: --topbar-height (52px)
bg: --bg-base
border-bottom: 1px solid --border-subtle
padding: 0 24px
position: sticky, top: 0
z-index: 40
```

### Modal / Dialog

```
overlay: rgba(0,0,0,0.6) backdrop-blur(4px)
panel: bg: --bg-elevated, border: --border-default, radius: --radius-lg, shadow: --shadow-xl
max-width: 480px default
padding: 24px
```

### Session Stream / Terminal

```
bg: #0D0D10
border: 1px solid --border-subtle
border-radius: --radius-md
font: --font-mono, 12px
line-height: 1.6
color: #E2E2DC
padding: 16px

Line types:
  default output: #E2E2DC
  tool call (⚡):  --accent
  thinking (💭):  --text-muted, italic
  error (✗):     #F87171
  success (✓):   #6EE7B7
  timestamp:     --text-disabled, font-size: 11px
```

### Kanban Column

```
width: 260px, min-width: 260px
bg: --bg-surface
border: 1px solid --border-subtle
border-radius: --radius-md
padding: 12px

Header:
  status badge + column name + task count
  font: 12px, 600, uppercase, letter-spacing 0.08em

Task card gap: 8px
```

### Task Card

```
bg: --bg-elevated
border: 1px solid --border-subtle
border-radius: --radius-md
padding: 12px
cursor: pointer

hover: border-color: --border-default, shadow: --shadow-md

Contents:
  title: 13px, 500, --text-primary
  description: 12px, --text-muted, 2-line clamp
  footer row: autonomy badge + session status indicator
  active session dot: 6px circle, --accent, subtle pulse animation
```

-----

## Layout Structure

```
AppShell
├── Sidebar (fixed, 240px)
│   ├── Logo / wordmark (top)
│   ├── Project switcher
│   ├── Nav items (Board, Sessions, Context, Settings)
│   └── Footer (Clerk UserButton + user name)
└── Main (margin-left: 240px)
    ├── TopBar (sticky)
    │   ├── Breadcrumb
    │   └── Actions (New Task, Run Agent)
    └── ContentArea (padding: 24px)
        ├── KanbanBoard (default)
        ├── SessionView (task active)
        └── ProjectSettings
```

**Session panel:** slides in from the right as an overlay panel (position fixed, right: 0) when a session is active. Does not push the kanban board. Width: `--session-panel-width` (420px).

-----

## Page-Specific Rules

### Landing Page

- Instrument Serif italic for the hero headline
- One-sentence product description in Geist, `--text-secondary`
- CTA button: `primary` variant, `lg` size
- Feature sections: alternating layout, mock UI placeholders as styled divs
- Pricing: two columns (Open Source / Cloud), clean feature lists
- No hero images — typography and layout carry the visual weight

### Dashboard / Kanban

- Kanban columns scroll horizontally if overflow
- Board itself scrolls, not the page
- Empty column state: dashed border, muted label “No tasks”
- Loading state: skeleton cards in each column

### Session View

- Full-height right panel
- Header: task title + status badge + close button
- Stream area: scrolls to bottom automatically
- Footer: interrupt button (danger variant) when session is running

-----

## Do / Don’t

|Do                                 |Don’t                              |
|-----------------------------------|-----------------------------------|
|Use CSS variables for every color  |Hardcode hex values in components  |
|Instrument Serif for hero text only|Use serif font in UI chrome        |
|Subtle 150ms transitions           |Bouncy or dramatic animations      |
|Empty states with copy and action  |Leave blank white/dark space       |
|Skeleton loaders for async data    |Raw spinners                       |
|Pill badges for status/autonomy    |Square badges for these            |
|6px border radius by default       |Fully rounded or fully square cards|
|Geist Mono for terminal/code output|Use UI font for code               |
|Clerk UserButton as-is             |Style or wrap Clerk components     |
|Build from primitives              |Reach for shadcn or Radix          |

-----

## File Structure

```
web/
  styles/
    tokens.css          ← CSS variables (this document's color/spacing system)
    globals.css         ← resets, base styles, font imports
  components/
    ui/                 ← primitives (Button, Badge, Input, Card, Modal)
    forge/              ← feature components (TaskCard, KanbanColumn, SessionStream, Sidebar)
  app/
    layout.tsx          ← font imports, token stylesheet, AppShell
    page.tsx            ← landing
    dashboard/
      page.tsx          ← kanban board
      layout.tsx        ← AppShell with sidebar
```
