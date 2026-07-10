# Vyzorix UI Theme

## Colors

### Brand Accents
| Token | Hex | Usage |
|-------|-----|-------|
| `rose-400` | `#fb7185` | Keywords, highlights, active states |
| `rose-500` | `#f43f5e` | Primary accent, logo, chart lines |
| `rose-600` | `#e11d48` | Primary buttons, badges |
| `rose-700` | `#be123c` | Hover states |
| `rose-800` | `#9f1239` | Pressed states |

### Neutrals (UI Only)
| Token | Hex | Usage |
|-------|-----|-------|
| `white` | `#ffffff` | Text on dark |
| `gray-50` | `#fafafa` | Light backgrounds |
| `gray-100` | `#f5f5f5` | Light borders |
| `gray-200` | `#e5e5e5` | Light secondary |
| `gray-300` | `#d4d4d4` | Light muted |
| `gray-400` | `#a3a3a3` | Secondary text |
| `gray-500` | `#737373` | Muted text |
| `gray-600` | `#525252` | Disabled |
| `gray-700` | `#404040` | Strong borders |
| `gray-800` | `#262626` | Tertiary bg |
| `gray-900` | `#171717` | Secondary bg |
| `black` | `#0d0d0d` | Primary bg |
| `gray-950` | `#030303` | Deepest bg |

### Default Theme (Dark)
```css
--bg: #0d0d0d;           /* Primary background */
--bg-secondary: #171717;  /* Cards, headers */
--bg-tertiary: #262626;   /* Hover, active bg */
--border: #2e2e2e;        /* Borders */
--border-strong: #404040; /* Emphasized borders */
--text: #e5e5e5;          /* Primary text */
--text-secondary: #a3a3a3; /* Secondary text */
--text-muted: #6b6b6b;    /* Muted text */
```

## Layout Rules

1. **No gradients** - All backgrounds are solid colors
2. **Bordered sections** - Use `border: 1px solid var(--border)` not shadows
3. **Flat design** - No box shadows
4. **Monospace** - Use for technical data (IMEI, versions, code)

## Spinner Component (Mini Block)

Standard inline spinner for buttons and loading states.

```jsx
<div className="relative w-4 h-4">
  {/* Outer ring - white border, rotating */}
  <div className="absolute inset-0 border-2 border-white rounded-[3px] animate-block-spin"></div>
</div>
```

**CSS Animation:**
```css
@keyframes block-spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
.animate-block-spin {
  animation: block-spin 2s linear infinite;
}
```

**Specs:**
| Element | Size | Color | Border-radius | Animation |
|---------|------|-------|--------------|-----------|
| Container | 16x16px | - | - | - |
| Outer ring | inset-0 | border-white (2px) | 3px | rotate 2s linear infinite |

## Skeleton Component

Loading placeholder with shimmer animation.

```css
.skeleton {
  background: linear-gradient(90deg, var(--bg-tertiary) 25%, var(--bg-secondary) 50%, var(--bg-tertiary) 75%);
  background-size: 200% 100%;
  animation: skeleton-shimmer 1.5s ease-in-out infinite;
  border-radius: 4px;
}

@keyframes skeleton-shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
```

**Specs:**
| Property | Value |
|----------|-------|
| Background | Shifts from bg-tertiary â bg-secondary â bg-tertiary |
| Animation | 1.5s ease-in-out, infinite loop |
| Border-radius | 4px |

## Component Strategy

### Use packages/ui (Radix-based, style override)
Override variants to match theme. All Radix primitives for accessibility.

| Component | Source | Notes |
|-----------|--------|-------|
| Button | packages/ui/button | Variants: primary (rose-600), outline, ghost, secondary |
| Input | packages/ui/input | Flat style, no shadows |
| Textarea | packages/ui/textarea | Same styling as Input |
| Select | packages/ui/select | Dropdown with bordered style |
| Tabs | packages/ui/tabs | Bottom border accent on active |
| Dialog | packages/ui/dialog | Bordered overlay, no blur |
| Sheet | packages/ui/sheet | Side drawer, bordered |
| Badge | packages/ui/badge | Rose background for alerts |
| Progress | packages/ui/progress | Thin bar, rose fill |
| Table | packages/ui/table | Bordered cells |
| Tooltip | packages/ui/tooltip | Simple bordered |
| Switch | packages/ui/switch | Rose accent on active |
| Checkbox | packages/ui/checkbox | Rose accent |
| Toggle | packages/ui/toggle | Button group style |
| Separator | packages/ui/separator | Horizontal dividers |
| Skeleton | packages/ui/skeleton | Loading placeholders |
| Spinner | packages/ui/spinner | Loading indicator |
| Alert | packages/ui/alert | Bordered, rose accent for error |

### Create Custom (Dashboard-specific)
These require exact layout/styling per mocks.

| Component | Location | Purpose |
|-----------|----------|---------|
| Section | ui/components/section | Bordered container with header |
| SectionHeader | ui/components/section | Uppercase label, secondary bg |
| MetricCard | ui/components/metric-card | Risk/Thermal/Uptime/Buffer display |
| ConnectionStatus | ui/components/connection-status | WS/FCM indicators |
| ActivityFeed | ui/components/activity-feed | Recent events list |
| CodeBlock | ui/components/code-block | Syntax highlighting |
| CopyButton | ui/components/copy-button | Copy to clipboard |
| StatusDot | ui/components/status-dot | Green/rose indicators |
| TimeRangeSelector | ui/components/time-range | 1h/6h/24h/7d buttons |
| NavItem | ui/components/nav-item | Sidebar navigation |
| Header | ui/components/header | Page header with status |
| Chart (SVG) | ui/components/chart | Line chart for metrics |
| CopyButton | ui/components/copy-button | Copy code snippets |

## Reference

See `apps/vyoriX/mocks/` for implementation examples:
- `dashboard.html` - Overview page layout (Section, MetricCard, ConnectionStatus)
- `metrics.html` - Charts and stats (TimeRangeSelector, Chart)
- `developer.html` - Code syntax highlighting (CodeBlock, CopyButton)
- `api-reference.html` - API documentation (Table, CodeBlock)
