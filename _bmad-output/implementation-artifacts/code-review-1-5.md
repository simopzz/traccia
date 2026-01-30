**🔥 CODE REVIEW FINDINGS, simo!**

**Story:** 1-5-pinned-events-logic
**Git vs Story Discrepancies:** 2 found
**Issues Found:** 0 High, 2 Medium, 2 Low

## 🔴 CRITICAL ISSUES
*None. Good job on the core logic!*

## 🟡 MEDIUM ISSUES
- **UX/Logic Logic:** Dragging a Pinned Event to a new position in the list does NOT update its time.
  - **Details:** In `ReorderEvents`, if an event is pinned, its existing `StartTime` is preserved (`newStart = *evt.StartTime`). If a user drags a pinned event to a new visual slot, the backend will calculate the order, but the Pinned Event will force its *old* time. On page refresh, the event will snap back to its original chronological position (since `GetEvents` sorts by time).
  - **Impact:** Violates user intent. If I drag "Lunch (Pinned 13:00)" to 10:00, I expect it to move to 10:00. Instead, it stays at 13:00 and snaps back.
  - **Fix:** Either disable dragging for Pinned events (visual lock) OR update the Pinned Time to the new calculated slot time when reordered.
- **Documentation:** `Makefile` has uncommitted changes but is not listed in the Story's File List.

## 🟢 LOW ISSUES
- **Test Coverage:** `TestReorderEvents_Pinned` verifies that a pinned event stays put, but does not explicitly test the **Overlap** scenario (AC7) where an unpinned event is dragged *into* a pinned event's time slot. While the code logic handles it, a specific test case for the "Overlap/Conflict" state would be safer.
- **Documentation:** `NOTES.md` is untracked.

