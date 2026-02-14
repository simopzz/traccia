---
stepsCompleted: [step-01-validate-prerequisites, step-02-design-epics, step-03-create-stories, step-04-final-validation]
inputDocuments:
  - _bmad-output/planning-artifacts/prd.md
  - _bmad-output/planning-artifacts/architecture.md
  - _bmad-output/planning-artifacts/ux-design-specification.md
---

# traccia - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for traccia, decomposing the requirements from the PRD, UX Design, and Architecture into implementable stories.

## Requirements Inventory

### Functional Requirements

FR1: Users can create a trip with a name, destination, and date range.
FR2: Users can view a list of their trips.
FR3: Users can edit a trip's name, destination, and date range.
FR4: Users can delete a trip and all its associated events.
FR5: Users can view a trip organized as a day-by-day timeline spanning the trip's date range.
FR6: Users can add an event to a specific day within a trip.
FR7: Users can select an event type from a closed set: Activity, Food, Lodging, Transit, Flight.
FR8: Each event type captures type-specific attributes (Flight: airline, flight number, origin/destination airports, terminal/gate; Lodging: check-in/check-out, booking reference; Transit: origin, destination, transport mode).
FR9: All event types capture shared attributes: name, location (address), start time, end time, notes.
FR10: Users can edit any attribute of an existing event.
FR11: Users can delete an event from a trip.
FR12: Users can mark an event as pinned (fixed-time, immovable) or flexible (repositionable).
FR13: The timeline displays events grouped by day in chronological order.
FR14: The timeline allocates visual weight proportional to event count per day — days with more events are visually distinguishable from days with fewer events.
FR15: Users can drill down from the day view to see full event details.
FR16: Users can reorder flexible events within a day via drag-and-drop.
FR17: Pinned events remain anchored at their position during reordering operations.
FR18: The system suggests a start time for new events based on the preceding event's end time.
FR19: Users can move an event from one day to another within the same trip.
FR20: Users can generate a print-ready PDF of their trip. (Phase 1.5)
FR21: The PDF displays events in a day-by-day chronological layout. (Phase 1.5)
FR22: Each event in the PDF shows its address. (Phase 1.5)
FR23: Each event with a location includes a QR code linking to Google Maps. (Phase 1.5)
FR24: Users can create an account and log in. (Phase 2)
FR25: Users can access their trips from multiple devices. (Phase 2)
FR26: Users can generate a shareable link for a trip. (Phase 2)
FR27: Recipients can view a trip via the shared link without creating an account. (Phase 2)
FR28: Shared views are read-only (no editing capabilities). (Phase 2)
FR29: The system estimates travel time between consecutive events based on their locations. (Phase 2)
FR30: The system flags connections where the time gap between events is shorter than the estimated travel time. (Phase 2)
FR31: Users can view weather forecasts per location for their trip dates. (Phase 2)

### NonFunctional Requirements

NFR1: Page loads and partial page updates complete in under 1 second on a connection of 10 Mbps or higher.
NFR2: Drag-and-drop reordering completes the visual update in under 100ms — no perceptible delay between drop and re-rendered state.
NFR3: PDF generation (Phase 1.5) may take up to 10 seconds; acceptable given it's a one-time export action.
NFR4: Trip and event data is durably persisted. No data loss on normal server restarts, verified by restart-and-query test.
NFR5: Event reordering operations are atomic — a failed reorder does not leave events in an inconsistent state.

### Additional Requirements

**From Architecture:**
- Existing codebase — not greenfield. Layered architecture already established (handler → service → domain ← repository)
- Event type modeling: Base + Detail Tables (base `events` table + `flight_details`, `lodging_details`, `transit_details` tables). Activity and Food use base table only.
- Position algorithm: Gap-based integers (gap of 1000). Renumber entire day when gaps exhaust.
- Day scoping: Explicit `event_date DATE` column derived from `start_time` by service layer
- HTMX swap strategy: Day-level swaps after all event mutations. Server returns full day HTML. No partial card swaps.
- Drag-and-drop: SortableJS with Alpine.js event binding. Optimistic UI for < 100ms, server reconciliation via day swap.
- Form morphing: Alpine.js `x-show` for TypeSelector — all 5 form variants in DOM, instant toggle
- Event creation: HTMX-fetched Sheet panel with context defaults (day, start time, end time)
- Validation: Service layer owns all business rules; DB constraints as safety net
- Migration: Replace initial migration with target schema (detail tables, event_date, gap-based positions, notes, Flight category). No data migration — no production data.
- sqlc generated code isolated in `repository/sqlcgen/` (separate Go package)
- Wide JOIN read strategy: `GetEventWithDetails` LEFT JOINs all 3 detail tables. `event_store.go` maps to domain types.
- Transactional event creation: `event_store.go` owns the transaction, detail stores participate via `sqlcgen.DBTX`
- Cross-day move: OOB swap pattern — target day as primary response, source day via `hx-swap-oob`
- `PDFExporter` interface defined in domain layer now; implementation deferred to Phase 1.5
- Tailwind CSS integrated with dev workflow: `just dev` runs air + Tailwind watcher concurrently
- No auth, no CSRF, no CI/CD in MVP. Auth stubs ready for Phase 2.

**From UX Design:**
- Desktop-first design, single breakpoint at 768px. Mobile-first CSS with `md:` prefix for desktop enhancements.
- Swiss Bordered Cards direction: 1px solid border, 2px hard shadow, type icons with colored backgrounds
- Timeline spine as vertical connector between events
- Fixed card height for MVP. Duration communicated via written time range.
- Lock icon for pinned events (replaces text label). DragHandle hidden on pinned events.
- TypeSelector: Compact horizontal icon bar (5 icons), `role="radiogroup"`, Alpine.js form morphing
- EventCard: Progressive disclosure via expand/collapse (Collapsible). Context menu (Dropdown) for actions.
- Sheet panel for event creation (right on desktop, bottom on mobile). Inline editing for modifications.
- EmptyDayPrompt for days with no events. Empty trip list with "Plan your first trip" CTA.
- templui components for standard UI (forms, dialogs, toasts, navigation). Custom components for timeline.
- Day tabs for switching. Breadcrumb navigation (Trip List → Trip → Day).
- Smart defaults: start time from preceding event, end time from type-based duration, day from context
- Undo toast (8s window) for delete actions. No "Are you sure?" for single events. Dialog for trip deletion.
- Field-level validation with rose border + inline error message. Server errors via toast.
- WCAG 2.1 AA compliance: contrast ratios, keyboard navigation, screen reader support, 44x44px touch targets
- Semantic HTML: timeline as `<ol>`, events as `<li>`, days as `<section>` with `aria-label`
- `aria-live="polite"` for HTMX updates. Focus management after swaps.
- 200% text zoom support. `prefers-reduced-motion` respected.
- Tabular numerals (`font-variant-numeric: tabular-nums`) mandatory on all time displays
- Monospace treatment for structured data (flight numbers, booking references, addresses)
- DayOverview component for trip-level scanning: event count, time span, type composition icons

### FR Coverage Map

| FR | Epic | Description |
|---|---|---|
| FR1 | Epic 1 | Create trip with name, destination, date range |
| FR2 | Epic 1 | View trip list |
| FR3 | Epic 1 | Edit trip details |
| FR4 | Epic 1 | Delete trip and associated events |
| FR5 | Epic 1 | View trip as day-by-day timeline |
| FR6 | Epic 1 | Add event to a specific day |
| FR7 | Epic 1 | Select event type from closed set |
| FR8 | Epic 1 | Type-specific attributes per event type |
| FR9 | Epic 1 | Shared attributes across all event types |
| FR10 | Epic 1 | Edit any event attribute |
| FR11 | Epic 1 | Delete an event |
| FR12 | Epic 1 | Mark event as pinned or flexible |
| FR13 | Epic 1 | Timeline displays events grouped by day |
| FR14 | Epic 2 | Visual day density distinction |
| FR15 | Epic 1 | Drill down to full event details |
| FR16 | Epic 2 | Drag-and-drop reordering |
| FR17 | Epic 2 | Pinned event anchoring during reorder |
| FR18 | Epic 1 | Auto-suggested start times |
| FR19 | Epic 2 | Cross-day event moves |
| FR20 | Epic 3 | Generate print-ready PDF |
| FR21 | Epic 3 | PDF day-by-day chronological layout |
| FR22 | Epic 3 | PDF shows event addresses |
| FR23 | Epic 3 | PDF QR codes to Google Maps |
| FR24 | Epic 4 | User account creation and login |
| FR25 | Epic 4 | Multi-device trip access |
| FR26 | Epic 5 | Generate shareable trip link |
| FR27 | Epic 5 | View trip via shared link without account |
| FR28 | Epic 5 | Shared views are read-only |
| FR29 | Epic 6 | Travel time estimation between events |
| FR30 | Epic 6 | Flag impossible connections |
| FR31 | Epic 6 | Weather forecasts per location |

## Epic List

### Epic 1: Trip & Event Management
Users can create trips, add all 5 typed events with type-specific attributes, edit/delete events, mark events as pinned or flexible, and view a day-by-day timeline with drill-down and auto-suggested start times.
**FRs covered:** FR1-FR13, FR15, FR18

### Story 1.1: Trip CRUD & Timeline Shell

As a traveler,
I want to create, view, edit, and delete trips and see them as a day-by-day timeline,
So that I can organize my travel plans with a clear structure from first day to last.

**Acceptance Criteria:**

**Given** a user is on the trip list page
**When** they click "Create Trip" and fill in name, destination, and date range
**Then** a new trip is created and appears in the trip list
**And** the trip timeline page shows one section per day spanning the date range

**Given** a user has created trips
**When** they visit the trip list page
**Then** all trips are displayed with name, destination, and dates

**Given** a user is viewing a trip
**When** they edit the trip's name, destination, or date range
**Then** the changes are persisted and the timeline adjusts to the new date range

**Given** a user is viewing a trip
**When** they delete the trip and confirm via the confirmation dialog
**Then** the trip and all associated events are permanently removed
**And** the user is returned to the trip list

**Given** a user is viewing a trip timeline
**When** no events exist for a day
**Then** an EmptyDayPrompt is displayed with an "Add Event" call-to-action

### Story 1.2: Activity & Food Event Creation

As a traveler,
I want to add Activity and Food events to specific days in my trip with smart time defaults,
So that I can build the shape of each day quickly without filling every field from scratch.

**Acceptance Criteria:**

**Given** a user is viewing a trip day
**When** they click "Add Event"
**Then** a Sheet panel opens with a TypeSelector (5 type icons) and form fields

**Given** the Sheet panel is open
**When** the user selects Activity or Food from the TypeSelector
**Then** the form displays shared fields: name, location/address, start time, end time, notes, and pinned toggle

**Given** the user is adding an event to a day with existing events
**When** the form loads
**Then** start time is pre-filled from the preceding event's end time
**And** end time is calculated from type-based duration (Activity ~2hr, Food ~1.5hr)

**Given** the user is adding the first event of a day
**When** the form loads
**Then** start time defaults to 9:00 AM

**Given** the user submits a valid event form
**When** the server processes the request
**Then** the event appears in the timeline grouped by day in chronological order
**And** the event card shows type icon, name, time range, and location
**And** the day's HTML is replaced via HTMX swap

**Given** the user toggles the pinned switch during creation
**When** the event is saved
**Then** the event displays a lock icon in the card header

**Given** the user submits a form with missing required fields
**When** validation fails
**Then** field-level error messages appear with rose borders on invalid fields
**And** all other entered values are preserved

### Story 1.3: Event Edit, Delete & Detail View

As a traveler,
I want to edit event details, delete events, and expand cards to see full information,
So that I can refine my plan and access all details without leaving the timeline.

**Acceptance Criteria:**

**Given** a user clicks/taps on an event card in the timeline
**When** the card expands
**Then** full event details are shown (all shared attributes including notes)
**And** the expansion uses progressive disclosure (Collapsible)

**Given** a user is viewing an expanded event card
**When** they edit a field (name, time, location, notes, pinned status)
**Then** the change is saved and the day's timeline updates via HTMX swap

**Given** a user deletes an event
**When** the deletion is processed
**Then** the event is removed from the timeline
**And** a toast appears with "Event removed. [Undo]" for 8 seconds
**And** the day's HTML updates via HTMX swap

**Given** the undo toast is visible
**When** the user clicks "Undo" within 8 seconds
**Then** the event is restored to its original position in the timeline

**Given** a user edits an event's start time
**When** the change is saved
**Then** the event_date is recalculated from the new start time by the service layer

### Story 1.4: Flight Events

As a traveler,
I want to add Flight events with airline, flight number, airports, terminals, and gates,
So that I can capture all flight details in my trip timeline with the right level of specificity.

**Acceptance Criteria:**

**Given** the user selects Flight from the TypeSelector
**When** the form morphs
**Then** additional fields appear: airline, flight number, departure airport, arrival airport, departure terminal, arrival terminal, departure gate, arrival gate, booking reference

**Given** the user submits a valid Flight event
**When** the event is saved
**Then** the flight_details are persisted in the flight_details table within the same transaction as the base event
**And** the FlightCardContent displays flight-specific metadata (airline, flight number, airports)

**Given** a user edits a Flight event
**When** they modify flight-specific fields
**Then** the flight_details are updated and the card reflects the changes

**Given** a user deletes a Flight event
**When** the event is deleted
**Then** both the base event and flight_details are removed (CASCADE)

### Story 1.5: Lodging Events

As a traveler,
I want to add Lodging events with check-in/check-out times and booking reference,
So that I can track accommodation details alongside my daily activities.

**Acceptance Criteria:**

**Given** the user selects Lodging from the TypeSelector
**When** the form morphs
**Then** additional fields appear: check-in time, check-out time, booking reference

**Given** the user submits a valid Lodging event
**When** the event is saved
**Then** the lodging_details are persisted in the lodging_details table within the same transaction
**And** the LodgingCardContent displays lodging-specific metadata (check-in/out, booking ref)

**Given** a user edits a Lodging event
**When** they modify lodging-specific fields
**Then** the lodging_details are updated and the card reflects the changes

**Given** a user deletes a Lodging event
**When** the event is deleted
**Then** both the base event and lodging_details are removed (CASCADE)

### Story 1.6: Transit Events

As a traveler,
I want to add Transit events with origin, destination, and transport mode,
So that I can plan how I get between places and see transit legs in my timeline.

**Acceptance Criteria:**

**Given** the user selects Transit from the TypeSelector
**When** the form morphs
**Then** additional fields appear: origin, destination, transport mode

**Given** the user submits a valid Transit event
**When** the event is saved
**Then** the transit_details are persisted in the transit_details table within the same transaction
**And** the TransitCardContent displays transit-specific metadata (origin → destination, mode)

**Given** a user edits a Transit event
**When** they modify transit-specific fields
**Then** the transit_details are updated and the card reflects the changes

**Given** a user deletes a Transit event
**When** the event is deleted
**Then** both the base event and transit_details are removed (CASCADE)

---

## Epic 2: Timeline Interaction

Users can reorder events via drag-and-drop, see pinned events anchor in place, move events between days, and visually distinguish packed days from light ones.

### Story 2.1: Drag-and-Drop Reordering

As a traveler,
I want to drag flexible events to reorder them within a day while pinned events stay anchored,
So that I can reshape my schedule intuitively without breaking fixed commitments.

**Acceptance Criteria:**

**Given** a day has multiple flexible events
**When** the user drags a flexible event to a new position
**Then** the visual update completes in under 100ms (optimistic UI via SortableJS + Alpine.js)
**And** an HTMX PUT request sends the new event ID order to the server

**Given** the server receives a reorder request
**When** it validates and assigns new gap-based positions (1000, 2000, 3000...)
**Then** the server returns the full day HTML
**And** HTMX swaps `#day-{date}` to reconcile client and server state

**Given** a day contains pinned events
**When** the user attempts to drag a pinned event
**Then** the drag is rejected — pinned events show a lock icon and no DragHandle

**Given** a reorder payload where pinned events have changed position
**When** the server validates the request
**Then** the server rejects the reorder and returns the correct order (reverting the optimistic move)

**Given** a day with events at positions where gaps have exhausted (gap < 1)
**When** a reorder is performed
**Then** the server renumbers the entire day's positions in a single atomic transaction (NFR5)

**Given** a reorder request fails due to a server error
**When** the HTMX response arrives
**Then** the full day HTML swap restores the correct server-side order
**And** an error toast is displayed

### Story 2.2: Cross-Day Event Moves

As a traveler,
I want to move an event from one day to another within my trip,
So that I can rebalance packed days by shifting events to lighter ones.

**Acceptance Criteria:**

**Given** a user drags a flexible event to a different day's drop zone
**When** the move is processed
**Then** the event is removed from the source day and appended to the target day at position max + 1000

**Given** a cross-day move is completed
**When** the server responds
**Then** the target day HTML is returned as the primary HTMX response
**And** the source day HTML is returned via `hx-swap-oob="outerHTML:#day-{source-date}"`
**And** both days update visually

**Given** a user attempts to move a pinned event to a different day
**When** the drag is initiated
**Then** the drag is rejected — pinned events cannot be moved between days

**Given** a cross-day move is performed
**When** the event's start time places it on the target date
**Then** the service layer updates `event_date` to match the target day

### Story 2.3: Day Density & Overview

As a traveler,
I want to see at a glance which days are packed and which are light,
So that I can spot imbalances and plan my trip with a realistic pace.

**Acceptance Criteria:**

**Given** a trip has days with varying event counts
**When** the user views the trip timeline
**Then** days with more events are visually distinguishable from days with fewer events (visual weight proportional to event count)

**Given** the trip timeline is displayed
**When** the user views the DayOverview for each day
**Then** each day shows: day label, event count, time span (earliest start → latest end), and type composition icons (miniature type icons)

**Given** the user clicks on a DayOverview
**When** the day expands
**Then** the full TimelineDay view is shown with all EventCards

**Given** a day has zero events
**When** it is displayed in the trip overview
**Then** it appears muted/empty with the EmptyDayPrompt visible on expansion

---

## Epic 3: Survival Export (Phase 1.5)

Users can generate a print-ready PDF with day-by-day layout, addresses, and QR codes linking to Google Maps.

### Story 3.1: PDF Generation & Day-by-Day Layout

As a traveler,
I want to generate a print-ready PDF of my trip with all events organized by day,
So that I have a physical backup of my plan with addresses I can reference offline.

**Acceptance Criteria:**

**Given** a user is viewing a trip with events
**When** they click "Generate Survival Export"
**Then** a PDF is generated within 10 seconds (NFR3)
**And** the PDF is available for download/save

**Given** the PDF is generated
**When** the user opens it
**Then** events are displayed in a day-by-day chronological layout matching the trip's date range

**Given** an event has a location/address
**When** it appears in the PDF
**Then** the address is clearly displayed alongside the event name and time

**Given** a day has no events
**When** it appears in the PDF
**Then** the day is either omitted or shown as a minimal placeholder (no empty page waste)

**Given** the trip has events of different types
**When** the PDF is rendered
**Then** event type is identifiable (type label or icon) and type-specific metadata is included (flight number, booking reference, etc.)

### Story 3.2: QR Codes & Print Optimization

As a traveler,
I want QR codes linking to Google Maps for each location and a print-optimized layout,
So that I can navigate to any stop by scanning from a printout even without internet.

**Acceptance Criteria:**

**Given** an event has a location/address
**When** it appears in the PDF
**Then** a QR code linking to Google Maps for that address is displayed alongside the event
**And** the QR code is minimum 2cm x 2cm for reliable scanning from paper

**Given** an event has no location
**When** it appears in the PDF
**Then** no QR code is rendered for that event

**Given** the PDF is intended for printing
**When** it is rendered
**Then** the layout uses black-on-white color scheme, removes background fills, and maximizes contrast
**And** typography is optimized for paper legibility (sufficient font sizes, clear hierarchy)

**Given** a trip spans multiple days with many events
**When** the PDF is generated
**Then** page breaks fall between days (not mid-day) where possible
**And** the layout fits A4 paper dimensions

---

## Epic 4: Authentication & Accounts (Phase 2)

Users can create accounts, log in, and access trips from multiple devices.

### Story 4.1: User Registration & Login

As a traveler,
I want to create an account and log in,
So that my trip data is secure and tied to my identity.

**Acceptance Criteria:**

**Given** a visitor is not logged in
**When** they visit the app
**Then** they are redirected to a login/registration page

**Given** a visitor is on the registration page
**When** they create an account via Supabase authentication
**Then** their account is created and they are logged in and redirected to the trip list

**Given** a registered user is on the login page
**When** they enter valid credentials
**Then** they are authenticated and redirected to the trip list

**Given** a registered user enters invalid credentials
**When** they attempt to log in
**Then** an error message is displayed and they remain on the login page

**Given** authentication is enabled
**When** any request hits a protected route
**Then** the auth middleware validates the session
**And** CSRF protection is enforced on all mutation requests

**Given** a logged-in user
**When** they log out
**Then** their session is terminated and they are redirected to the login page

### Story 4.2: Multi-Device Trip Access

As a traveler,
I want to access my trips from any device,
So that I can plan on my laptop and check my itinerary on my phone.

**Acceptance Criteria:**

**Given** a user is authenticated
**When** they view the trip list
**Then** only trips belonging to their `user_id` are displayed

**Given** a user creates a trip
**When** the trip is saved
**Then** the trip's `user_id` is set to the authenticated user's ID

**Given** a user logs in from a different device
**When** they view the trip list
**Then** they see the same trips as on their original device

**Given** a user attempts to access another user's trip by URL
**When** the server processes the request
**Then** a 404 is returned (trip not found for this user)

**Given** existing trips created before authentication was enabled
**When** the migration runs
**Then** orphaned trips (null `user_id`) are handled gracefully per migration strategy

---

## Epic 5: Trip Sharing (Phase 2)

Users can generate shareable read-only links for companions who don't need an account.

### Story 5.1: Shareable Link Generation

As a traveler,
I want to generate a shareable link for my trip,
So that I can send it to travel companions without requiring them to create an account.

**Acceptance Criteria:**

**Given** a user is viewing their trip
**When** they click "Share Trip"
**Then** a unique shareable link is generated using a high-entropy token
**And** the link is displayed with a copy-to-clipboard button

**Given** a shareable link has been generated for a trip
**When** the user views the share option again
**Then** the existing link is shown (not regenerated)

**Given** a user wants to revoke a shared link
**When** they regenerate the link
**Then** the old token is invalidated and a new link is created
**And** the old link no longer provides access

**Given** the shareable link token
**When** it is examined
**Then** it does not expose the trip's primary key ID or any internal identifiers

### Story 5.2: Read-Only Shared View

As a travel companion,
I want to view a trip via a shared link without creating an account,
So that I can check the plan, find addresses, and see what's happening each day.

**Acceptance Criteria:**

**Given** a recipient opens a valid shareable link
**When** the page loads
**Then** the trip timeline is displayed with all days and events in read-only mode
**And** no login or signup is required

**Given** the shared view is displayed
**When** the recipient interacts with it
**Then** no editing controls are available (no edit, delete, drag-and-drop, or add event buttons)
**And** event cards are expandable for detail viewing only

**Given** a recipient opens an invalid or revoked shareable link
**When** the page loads
**Then** a friendly error message is shown ("This link is no longer valid" or similar)

**Given** the shared view is accessed on mobile
**When** the page renders
**Then** the responsive layout works identically to the authenticated mobile view minus edit controls

---

## Epic 6: Logistics Intelligence (Phase 2)

System estimates travel times, flags impossible connections, and shows weather averages with packing suggestions per location.

### Story 6.1: Travel Time Estimation

As a traveler,
I want to see estimated travel time between consecutive events,
So that I know how long it takes to get from one stop to the next and can plan realistic transitions.

**Acceptance Criteria:**

**Given** two consecutive events in a day both have locations
**When** the timeline is displayed
**Then** an estimated travel time is shown between the two event cards
**And** the estimate is sourced from a routing API (Geoapify)

**Given** an event has no location
**When** travel time is calculated
**Then** no estimate is shown for connections involving that event

**Given** events are reordered or moved between days
**When** the timeline updates
**Then** travel time estimates recalculate for the affected connections

**Given** the routing API is unavailable or returns an error
**When** travel time is requested
**Then** no estimate is shown (graceful degradation)
**And** no error is surfaced to the user

### Story 6.2: Connection Risk Flags

As a traveler,
I want to be warned when the gap between events is shorter than the travel time,
So that I can spot impossible connections before they ruin my day.

**Acceptance Criteria:**

**Given** the time gap between two consecutive events is shorter than the estimated travel time
**When** the timeline is displayed
**Then** a SignalIndicator appears between the two events with `signal-risk` styling (rose icon + color)
**And** a travel-aware message is shown (e.g., "That's a sprint between Shibuya and Asakusa — 15 min gap, 40 min travel")

**Given** the time gap is tight but feasible (gap within 10 minutes of travel time)
**When** the timeline is displayed
**Then** a SignalIndicator appears with `signal-warn` styling (amber icon + color)
**And** a contextual message is shown (e.g., "Tight connection — you'll need to move fast")

**Given** the time gap comfortably exceeds the estimated travel time
**When** the timeline is displayed
**Then** no SignalIndicator is shown (safe connections are silent)

**Given** a SignalIndicator is displayed
**When** the user examines it
**Then** the indicator uses icon + color (never color alone) for accessibility
**And** the message references specific locations from the events

### Story 6.3: Weather Averages & Packing Context

As a traveler,
I want to see historical weather averages for each location on my trip dates,
So that I can plan activities appropriate to the conditions and pack accordingly.

**Acceptance Criteria:**

**Given** a trip has events with locations and dates
**When** the user views the trip timeline
**Then** historical weather averages are displayed per day or per event location
**And** the data is sourced from a weather API (Visual Crossing)

**Given** weather averages are available for a location and date
**When** the data is displayed
**Then** average temperature range and typical conditions (sunny, rainy, etc.) are shown
**And** packing suggestions are derived from the conditions (e.g., "Expect rain — bring layers and an umbrella")

**Given** multiple locations in a trip span different climates
**When** the user views the trip overview
**Then** packing suggestions are aggregated across all locations to cover the full trip

**Given** the weather API is unavailable or returns an error
**When** weather data is requested
**Then** the timeline displays without weather information (graceful degradation)
**And** no error is surfaced to the user
