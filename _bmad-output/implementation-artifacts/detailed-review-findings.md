# 🔍 Detailed Issue Analysis

Here is the deep dive into the findings.

## 1. 🔴 False Claim: Documented Files are Ignored (High)

The story's **File List** claims the following files were modified/delivered:
- `internal/features/timeline/components_templ.go`
- `internal/features/timeline/home_templ.go`
- `internal/features/timeline/view_templ.go`

**The Reality:**
Your `.gitignore` file (Line 104) explicitly ignores them:
```gitignore
104| *templ.go
```
These files were **NOT** committed to the repo (verified via `git diff`). Including them in the Story documentation implies they are part of the source control history, which is false. They are build artifacts (generated code), not source code.

**Recommendation:** Remove them from the Story File List.

## 2. 🟡 Visual Truth Distortion (Medium)

**Context:**
The core value prop of Traccia is "Visualizing Time."
- 1 Hour = 64px.
- 15 Minutes = 16px.

**The Code (`components.templ`):**
```go
0016| 	if hours < 0.25 { // min 15m
0017| 		hours = 0.25
0018| 	}
0019| 	return int(hours * 64) // Returns 16px for 15m events
```
BUT, the HTML container forces a minimum height that contradicts this:
```html
0050| 	<div class="... min-h-[60px] ...">
```

**The Problem:**
- A 15-minute event calculates to **16px** height.
- The CSS forces it to **60px**.
- 60px represents `60 / 64 = 0.93 hours` (~56 minutes).
- **Result:** A 15-minute coffee break looks visually identical to a 1-hour meeting. The user loses the ability to distinguish "quick" vs "long" tasks at a glance.

**Recommendation:** Reduce `min-h` to `min-h-[40px]` (standard touch target size) or strictly follow the calculated height if it's larger than a touch minimum.

## 3. 🟡 Friction-Heavy Default Time (Medium)

**Context:**
AC 12 says: *"ideally default to the day of the last event"*.
The goal is to let the user add events sequentially (A -> B -> C).

**The Code (`view.templ`):**
```go
0011| 		if len(events) > 0 && events[len(events)-1].StartTime != nil {
0012| 			defaultStartTime = events[len(events)-1].StartTime.Format("2006-01-02T15:04")
0013| 		}
```

**The Problem:**
It defaults to the **StartTime** of the previous event.
- **Scenario:** You add "Museum" at 10:00.
- **Next Action:** You click "Add Event" for "Lunch".
- **Result:** Default is 10:00.
- **Friction:** You MUST change it to 12:00. If you forget, you have an instant overlap/conflict.

**Recommendation:** Default to `events[len-1].EndTime` (if available) OR `StartTime + 1 hour`.

---

## 4. 🟢 Timezone Fragility (Low)

**The Code (`view.templ`):**
```javascript
0060| let offset = d.getTimezoneOffset() * 60000;
0061| let localISOTime = (new Date(d - offset)).toISOString().slice(0, 16);
```

**The Problem:**
`getTimezoneOffset()` returns the offset of the **User's Browser**, not the **Trip's Destination**.
- **Scenario:** User is in New York (-5), Trip is in Tokyo (+9).
- **Result:** The "Smart Default" calculation might shift the day unexpectedly when converting to ISO strings if the offset math aligns poorly with the destination's "Day". Ideally, dates should be handled in "Trip Local Time" or strictly UTC to avoid "Reviewing my Tokyo trip from NY messes up the dates."

**Recommendation:** For now, this is acceptable for an MVP, but note it as a technical debt item.

---

### How would you like to proceed?

1.  **Fix them automatically** - I'll update the CSS, fix the Go logic for default time, and clean up the Story file.
2.  **Create action items** - Add to story Tasks/Subtasks.
