## Non-Functional Requirements

### Reliability & Availability
*   **Export Robustness:** The PDF Generation service must have a success rate of > 99.5%. If the real-time generation fails, the system must support an asynchronous fallback ("Email me when ready").
*   **Data Durability:** Trip data is persisted to a backend database. Users must be able to restore their session across devices using a unique "Trip ID" or "Edit Token."

### Performance
*   **Mobile Load Time:** The "Reluctant Companion" read-only view must achieve a "Time to Interactive" of < 2 seconds on 3G networks (to support usage in airports/transit).
*   **Export Latency:** Synchronous PDF generation should complete within 15 seconds. If longer, the UI must switch to the asynchronous "Email" flow.

### Security & Privacy
*   **Link Entropy:** "Shareable Links" (Read-Only) and "Edit Tokens" must use high-entropy strings (e.g., UUIDv4 or 16-char random alphanumeric) to prevent enumeration attacks.
*   **Data Minimization:** The "Survival Export" PDF must not contain any tracking pixels or external resource dependencies (images/fonts must be embedded) to ensure privacy and offline rendering.

### Accessibility
*   **Print Legibility:** The "Survival Export" PDF must adhere to minimum contrast ratios (WCAG AA) and use a minimum font size of 12pt for body text to ensure readability under poor lighting conditions (e.g., in a taxi at night).
