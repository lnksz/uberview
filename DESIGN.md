---
name: Uberview
description: A compact Monokai desktop workbench for triaging tickets across providers.
colors:
  workbench: "#1e1f1c"
  panel: "#272822"
  ink: "#f8f8f2"
  muted-ink: "#c7c7bb"
  selection-green: "#a6e22e"
  link-cyan: "#66d9ef"
  chrome: "#3e3d32"
  error-pink: "#f92672"
  rule-olive: "#75715e"
  weekend: "#30312b"
  inverse-yellow: "#e6db74"
  provider-violet: "#ae81ff"
  provider-orange: "#fd971f"
  pure-white: "#fff"
  pure-black: "#000"
typography:
  title:
    fontFamily: "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    fontSize: "0.82rem"
    fontWeight: 800
    lineHeight: 1.35
    letterSpacing: "0.08em"
  ticket-title:
    fontFamily: "Geneva, Tahoma, Helvetica Neue, sans-serif"
    fontSize: "0.8rem"
    fontWeight: 600
    lineHeight: 1.35
    letterSpacing: "normal"
  body:
    fontFamily: "Geneva, Tahoma, Helvetica Neue, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.35
    letterSpacing: "normal"
  operational:
    fontFamily: "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    fontSize: "0.7rem"
    fontWeight: 400
    lineHeight: 1.35
    letterSpacing: "0.03em"
  label:
    fontFamily: "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    fontSize: "0.7rem"
    fontWeight: 800
    lineHeight: 1.35
    letterSpacing: "0.08em"
rounded:
  square: "0"
spacing:
  micro: "0.08rem"
  xs: "0.25rem"
  sm: "0.35rem"
  md: "0.55rem"
  lg: "0.7rem"
  panel: "1rem"
components:
  workspace-tab:
    backgroundColor: "{colors.chrome}"
    textColor: "{colors.ink}"
    typography: "{typography.label}"
    rounded: "{rounded.square}"
    padding: "0 0.9rem"
    height: "34px"
  workspace-tab-active:
    backgroundColor: "{colors.selection-green}"
    textColor: "{colors.panel}"
    typography: "{typography.label}"
    rounded: "{rounded.square}"
    padding: "0 0.9rem"
    height: "34px"
  action-button:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    typography: "{typography.operational}"
    rounded: "{rounded.square}"
    padding: "0.25rem 0.55rem"
    height: "28px"
  action-button-active:
    backgroundColor: "{colors.inverse-yellow}"
    textColor: "{colors.panel}"
    typography: "{typography.operational}"
    rounded: "{rounded.square}"
    padding: "0.25rem 0.55rem"
    height: "28px"
  provider-row:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.square}"
    padding: "0.55rem 0.7rem"
  ticket-row:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    typography: "{typography.operational}"
    rounded: "{rounded.square}"
    padding: "0.42rem 0.6rem"
  ticket-label:
    backgroundColor: "transparent"
    textColor: "{colors.ink}"
    typography: "{typography.operational}"
    rounded: "{rounded.square}"
    padding: "0.08rem 0.28rem"
  gantt-bar:
    textColor: "{colors.pure-black}"
    typography: "{typography.operational}"
    rounded: "{rounded.square}"
    padding: "0 0.45rem"
    height: "18px"
---

# Design System: Uberview

## Overview

**Creative North Star: "The Monokai Workbench"**

Uberview is a compact desktop operating surface modeled on a precise editor workbench. Charcoal fields, syntax-bright state colors, square rules, dense monospace data, and a persistent statusline make provider health and ticket urgency legible without pretending to be a terminal. Restrained sans-serif ticket titles provide a deliberate reading contrast inside the operational grid.

The system is flat, direct, and keyboard-first. It names the product and primary destinations plainly as Uberview, Tickets, Gantt, and Providers. Decorative cards, soft dashboard chrome, and hidden interaction conventions are outside the shipped language.

**Key Characteristics:**
- Monokai charcoal surfaces with sparse syntax-bright signals.
- Square geometry, solid structural rules, and dotted row separators.
- Dense monospace controls and metadata paired with sans-serif ticket titles.
- Green bordered keyboard selection and visible white control focus.
- Desktop-only operation at full-screen and half-screen widths.

## Colors

The palette is the shipped Sublime Text Monokai system: dark editor surfaces establish continuity while bright syntax colors communicate state, provider identity, selection, and action.

### Primary
- **Selection Green:** The active Tickets or Gantt tab, online state, selected-row border, and statusline view badge. It is the strongest orientation signal and must remain sparse.
- **Link Cyan:** The first provider identity color and the today label in Gantt. It may identify actionable or temporal information but does not replace the green selection state.

### Secondary
- **Inverse Yellow:** Hovered, focused, or active outlined controls, ticket-count badges, and the distinct Cached provider state use this background with dark inverse text.
- **Error Pink:** Offline hatching and inline provider errors. Preserve textual error messaging; color is reinforcement, not the only cue.
- **Provider Violet:** The second color in the deterministic provider cycle.
- **Provider Orange:** The third color in the deterministic provider cycle.

### Neutral
- **Workbench:** The page field beneath the workbench panel.
- **Panel:** The primary content, table, Gantt, button, and statusline surface.
- **Chrome:** The menu strip, section headers, hover rows, and desktop dot pattern.
- **Weekend:** A subtle Gantt weekend surface differentiated from ordinary rows.
- **Ink:** Primary text and structural borders.
- **Muted Ink:** Secondary metadata, inactive tabs, URLs, dates, statuses, and priorities.
- **Rule Olive:** Dotted separators, secondary outlines, and the alternating provider pattern.
- **Pure White / Pure Black:** Runtime-selected Gantt bar text only, chosen for contrast against each provider color.

### Provider Color Cycle
Provider assignment is deterministic by first-seen order and wraps after six entries: Link Cyan, Provider Violet, Provider Orange, Selection Green, Error Pink, then Inverse Yellow. Use the assigned color on the provider mark and Gantt bar. Do not imply health with this cycle; online and offline states remain green and pink respectively.

### Named Rules
**The Sparse Syntax Rule.** Bright colors identify state, source, or keyboard position; they never become broad decorative fills.

**The Dual-Cue Status Rule.** Online uses a filled green square, offline uses a pink border and diagonal hatch, and Cached uses a yellow square; every mark is paired with visible status or error text. Cached means stored provider data and must not claim live online reachability.

## Typography

**Display Font:** The system has no promotional display face.
**Body Font:** Geneva, with Tahoma and Helvetica Neue fallbacks.
**Label/Mono Font:** UI monospace, with SFMono-Regular, Menlo, Consolas, and monospace fallbacks.

**Character:** Monospace type carries controls, dates, references, status, keyboard hints, and other operational data. The sans-serif face is reserved for ticket titles and task links so long work descriptions remain readable without softening the tool-like hierarchy.

### Hierarchy
- **Product Title:** Heavy, compact, uppercase mono with wide tracking. Use only for Uberview in the top strip.
- **Ticket Title:** Semibold sans serif. Preserve the dotted underline on links and allow wrapping in Tickets; truncate with ellipsis in the Gantt task column.
- **Body:** Restrained sans serif at the page base size. It is not a marketing body-copy style.
- **Operational:** Small mono for row data, provider details, controls, Gantt labels, and the statusline.
- **Label:** Heavy uppercase mono with wide tracking for section titlebars and table headers.

### Named Rules
**The Data Is Mono Rule.** Dates, statuses, priorities, references, provider metadata, controls, and keyboard guidance stay monospace; only human-readable ticket titles switch to sans serif.

**The Direct Naming Rule.** Visible navigation and section language uses Uberview, Tickets, Gantt, and Providers. Do not add ornamental subtitles or rename these destinations with abstract information-architecture jargon.

**The Configured Name Rule.** Provider identity always uses the configured provider `name` in Providers, Tickets, and Gantt. Never substitute generated enumeration such as P01 or P02.

## Layout

The page is a desktop workbench with a sticky 34px top strip and one bordered main panel. At widths above 900px, the top strip uses three columns for Uberview, the Tickets/Gantt switch, and right-aligned keyboard hints. In Tickets, the main panel is centered, capped at 1500px, held 16px from viewport sides, and separated vertically by a fluid 12px to 42px margin. In Gantt, the same main workbench expands to the full viewport width with no side margins so the timeline receives the maximum horizontal field. Providers use an auto-fitting grid with a 330px minimum, yielding three compact columns in the full-screen reference.

At 900px and below, the shipped half-screen mode turns the top strip into two columns and moves keyboard hints to a full-width second row. The Tickets panel uses 10px outer margins, Providers become one column, and Tickets deliberately retain a 980px minimum table width inside horizontal overflow. Gantt remains full viewport width and scrolls inside its own viewport. Its task column defaults to 520px and can be resized from 360px to 900px by pointer or keyboard. Gantt day widths are 38px for month, 22px for quarter, and 14px for year, with 36px task rows.

Vertical rhythm is intentionally compressed: titlebars are about 30-32px, provider rows use compact inset spacing, controls are at least 28px high, and ticket cells use the `sm`/`md` range rather than card-scale padding. Full-screen and half-screen desktop widths are supported. Do not infer or create a mobile-specific card transformation from the half-screen rules.

**The Workbench Frame Rule.** Keep one bordered operating frame; do not fragment Tickets, Gantt, Providers, or status into floating dashboard cards.

**The Preserve Density Rule.** At half-screen width, reflow Providers and enable horizontal work areas rather than enlarging controls or converting table rows into mobile cards.

**The Full-Width Gantt Rule.** Expand the existing workbench frame to full viewport width for Gantt; do not place the timeline inside the centered 1500px Tickets cap.

## Elevation & Depth

The system is flat. It uses no ambient card shadows. Depth comes from tonal layering, a 1px ink frame, solid section boundaries, dotted row rules, sticky headers, subtle 4px dot fields, and inset focus marks. The Gantt today marker is a 2px inset vertical ink rule; keyboard ticket selection is formed with 2px inset green edges so it does not shift layout.

### Named Rules
**The Flat Workbench Rule.** Surfaces remain flush at rest; establish hierarchy with tone, rules, patterns, and sticky positioning rather than drop shadows.

## Shapes

All shipped corners are square because the final cascade applies a global zero radius. Structural frames use 1px solid ink borders. Repeated data rows use 1px dotted olive separators. Labels use dotted current-color outlines, provider marks use solid provider-colored outlines, and status indicators are 9px square marks. Gantt bars are rectangular 18px strips despite earlier rounded declarations in the source cascade.

Focus geometry is explicit: controls and links receive a 2px ink outline with a 2px offset; active workspace tabs draw focus inward. Keyboard-selected ticket and Gantt rows use a 2px green border or inset equivalent without inserting content or shifting columns.

**The Square Rule.** Do not introduce rounded cards, pills, soft containers, or capsule Gantt bars. Zero radius is a system invariant.

## Components

### Workspace Navigation
- **Structure:** A sticky 34px chrome strip containing Uberview, Tickets, Gantt, and keyboard hints.
- **Default:** Tabs use compact uppercase mono; inactive tabs retain ink at reduced opacity and normal weight.
- **Active / Hover / Focus:** Fill the tab green with panel-colored text. Keep the tab boundary and use an inward focus outline.
- **Keyboard:** Show `T Tickets`, `G Gantt`, `R Refresh`, `J/K Move`, and `Enter Open`. `R` refreshes all visible providers.

### Provider Rows
- **Structure:** Provider mark, status square, provider details, Refresh, and Hide/Show form a dense grid row. Odd rows carry the subtle dot field.
- **Identity:** The provider mark displays the configured provider `name` and uses the deterministic provider color cycle. The same exact name appears in the Tickets source column and in brackets before each Gantt task title; generated provider numbers are not part of the interface.
- **States:** Hidden rows reduce opacity. Online is filled green, offline is pink hatched, and Cached is filled yellow. Cached explicitly means stored data whose current provider reachability is unknown; never present it as Online. Refresh failure keeps previously loaded tickets visible and presents pink alert text above a dotted separator.
- **Actions:** Refresh and Hide/Show are square outlined buttons. Hover and focus invert to yellow; loading uses disabled text plus diagonal rule hatching.

### Ticket Table
- **Structure:** The table has a 980px minimum width, sticky uppercase headers, dotted row rules, and compact cells. Source, priority, dates, status, and reference are muted until hover or focus.
- **Source Identity:** Each source cell uses the configured provider `name` directly, without a generated prefix or index.
- **Title:** Semibold sans serif with a dotted underline. Ticket labels wrap below as square dotted tags.
- **Selection:** Hover applies the chrome row tone. Keyboard focus keeps the standard panel fill and adds only the green rectangular selection, without inserting a marker or moving content.
- **Sorting:** Small inline SVG controls expose current and next order through an accessible label; sorted headers expose `aria-sort`.

### Gantt
- **Toolbar:** Month, Quarter, Year, and Today use the standard outlined control. Underline `M`, `Q`, `Y`, and the `d` in Today to expose Gantt-only keyboard shortcuts; the active zoom uses the yellow inverse state.
- **Frame:** Activating Gantt expands the main workbench to the full viewport width. The timeline retains internal horizontal scrolling where its date range exceeds that field.
- **Grid:** Month and day headers remain sticky. Dotted rules define days and rows; weekends receive the weekend tone and today receives a vertical ink marker. Implement the body with one repeating background-grid node per ticket and shared spanning day bands for weekends, month starts, and today. Do not recreate a separate cell for every ticket-day intersection; that node explosion is outside the shipped performance model.
- **Tasks:** The left task column is sticky, keyboard-focusable, defaults to 520px, and resizes from 360px to 900px by pointer or keyboard. Each task begins with the configured provider `name` in brackets. Provider-colored 18px bars show their titles with runtime-selected black or white text.
- **Selection:** A focused task uses the chrome row tone and green inset outline, matching Tickets.

### Labels And Statusline
- **Labels:** Ticket labels are transparent square tags with dotted outlines; they are metadata, not colored category pills.
- **Statusline:** A 29px minimum-height footer closes the frame with current view and read-only context. Shortcut guidance lives only in the top strip and Gantt toolbar. The active view is a green badge with dark text.

### Empty, Loading, And Error States
- **Empty:** Center content in a minimum 260px dotted field and lead with the outlined `NO TICKET DATA` marker.
- **Loading:** The first uncached Gantt build shows a `Building Gantt...` status with a 20px two-rule spinner and 1-second linear rotation. Mark the content region busy while building and clear that state when rendered; avoid additional decorative motion.
- **Error:** Use pink text and a pink structural edge or hatch, always accompanied by plain-language recovery guidance.

### Rendered View Cache
- **Behavior:** Cache rendered Tickets and Gantt markup in-browser for 15 seconds, keyed by the active sort or zoom state and the current data revision. Switching back to a valid cached view is immediate and does not show the Gantt build spinner.
- **Invalidation:** Clear rendered markup when provider visibility, provider data, or Gantt task width changes. This is interaction-performance guidance, not a visual token or a claim about provider freshness.

## Do's and Don'ts

### Do:
- **Do** preserve the single compact desktop workbench at full-screen and half-screen widths.
- **Do** use exact Monokai roles and the six-color provider cycle from the normative tokens.
- **Do** retain visible focus, green row selection, dotted links, and text or pattern cues alongside color.
- **Do** keep keyboard behavior consistent: T for Tickets, G for Gantt, R to refresh visible providers, J/K or arrows to move, and Enter to open. In Gantt only, M/Q/Y select the time window and D jumps to today.
- **Do** keep operational data dense and monospace while setting ticket titles in the restrained sans face.
- **Do** use the configured provider name consistently in Providers, Tickets, and Gantt.
- **Do** preserve the full-width Gantt frame, optimized row-grid/day-band structure, and 15-second rendered-view cache.

### Don't:
- **Don't** add rounded dashboard cards, pills, gradients, glass effects, ambient shadows, or spacious marketing-style sections.
- **Don't** turn half-screen behavior into a mobile card layout; mobile-specific design is out of scope.
- **Don't** use provider identity colors as health indicators or rely on color alone for status, errors, or focus.
- **Don't** replace Uberview, Tickets, Gantt, or Providers with clever labels, subtitles, or abstract jargon.
- **Don't** treat files under `.impeccable/review/` as shipping assets; the shipped frontend contains no raster product assets.
- **Don't** generate P01/P02-style provider identities or describe Cached data as live Online data.
- **Don't** render Gantt as one DOM cell per ticket per day; use one row-grid node per ticket plus shared day bands.
