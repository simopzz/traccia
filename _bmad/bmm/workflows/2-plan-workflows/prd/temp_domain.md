## Domain-Specific Requirements

### Data Freshness & Accuracy
*   **Opening Hours Validity:** The system must account for "Point-in-Time" accuracy. A cafe open *now* might be closed when the user arrives in 2 hours. "Opportunity Filler" suggestions must validate `open_now` against the *future* arrival time, not the query time.
*   **Staleness Risk:** Cached Google Places data (stored to save costs) must include a "Last Fetched" timestamp and be invalidated after 24 hours (or per Google's Terms of Service caching allowances) to prevent sending users to permanently closed venues.

### Timezone & Localization
*   **Multi-Timezone Itineraries:** The "Orchestrator Timeline" must handle flight segments where Start Time (London) and End Time (Tokyo) are in different zones. Durations must be calculated using UTC deltas, not local clock time differences.
*   **Local Address Formats:** "Survival Export" must render addresses in the destination's local format and script (e.g., Japanese script for Tokyo addresses) to be usable by local taxi drivers.

### API & Third-Party Constraints
*   **Google Places Quotas:** The system must implement "Debouncing" on search inputs and "Aggressive Caching" (within TOS limits) for popular queries to prevent API cost explosions.
*   **Deeplinking Reliability:** Generated links to external maps (Google/Apple Maps) must use universal URI schemes (e.g., `geo:lat,long` or `https://maps.google.com/?q=...`) that work on both iOS and Android without requiring app installation.
