## Innovation & Novel Patterns

### Detected Innovation Areas

*   **Human-Centric Pacing (The "Rhythm Guardian"):**
    *   *Concept:* Shifting from "Logistics Optimization" (shortest path) to "Energy Optimization" (best flow).
    *   *Novelty:* Existing tools treat users as cargo (A to B). We treat users as biological batteries.
    *   *Mechanism:* MVP uses "Time Gaps" as a proxy for rest. Future versions will use "Intensity Tags" (Museum = High Mental Load, Beach = Low Load) to model fatigue.

*   **Constraint-Based Discovery (The "Opportunity Filler"):**
    *   *Concept:* Search without search terms. The constraints *are* the query.
    *   *Novelty:* Reverses the flow from "User asks -> App answers" to "Context exists -> App suggests."
    *   *Mechanism:* A "Slot-Fitting" algorithm that scores Google Places results not just by rating, but by "Fit" (Duration match + Category vibe).

*   **The "Anti-Cloud" Safety Net (Survival Export):**
    *   *Concept:* Digital-first planning for Analog-first execution.
    *   *Novelty:* Most apps fight to keep you online (retention). We explicitly build for being offline (reliability).
    *   *Mechanism:* The "Tactical Field Guide" format—bilingual, high-contrast, essential data only—mimicking military/expedition briefs.

### Market Context & Competitive Landscape

*   **Status Quo:** Competitors like Wanderlog and TripIt are "Container" apps—they hold what you put in.
*   **Our Wedge:** We are an "Orchestrator" app—we structure what you put in.
*   **Defense:** The "Rhythm" data (how long people *actually* need to recover after the Louvre) becomes a proprietary dataset that Google Maps cannot easily replicate without explicit user intent modeling.

### Validation Approach

*   **The "Impossible Itinerary" Test:** Feed the system known "bad" itineraries (e.g., 4 museums in one day). Success = System flags it red.
*   **The "Blind Drop" Test:** Users navigate a transit gap using *only* the printed Survival Export. Success = No phone usage.

### Risk Mitigation

*   **Google Dependency:** The "Opportunity Filler" relies entirely on the Places API.
    *   *Fallback:* If API costs spike or access is cut, revert to "Generic Prompts" (e.g., "Find a coffee shop here") without specific venue data.
*   **Subjectivity of "Rhythm":** Users burn out at different rates.
    *   *Mitigation:* MVP uses conservative defaults (Standard 15% buffer). "Smart Rhythm" (Growth) allows user calibration ("I am an Athlete" vs "I am Relaxed").
