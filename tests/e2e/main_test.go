//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"testing"

	"github.com/playwright-community/playwright-go"
)

var baseURL = "http://localhost:8080"

func TestMain(m *testing.M) {
	// Initialize playwright
	err := playwright.Install()
	if err != nil {
		fmt.Printf("could not install playwright: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}
