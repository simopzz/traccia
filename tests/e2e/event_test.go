//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

func TestEventCreationFlow(t *testing.T) {
	pw, browser, context, page := setupBrowser(t)
	defer teardownBrowser(pw, browser, context)

	baseURL := fmt.Sprintf("http://localhost:%s", testPort)

	t.Run("Create Flight Event", func(t *testing.T) {
		// 1. Create a trip first
		if _, err := page.Goto(baseURL + "/trips/new"); err != nil {
			t.Fatalf("could not goto: %v", err)
		}

		tripName := fmt.Sprintf("Event E2E Trip %d", time.Now().Unix())
		page.Locator("#name").Fill(tripName)
		page.Locator("#start_date").Fill(time.Now().Format("2006-01-02"))
		page.Locator("#end_date").Fill(time.Now().AddDate(0, 0, 3).Format("2006-01-02"))
		page.Locator("button[type='submit']").Click()

		// Wait for redirect to detail page
		page.WaitForURL(func(url string) bool {
			return strings.Contains(url, "/trips/") && !strings.Contains(url, "/new")
		}, playwright.PageWaitForURLOptions{
			Timeout: playwright.Float(5000),
		})

		// 2. Open Add Event sheet by clicking the empty day prompt for Day 1
		page.Locator("text=+ Add event to Day 1").Click()

		// Wait for the sheet form to appear
		page.Locator("#sheet-form").WaitFor()

		// 3. Fill the common event details
		eventName := fmt.Sprintf("Morning Flight to Paris %d", time.Now().Unix())
		page.Locator("#sheet-title").Fill(eventName)
		page.Locator("#sheet-start-time").Fill("08:00")
		page.Locator("#sheet-end-time").Fill("10:00")

		// Select Flight type
		page.Locator("button[role='radio']:has-text('Flight')").Click()

		// 4. Fill Flight-specific details
		page.Locator("input[name='departure_airport']").Fill("LHR")
		page.Locator("input[name='arrival_airport']").Fill("CDG")

		// 5. Submit
		page.Locator("button:has-text('Create Event')").Click()

		// 6. Verify the event is added to the timeline
		// First wait for the sheet to close / the title to appear in the DOM
		eventTitleLocated := page.Locator(fmt.Sprintf("text=%s", eventName))
		eventTitleLocated.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3000),
		})

		count, _ := eventTitleLocated.Count()
		if count == 0 {
			t.Errorf("Expected newly created event with text %s to be visible", eventName)
		}
	})
}
