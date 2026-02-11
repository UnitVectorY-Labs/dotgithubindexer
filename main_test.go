package main

import (
	"os"
	"testing"
)

func TestExtractActionUses(t *testing.T) {
	// Read a sample workflow file
	content, err := os.ReadFile(".github/workflows/build-go.yml")
	if err != nil {
		t.Fatalf("Error reading workflow file: %v", err)
	}

	uses := extractActionUses(string(content), "dotgithubindexer", ".github/workflows/build-go.yml")

	if len(uses) == 0 {
		t.Error("Expected to find action uses, but found none")
	}

	// Check that we found the expected actions
	foundCheckout := false
	foundSetupGo := false
	foundCache := false
	foundCodecov := false

	for _, use := range uses {
		t.Logf("Found action: %s, version: %s", use.Action, use.Version)

		switch use.Action {
		case "actions/checkout":
			foundCheckout = true
			// Check that version includes the comment
			if use.Version == "" {
				t.Error("Expected version for actions/checkout, but got empty string")
			}
		case "actions/setup-go":
			foundSetupGo = true
		case "actions/cache":
			foundCache = true
		case "codecov/codecov-action":
			foundCodecov = true
		}
	}

	if !foundCheckout {
		t.Error("Expected to find actions/checkout")
	}
	if !foundSetupGo {
		t.Error("Expected to find actions/setup-go")
	}
	if !foundCache {
		t.Error("Expected to find actions/cache")
	}
	if !foundCodecov {
		t.Error("Expected to find codecov/codecov-action")
	}
}

func TestParseUsesString(t *testing.T) {
	testCases := []struct {
		name            string
		usesStr         string
		workflowContent string
		expectedAction  string
		expectedVersion string
	}{
		{
			name:    "Simple version",
			usesStr: "actions/checkout@v4",
			workflowContent: `
      - uses: actions/checkout@v4
`,
			expectedAction:  "actions/checkout",
			expectedVersion: "v4",
		},
		{
			name:    "Version with comment",
			usesStr: "actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd",
			workflowContent: `
      - uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2
`,
			expectedAction:  "actions/checkout",
			expectedVersion: "de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2",
		},
		{
			name:    "No version",
			usesStr: "actions/checkout",
			workflowContent: `
      - uses: actions/checkout
`,
			expectedAction:  "actions/checkout",
			expectedVersion: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			action, version := parseUsesString(tc.usesStr, tc.workflowContent)

			if action != tc.expectedAction {
				t.Errorf("Expected action %q, got %q", tc.expectedAction, action)
			}

			if version != tc.expectedVersion {
				t.Errorf("Expected version %q, got %q", tc.expectedVersion, version)
			}
		})
	}
}
