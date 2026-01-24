## Functional Requirements

### 1. Timeline Orchestration
*   **FR1:** Users can create a trip with a specific destination and date range.
*   **FR2:** Users can manually add events to the trip timeline with fields: Title, Location (Address + Lat/Long), Start Time, End Time, Category (e.g., Food, Activity, Lodging).
*   **FR3:** Users can drag-and-drop events to reschedule them within the timeline view.
*   **FR4:** Users can view the timeline in a linear, single-stream format that visualizes gaps between events.
*   **FR5:** The system must calculate durations using UTC deltas to support multi-timezone trips correctly.

### 2. Rhythm Guardian (Risk Detection)
*   **FR6:** The system must calculate the geographical distance between consecutive events using Haversine formula (Crow-Flies) based on user-provided Lat/Long.
*   **FR7:** The system must flag a "Transit Risk" alert if (Time Gap between events) < (Estimated Travel Time based on Distance + Buffer).
*   **FR8:** Users can manually override the "Travel Time" for any specific gap to clear a risk flag.
*   **FR9:** The system must visually indicate the "Risk Level" of a connection (e.g., Green for safe, Yellow for tight, Red for impossible).

### 3. Survival Export
*   **FR10:** Users can generate a printable "Tactical Field Guide" (PDF) for any trip.
*   **FR11:** The Export must render "Taxi Cards" for each lodging/activity, displaying the address in large, high-contrast text.
*   **FR12:** The Export must include a static QR code for each event that deep-links to Google Maps (using `geo:` or `https://maps.google.com/?q=` URI schemes).
*   **FR13:** The Export must follow a "Day-by-Day" chronological layout.

### 4. Data Management & Storage
*   **FR14:** Users can save their itinerary data (persisted via local storage or basic backend DB for MVP).
*   **FR15:** Users can "Clear/Reset" a trip to start over.
*   **FR16:** The system must validate input data types (e.g., End Time cannot be before Start Time).

### 5. Access & Sharing (MVP)
*   **FR17:** Users can generate a "Shareable Link" (hashed URL) that provides read-only access to the itinerary.
*   **FR18:** Read-only views must load without requiring user authentication.
*   **FR19:** Read-only views must be responsive and legible on mobile browser viewports (375px width).
