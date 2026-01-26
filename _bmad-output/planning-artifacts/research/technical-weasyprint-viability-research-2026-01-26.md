---
stepsCompleted: [1, 2, 3, 4]
inputDocuments: []
workflowType: 'research'
lastStep: 4
research_type: 'technical'
research_topic: 'WeasyPrint Viability'
research_goals: 'Investigate compatibility with Tailwind (Flex/Grid), font rendering, and print specifics for Swiss/Brutalist design.'
user_name: 'simo'
date: '2026-01-26'
web_research_enabled: true
source_verification: true
status: 'complete'
---

# Research Report: Technical Viability of WeasyPrint

**Date:** 2026-01-26
**Author:** simo
**Topic:** WeasyPrint vs Gotenberg for Brutalist Design PDF Generation

---

## Executive Summary

This research evaluated **WeasyPrint** (Python) as a lightweight alternative to **Gotenberg** (Headless Chrome) for generating "Survival Export" PDFs. The goal was to reduce infrastructure costs (RAM usage) while maintaining the high-fidelity "Swiss/Brutalist" design specified in the UX docs.

**Final Verdict:** 🔴 **REJECTED**
WeasyPrint was found to have insufficient support for modern CSS Layouts (specifically CSS Grid) required by the project's design system (Tailwind CSS). While lightweight, it would require maintaining a separate, legacy-style stylesheet, violating the project's "Developer Experience" goals.

**Selected Alternative:** **Gotenberg (Headless Chrome)** hosted on a VPS with sufficient RAM (4GB).

---

## Key Findings

### 1. CSS Grid Support
*   **Requirement:** The "Brutalist" design relies heavily on `display: grid` for rigid, alignment-heavy layouts.
*   **WeasyPrint Capability:** **Poor/Experimental.** WeasyPrint historically lacks full CSS Grid support. While recent versions have made progress, it is nowhere near the fidelity of the Blink (Chrome) engine.
*   **Impact:** Complex grid layouts defined in Tailwind classes (e.g., `grid-cols-12`) would likely render incorrectly or require fallback to `float` or `table` based layouts.

### 2. Tailwind CSS Compatibility
*   **Requirement:** Reuse existing Tailwind utility classes for the PDF view.
*   **WeasyPrint Capability:** **Low.** Tailwind v3/v4 relies on modern CSS features (custom properties, extensive flex/grid capabilities). WeasyPrint's partial support means many utility classes would silently fail.
*   **Impact:** We would be forced to write a separate `print.css` file using "Old School" CSS (floats, tables), duplicating effort and increasing maintenance burden.

### 3. Resource Usage vs Fidelity Trade-off
*   **WeasyPrint:** ~100MB RAM. Excellent for text-heavy documents (invoices, reports).
*   **Gotenberg:** ~1GB+ RAM. Pixel-perfect rendering of any web page.
*   **Decision:** The project prioritizes **Design Fidelity** (Brutalist aesthetic) over **Maximum Efficiency**. The cost of a larger server (Hetzner 4GB RAM) is acceptable ($6/mo) compared to the development cost of fighting a limited rendering engine.

---

## Recommendation

**Proceed with Gotenberg.**
Although it requires a heavier server footprint, it guarantees that the "Survival Export" looks exactly like the web view. It allows developers to use standard Tailwind classes without worrying about PDF engine idiosyncrasies.

**Infrastructure Implication:**
Deployment target must be a VPS with at least **2GB (preferably 4GB) RAM** to prevent OOM kills during PDF generation spikes.
