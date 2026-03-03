//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

func TestTripCreationFlow(t *testing.T) {
	// Basic setup
	pw, browser, context, page := setupBrowser(t)
	defer teardownBrowser(pw, browser, context)

	baseURL := fmt.Sprintf("http://localhost:%s", testPort)

	t.Run("Create Trip Success", func(t *testing.T) {
		// Clean up db first if needed (omitted for brevity, depending on testing strategy)

		if _, err := page.Goto(baseURL + "/trips/new"); err != nil {
			t.Fatalf("could not goto: %v", err)
		}

		// Fill out the form
		tripName := fmt.Sprintf("E2E Test Trip %d", time.Now().Unix())
		page.Locator("#name").Fill(tripName)
		page.Locator("#destination").Fill("Paris, France")

		startDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
		endDate := time.Now().AddDate(0, 0, 14).Format("2006-01-02")
		page.Locator("#start_date").Fill(startDate)
		page.Locator("#end_date").Fill(endDate)

		// Submit
		page.Locator("button[type='submit']").Click()

		// Verify redirect to trip details and content showing
		page.WaitForURL(func(url string) bool {
			return strings.Contains(url, "/trips/") || strings.Contains(url, "/trips")
		}, playwright.PageWaitForURLOptions{
			Timeout: playwright.Float(5000),
		})

		content, _ := page.Content()
		if strings.Contains(content, "Failed to create trip") {
			t.Fatalf("Server returned 500: Failed to create trip")
		}

		if !strings.Contains(page.URL(), "/trips") || strings.HasSuffix(page.URL(), "/new") {
			t.Errorf("Expected URL to be a trip detail page, got: %s", page.URL())
		}

		// Ensure success logic is working visually by checking that the title text is somewhere on the page
		// since we know we've navigated to the trips detail page successfully
		tripTitleLocated := page.Locator(fmt.Sprintf("text=%s", tripName))
		tripTitleLocated.WaitFor()

		count, _ := tripTitleLocated.Count()
		if count == 0 {
			t.Errorf("Expected heading with text %s, but wasn't found", tripName)
		}
	})

	t.Run("Create Trip Validation Error", func(t *testing.T) {
		if _, err := page.Goto(baseURL + "/trips/new"); err != nil {
			t.Fatalf("could not goto: %v", err)
		}

		// Fill out only required fields but make StartDate after EndDate
		page.Locator("#name").Fill("Invalid Trip Timeline")

		startDate := time.Now().AddDate(0, 0, 14).Format("2006-01-02")
		endDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
		page.Locator("#start_date").Fill(startDate)
		page.Locator("#end_date").Fill(endDate)

		page.Locator("button[type='submit']").Click()

		// Should stay on page and show error message
		page.Locator(".bg-rose-50").WaitFor()

		errMsg, _ := page.Locator(".bg-rose-50").TextContent()
		if errMsg == "" {
			t.Errorf("Expected an error message container to be visible")
		}
	})
}
