# Retrospective: Epic 1 - Trip Core & Timeline Orchestration
**Date:** 2026-01-30
**Facilitator:** Bob (Scrum Master)
**Participants:** simo (Project Lead), Alice (PO), Charlie (Dev), Dana (QA), Elena (Dev)

## 1. Executive Summary
Epic 1 was a success, delivering a functional "Single-Player" MVP for Trip Management. The team established a robust technical foundation (Go/Chi/HTMX/Tailwind) and delivered the core Timeline visualization and Drag-and-Drop scheduling. However, user feedback during the review highlighted critical usability and logic gaps (Event Sizing, Missing Details, "Pinned" Logic) that must be addressed before proceeding to Epic 2's logistics intelligence features.

## 2. Delivery Metrics
*   **Stories Completed:** 4/4 (100%)
*   **Velocity:** High (MVP delivered efficiently)
*   **Quality:** 100% Test Coverage on Service Layer; 0 Critical Bugs in Production.

## 3. Key Findings

### ✅ What Went Well (Successes)
*   **Visual Intuition:** The linear Timeline View effectively replaces spreadsheets.
*   **Technical Foundation:** The `internal/features` folder structure and Service/Handler separation proved robust.
*   **Frontend Velocity:** Tailwind v4 + HTMX allowed for rapid UI iteration.
*   **Complex Logic:** The Drag-and-Drop "Ripple" update works reliably (within its current naive scope).

### 🚧 Challenges & Learnings
*   **Event Sizing:** Proportional height (64px/hr) fails for very short events (unreadable) and very long events (wasteful).
*   **Information Hiding:** Event cards are missing critical data (Start/End times), forcing unnecessary clicks.
*   **Scheduling Naivety:** The system currently treats all events as "fluid." Moving one event shifts everything. This breaks real-world constraints (e.g., fixed train departures).
*   **Date Entry Friction:** Date pickers default to "today" rather than the trip's context, causing friction.

## 4. Action Plan (Pre-Epic 2)

The team agreed to a **Preparation Phase** to address critical logic and UX gaps before starting Epic 2.

| Priority | Action Item | Description | Owner | Status |
| :--- | :--- | :--- | :--- | :--- |
| **CRITICAL** | **Implement "Pinned" Events** | Add `is_pinned` to DB. Update Reorder Logic to respect pinned anchors. This is a **Blocker** for Epic 2 Risk Logic. | Charlie | To Do |
| **CRITICAL** | **Refactor Event Card UI** | Add Start/End times. Implement min/max height constraints for readability. | Elena | To Do |
| **HIGH** | **Smart Date Defaults** | Pre-fill date pickers based on Trip context and previous events. | Elena | To Do |

## 5. Epic 2 Readiness Assessment
*   **Status:** ⚠️ **BLOCKED**
*   **Reason:** Epic 2 (Rhythm Guardian) requires the concept of "Fixed/Pinned" events to calculate transit risks accurately.
*   **Resolution:** Complete the "Pinned Events" action item immediately.

## 6. Closing Thoughts
*"We built a solid naive scheduler. Now we need to make it smart."* - Bob
