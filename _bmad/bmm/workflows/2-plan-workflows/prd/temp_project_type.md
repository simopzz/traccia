## Web App Specific Requirements

### Technical Architecture
*   **Architecture Pattern:** **Server-Side Rendered (SSR)** with light interactivity via **HTMX + Alpine.js**.
    *   *Rationale:* Reduces build complexity vs full SPA. Matches the "Survival Export" ethos (HTML-first). Excellent for the "Reluctant Companion" (fast load on low-end devices).
*   **Frontend Stack:**
    *   **HTMX:** For swapping timeline segments and handling "Opportunity Filler" insertions without full reloads.
    *   **Alpine.js:** For client-side interactivity (toggling "Rhythm" details, modal interactions) without heavy hydration.
    *   **CSS:** Tailwind CSS (implied for rapid utility-first styling).

### Browser & Device Support
*   **Mobile-First Responsiveness:**
    *   The "Timeline" view must collapse gracefully from a horizontal Gantt (Desktop) to a vertical Stream (Mobile).
    *   **Critical:** "Survival Export" view must be readable on mobile *without* zooming (large touch targets, high contrast).
*   **Browser Matrix:**
    *   Support: Chrome (Desktop/Android), Safari (iOS/Desktop), Firefox.
    *   *Constraint:* No dependency on cutting-edge browser APIs that break on older iOS Safari versions (common for travelers with older phones).

### Performance Targets
*   **Load Time:** "Time to Interactive" < 1.5s on 3G networks (for the "Reluctant Companion" use case).
*   **Payload Size:** Minimal JS bundle (< 50kb gzipped) since we are avoiding React/Vue.

### Real-Time & State Strategy
*   **Sync Model (MVP):** "Polling" or Manual Refresh.
    *   If Sarah edits, Ben must reload to see changes.
    *   *UI Indicator:* Simple "Last Updated: 10 mins ago" timestamp to warn users of potential staleness.
*   **Offline Capability (PWA Lite):**
    *   While not a full native app, the web app should use a basic Service Worker to cache the *current* itinerary view for flaky connection access, falling back to the PDF Export for true offline safety.
