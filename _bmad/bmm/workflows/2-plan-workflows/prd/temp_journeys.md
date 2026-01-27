## User Journeys

### 1. Sarah: The "From Panic to Peace" Journey (Primary User - Reliability Focus)

**Backstory:** Sarah (34, Project Manager) is organizing a 10-day trip to Japan for her family. She is terrified of "The Domino Effect"—one missed train ruining the whole trip. She has 14 confirmation emails and a messy Google Doc.

*   **Opening Scene (The Anxiety):** It's Tuesday night, 11 PM. Sarah is staring at her spreadsheet. She realizes she has a dinner reservation in Shinjuku at 7 PM but her teamLab Planets ticket ends at 6 PM. "Can I make it across Tokyo in rush hour?" She feels the familiar chest-tightening panic of uncertainty.
*   **Rising Action (The Intervention):** She opens traccia and manually enters her events.
    *   *System Action:* As she enters the teamLab slot, the **Rhythm Guardian** immediately flashes a "Yellow Alert": *"Transit risk: Shinjuku is 45 mins away. You have 15 mins buffer. Recommended: Move dinner to 8 PM."*
    *   *Emotional Shift:* Validation. "Thank god I didn't find this out *on the train*." She moves the dinner. The timeline turns green.
*   **Climax (The Safety Net):** Two days before the flight. She hits "Generate Survival Export." The system churns for 5 seconds and spits out a bilingual PDF. She sees the "Taxi Card" for her Ryokan in Kyoto—big, bold Japanese text. She prints two copies. One goes in her carry-on, one in her husband's bag. She finally closes her browser tabs.
*   **Resolution (The Salvation):** Day 4 in Tokyo. Her phone overheats and dies while trying to find the hotel. Her husband pulls out the printout, shows the Taxi Card to a driver. They arrive in 15 minutes. The system worked when the tech failed.

**Requirements Revealed:**
*   **Visual Feedback:** Clear, color-coded "Risk Alerts" (Green/Yellow/Red) in the timeline.
*   **Transit Logic:** Buffer calculation must account for city-specific transit realities (or at least reasonable averages).
*   **Export Formatting:** The PDF must be legible at a glance (large fonts for addresses) and bilingual.

### 2. David: The "Serendipity Engineered" Journey (Primary User - Efficiency Focus)

**Backstory:** David (28, Solo Traveler) is in Berlin for 3 days. He hates "tourist traps" and planning every minute, but hates "wasting time" even more. He wants to find cool, niche spots without spending hours researching.

*   **Opening Scene (The Void):** It's 2 PM on a rainy Tuesday in Berlin. David just finished a museum visit early. He has a dinner reservation at 6 PM, but the next 4 hours are blank. He's standing on a street corner, cold, not wanting to scroll TripAdvisor.
*   **Rising Action (The Suggestion):** He opens the app. He sees the "4 Hour Gap" on his timeline. He taps **"Fill Opportunity."**
    *   *System Action:* The system sees: Location (Mitte), Weather (Rain), Time (Afternoon). It filters Google Places for "Indoor" + "High Rating" + "Coffee/Bookstore".
    *   *The Magic:* It suggests: "Do You Read Me?!" (Iconic bookstore) -> "Five Elephant" (Coffee). Total time: 2.5 hours. 10 min walk.
*   **Climax (The Execution):** He accepts. The timeline snaps shut. He walks to the bookstore, spends an hour browsing design magazines, then gets coffee. It feels like a perfect, "local" afternoon.
*   **Resolution (The Value):** He arrives at dinner refreshed, feeling like he "won" the afternoon. He didn't search; he just executed.

**Requirements Revealed:**
*   **Contextual Querying:** The "Opportunity Filler" must pass specific constraints (Weather, Open Now, Category) to the Google Places API.
*   **One-Tap Add:** Adding a suggestion must be frictionless; no complex forms.
*   **Geo-Fencing:** Suggestions must be strictly within a walkable/reasonable radius of the previous/next event.

### 3. Alex: The System Monitor (Admin/Dev - Operational Focus)

**Backstory:** Alex is the solo dev/admin running traccia. His nightmare is a Google Maps bill spike or a broken PDF generator leaving users stranded.

*   **Opening Scene (The Spike):** Alex gets a Slack alert: "Google Places API Quota at 80%." It's only the 15th of the month.
*   **Rising Action (The Investigation):** He logs into the Admin Dashboard. He sees a spike in "Opportunity Filler" requests from a specific user cohort (maybe a bot?).
*   **Climax (The Control):** He toggles a "Cache Aggressively" switch for the Places API (serving cached results for generic queries like "Coffee near Eiffel Tower" instead of fresh fetches).
*   **Resolution:** API usage flattens. Cost is contained. Service remains up for paying users.

**Requirements Revealed:**
*   **Quota Management:** Backend must track API calls per user/session to prevent abuse.
*   **Caching Layer:** Essential for Google Places data to manage costs (store "Place Details" for 24h where permitted).
*   **Health Dashboard:** Simple view of "PDF Generation Success Rate" and "API Latency."

### 4. Ben: The Reluctant Companion (Secondary User - Passive)

**Backstory:** Ben is Sarah's husband. He loves the trip but hates the planning. He doesn't want to install an app.

*   **Opening Scene:** They land in Tokyo. Sarah is fumbling with SIM cards. Ben just wants to know "Where are we going?"
*   **Rising Action:** Sarah sends him a link: `traccia.app/share/tokyo-trip-123`. He opens it.
*   **Climax:** It opens in his browser (no login required). It's a read-only, high-contrast timeline. "Hotel Check-in: 3 PM. Address: [Click to Copy]."
*   **Resolution:** He is informed and autonomous without being "onboarded."

**Requirements Revealed:**
*   **Public/Shareable Views:** Unique, unguessable URLs for read-only access.
*   **No-Login Access:** Consuming the itinerary must not require an account.
*   **Mobile Web Optimization:** The read-only view must be lightweight and fast on spotty airport Wi-Fi.

### Journey Requirements Summary

These journeys highlight four critical capability areas:
1.  **The Engine (Logic):** Transit calculation logic (Sarah) + Contextual Search filters (David).
2.  **The Artifact (Output):** High-fidelity, bilingual PDF generation (Sarah).
3.  **The Guardrails (Ops):** API Quota management and Caching strategies (Alex).
4.  **The Window (Access):** Frictionless, token-based public access for read-only views (Ben).
