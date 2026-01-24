## Success Criteria

### User Success

*   **Trust Loop (Retention):** > 40% of users who generate a "Survival Export" return to the app within 7 days of their trip ending. This is the primary proxy for "The system worked."
*   **Gap Fill Acceptance:** > 30% of "Opportunity Filler" suggestions are accepted and added to the itinerary. (Significantly higher than industry average of < 5%).
*   **Burnout Resolution:** > 60% of "Rhythm Guardian" warnings (e.g., "Too Rushed") result in a user modification to the schedule.

### Business Success

*   **Viral Coefficient (K-Factor):** > 1.2 new users referred per existing user via "Shared Itinerary" links or QR codes on printed exports.
*   **Stitching Volume:** Validation of the core problem solution is measured by the volume of external data points (events, bookings) added to a single itinerary.

### Technical Success

*   **Print Reliability:** "Survival Export" view renders correctly on mobile and desktop browsers 99% of the time. This is critical for the "Anxious Planner."
*   **Export Generation:** PDF/Static artifact generation must be robust; if this fails, the core value proposition ("Reliability") is broken.

### Measurable Outcomes

*   **Session "Depth":** Average Edit Session < 5 minutes. (Success = efficiency/speed of planning, not time spent in app).
*   **Safety Net Usage:** 50% of beta users generate a "Survival Export" before their trip.

## Product Scope

### MVP - Minimum Viable Product

*   **Core Logic:**
    *   **Orchestrator Timeline:** Linear, single-stream view. Manual entry of events.
    *   **Basic Rhythm Flags:** Time-based logic only (e.g., "Transit time > Gap time"). No complex energy scoring.
    *   **Opportunity Filler:** Google Places proxy. Detects gaps > 2 hours. Max 3 suggestions per gap.
*   **Output:**
    *   **Survival Export:** Dedicated `/print` route. CSS optimized for A4/Letter. Auto-generated "Taxi Cards" (bilingual).
*   **Data Sources:**
    *   Manual Entry only (Calendar Read moved to Growth).

### Growth Features (Post-MVP)

*   **Calendar Integration:** Google/Apple Calendar read access to auto-populate "Hard Constraints."
*   **Smart Rhythm:** Energy scoring (Mental vs Physical load) and weather integration.
*   **Collaborative Editing:** Shared write-access for groups.

### Vision (Future)

*   **The "Local Fixer":** Direct partnerships with local vendors for exclusive, bookable experiences.
*   **Group Voting:** "Tinder-style" voting on Opportunity Filler suggestions for groups.
*   **Email Parsing:** Automated ingestion of booking confirmations (TripIt-style).
