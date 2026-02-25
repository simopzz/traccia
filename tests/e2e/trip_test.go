//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

func TestTripCreation(t *testing.T) {
	// Check if server is running
	resp, err := http.Get(baseURL)
	if err != nil {
		t.Skipf("Server not running at %s, skipping E2E test. Error: %v", baseURL, err)
	}
	resp.Body.Close()

	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		t.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("could not create page: %v", err)
	}

	// 1. Navigate to home
	if _, err = page.Goto(baseURL); err != nil {
		t.Fatalf("could not goto %s: %v", baseURL, err)
	}

	// 2. Check title
	title, err := page.Title()
	if err != nil {
		t.Fatalf("could not get title: %v", err)
	}
	expectedTitle := "Trips | traccia"
	if title != expectedTitle {
		t.Errorf("expected title %q, got %q", expectedTitle, title)
	}

	// 3. Click "Create Trip"
	err = page.GetByRole("link", playwright.PageGetByRoleOptions{
		Name: "Create Trip",
	}).Click()
	if err != nil {
		t.Fatalf("could not click 'Create Trip': %v", err)
	}

	// 4. Fill trip form
	tripName := fmt.Sprintf("E2E Trip %d", time.Now().Unix())
	err = page.Locator("#name").Fill(tripName)
	if err != nil {
		t.Fatalf("could not fill trip name: %v", err)
	}

	now := time.Now()
	startDate := now.Format("2006-01-02")
	endDate := now.Add(24 * 7 * time.Hour).Format("2006-01-02")

	err = page.Locator("#start_date").Fill(startDate)
	if err != nil {
		t.Fatalf("could not fill start date: %v", err)
	}
	err = page.Locator("#end_date").Fill(endDate)
	if err != nil {
		t.Fatalf("could not fill end date: %v", err)
	}

	// 5. Submit
	err = page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Create Trip",
	}).Click()
	if err != nil {
		t.Fatalf("could not click submit: %v", err)
	}

	// 6. Verify trip is in list
	locator := page.GetByText(tripName)
	count, err := locator.Count()
	if err != nil {
		t.Fatalf("could not count locators: %v", err)
	}
	if count == 0 {
		t.Errorf("trip %q not found in list after creation", tripName)
	}
}

func TestTripCreationValidation(t *testing.T) {
	// Check if server is running
	resp, err := http.Get(baseURL)
	if err != nil {
		t.Skipf("Server not running at %s, skipping E2E test. Error: %v", baseURL, err)
	}
	resp.Body.Close()

	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		t.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("could not create page: %v", err)
	}

	// 1. Navigate to home
	if _, err = page.Goto(baseURL); err != nil {
		t.Fatalf("could not goto %s: %v", baseURL, err)
	}

	// 2. Click "Create Trip"
	err = page.GetByRole("link", playwright.PageGetByRoleOptions{
		Name: "Create Trip",
	}).Click()
	if err != nil {
		t.Fatalf("could not click 'Create Trip': %v", err)
	}

	// 3. Fill trip form with invalid date range
	err = page.Locator("#name").Fill("Invalid Trip")
	if err != nil {
		t.Fatalf("could not fill trip name: %v", err)
	}

	now := time.Now()
	startDate := now.Format("2006-01-02")
	endDate := now.Add(-24 * time.Hour).Format("2006-01-02")

	err = page.Locator("#start_date").Fill(startDate)
	if err != nil {
		t.Fatalf("could not fill start date: %v", err)
	}
	err = page.Locator("#end_date").Fill(endDate)
	if err != nil {
		t.Fatalf("could not fill end date: %v", err)
	}

	// 4. Submit
	err = page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Create Trip",
	}).Click()
	if err != nil {
		t.Fatalf("could not click submit: %v", err)
	}

	// 5. Verify error message
	expectedError := "invalid input: end date must be on or after start date"
	errorLocator := page.Locator(".bg-rose-50")
	isVisible, err := errorLocator.IsVisible()
	if err != nil {
		t.Fatalf("could not check error visibility: %v", err)
	}
	if !isVisible {
		t.Error("error message box not visible")
	}

	errorText, err := errorLocator.InnerText()
	if err != nil {
		t.Fatalf("could not get error text: %v", err)
	}
	if errorText != expectedError {
		t.Errorf("expected error text %q, got %q", expectedError, errorText)
	}
}
