# Sprint Change Proposal
**Date:** 2026-02-19
**Status:** Approved
**Scope:** Minor — direct implementation by development team

---

## Section 1: Issue Summary

**Problem statement:** The project's selected component library (templui v1.5.0) relies on a Tailwind class-conflict resolver (`github.com/Oudwins/tailwind-merge-go`) to allow callers to override component default styles. This Go module is not actively maintained and does not support Tailwind v4, which traccia uses (`@import "tailwindcss"`, `@theme {}`, `@source` CSS-first configuration). The dependency was already removed from `go.mod` and `TwMerge` was patched to `strings.Join` — a concatenation-only workaround that does not resolve class conflicts.

**Context:** Discovered during Story 1.3 implementation (`feature/story-1-3`). No handler templates currently import any templui components — all handlers use raw Tailwind HTML directly — so the issue is forward-looking rather than immediately breaking. However, if left unaddressed, every future story that uses the installed components would produce unpredictable styling results when component class overrides are applied.

**Evidence:**
- `internal/components/utils/templui.go`: `TwMerge` = `strings.Join(classes, " ")` (no conflict resolution)
- `go.mod`: `tailwind-merge-go` absent (already removed)
- `static/css/input.css`: Tailwind v4 syntax confirmed (`@import "tailwindcss"`, `@theme {}`, `@source`)
- All handler templates (`event.templ`, `trip.templ`, `layout.templ`): zero imports of any installed component

---

## Section 2: Impact Analysis

**Epic Impact:** No epic scope or acceptance criteria change. All epics define *what* to build; the component tooling affects only *how* the UI is implemented. All 6 epics remain valid and unchanged.

**Story Impact:** No story rewriting required. Stories define user-facing behaviour, not implementation technology. Current in-progress Story 1.3 is unaffected (no components in use).

**Artifact Conflicts:**

| Artifact | Impact |
|---|---|
| `architecture.md` | ADR-4 title and body, Technical Constraints line, Styling Solution section, Sheet Panel reference all cite templui |
| `ux-design-specification.md` | Design System Choice section, Rationale, Implementation Approach, Customization Strategy, Component Strategy section — all built around templui |
| `epics.md` | One line in UX Design requirements cites templui |
| `internal/components/` | 17 component directories + utils/ — installed but unused |
| `.templui.json` | Configuration file for removed tooling |

**Technical Impact:**
- Removing `internal/components/` breaks nothing (zero import references in handlers)
- The utility functions `TwIf`, `IfElse`, `MergeAttributes`, `RandomID`, `ScriptURL` in `utils/templui.go` are valuable and stack-agnostic — preserved by moving to `internal/handler/ui.go`
- Adding `@plugin "daisyui"` to `input.css` requires the standalone Tailwind CLI binary to support the `@plugin` directive (Tailwind v4 feature — confirmed compatible)
- No `go.mod` changes required (daisyUI is a CSS-only plugin, not a Go dependency)

---

## Section 3: Recommended Approach

**Selected path:** Direct Adjustment (Option 1)

**Approach:** Remove templui. Adopt daisyUI v5 as the CSS component foundation. Update all planning documents to reflect the new component strategy.

**Rationale:**
- templui is unused in handlers today — removal has zero blast radius
- daisyUI v5 is purpose-built for Tailwind v4 (`@plugin "daisyui"` CSS-first syntax), actively maintained (65 components, v5.5.18 current)
- daisyUI uses semantic class names (`btn`, `card`, `badge`) rather than composable utility overrides — eliminates the tailwind-merge problem entirely
- No Go library dependency: daisyUI is pure CSS, compatible with any template language
- Same component coverage: buttons, inputs, cards, modals, drawers (sheet panels), tabs, toasts, badges, skeletons, dropdowns — all the components traccia needs
- Custom templ components (EventCard, TypeSelector, TimelineDay, etc.) are unaffected — they use Tailwind utilities directly

**Effort estimate:** Low
**Risk level:** Low (nothing currently uses the removed components)
**Timeline impact:** None — the change is a prerequisite cleanup before future stories begin using UI components

---

## Section 4: Detailed Change Proposals

### Story Changes
None required.

### PRD Changes
None required.

### Architecture Changes (`architecture.md`)

**Edit A — Line 61, Technical Constraints:**
```
OLD: - Tailwind CSS, templui component library (copy-paste ownership), Alpine.js
NEW: - Tailwind CSS, daisyUI v5 (Tailwind v4 CSS plugin, semantic classes), Alpine.js
```

**Edit B — Lines 103–106, Styling Solution:**
```
OLD:
  - templui as component library (templ + Tailwind + HTMX native, copy-paste ownership model)

NEW:
  - daisyUI v5 as CSS plugin (@plugin "daisyui" in input.css) — provides semantic component
    classes (btn, card, badge, modal, drawer, tabs, toast, skeleton, etc.) with no JS runtime
    and no Go library dependency
  - Custom templ components for traccia-specific UI (EventCard, Sheet panel, TypeSelector)
```

**Edit C — Lines 147–149, ADR-4:**
```
OLD: ADR-4: templ + HTMX + Alpine.js + templui — ✅ Keep
     ...templui provides 40+ stack-native components with copy-paste ownership.

NEW: ADR-4: templ + HTMX + Alpine.js + daisyUI — ✅ Updated
     ...daisyUI v5 provides 65 semantic CSS component classes as a Tailwind v4 plugin —
     no Go library dependency, no class-conflict concerns. Custom templ components are built
     for traccia-specific UI.
     Note: templui was removed; its tailwind-merge-go dependency did not support Tailwind v4
     and had no maintained replacement.
```

**Edit D — Line 294, Sheet Panel:**
```
OLD: Sheet slides from right (desktop) or bottom (mobile). templui Sheet component.
NEW: Sheet slides from right (desktop) or bottom (mobile). Custom templ component using
     daisyUI Drawer + Alpine.js for toggle state.
```

### UX Design Spec Changes (`ux-design-specification.md`)

**Edit A — Line 37, Project Vision:**
```
OLD: templui provides the component foundation (templ + Tailwind + HTMX native, copy-paste ownership).
NEW: daisyUI v5 provides the component foundation as a Tailwind v4 CSS plugin (semantic classes:
     btn, card, badge, modal, drawer, tabs, toast, etc.). Custom templ components are built
     for traccia-specific UI.
```

**Edit B — Lines 221–238, Design System Choice + Rationale + Implementation Approach:**
Full section replacement. Design system: `Tailwind CSS + daisyUI v5`. Rationale updated to: Tailwind v4 native, no class-conflict concerns, covers standard UI needs, themeable via CSS vars, actively maintained. Implementation approach updated to: no JS runtime for most components, native `<dialog>` / checkbox patterns for interactive ones, same theme-layer and custom component convention.

**Edit C — Line 242, Customization Strategy:**
```
OLD: (replace templui defaults)
NEW: (override daisyUI base colors)
```

**Edit D — Lines 576–654, Component Strategy section:**
"Design System Components (templui)" → "Design System Components (daisyUI v5)" with same component categories and use cases mapped to daisyUI class names. "Component Implementation Strategy" updated to reference daisyUI classes and native HTML patterns instead of templui. Custom Components subsections (TimelineDay, EventCard, TypeSelector, DragHandle, DayOverview, EmptyDayPrompt, SignalIndicator) unchanged.

### Epics Changes (`epics.md`)

**Edit — Line 90:**
```
OLD: - templui components for standard UI (forms, dialogs, toasts, navigation).
     Custom components for timeline.
NEW: - daisyUI v5 classes for standard UI (forms, dialogs, toasts, navigation).
     Custom templ components for timeline.
```

### Code Changes

| Action | Target | Detail |
|---|---|---|
| DELETE | `internal/components/` | All 17 component dirs + utils/ subdirectory |
| DELETE | `.templui.json` | templui CLI configuration |
| CREATE | `internal/handler/ui.go` | Move TwIf, IfElse, MergeAttributes, RandomID, ScriptURL from utils/templui.go; drop TwMerge |
| MODIFY | `static/css/input.css` | Add `@plugin "daisyui";` after `@import "tailwindcss";` |

---

## Section 5: Implementation Handoff

**Scope classification:** Minor — direct implementation by development team.

**Handoff:** Development team (Simo) implements directly as part of Story 1.3 wrap-up or as a standalone cleanup task before Story 1.4 begins.

**Implementation order:**
1. Apply `@plugin "daisyui"` to `input.css`, run `just dev` to verify CSS builds
2. Delete `internal/components/` and `.templui.json`
3. Create `internal/handler/ui.go` with preserved utility functions
4. Verify `just build` and `just lint` pass (no broken imports)
5. Apply document updates to `architecture.md`, `ux-design-specification.md`, `epics.md`

**Success criteria:**
- `just build` succeeds with no references to removed components
- `just lint` passes
- `static/css/app.css` regenerates with daisyUI classes present (verify with `grep "\.btn" static/css/app.css`)
- Planning documents reflect daisyUI as the component foundation
- No handler templates broken
