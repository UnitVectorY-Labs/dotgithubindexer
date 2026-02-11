package main

import (
	"os"
	"strings"
	"testing"
)

func TestIntegrationWorkflowParsing(t *testing.T) {
	// Create a sample workflow content
	workflowContent := `
name: Test Workflow
on: [push]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2
      
      - name: Setup Go
        uses: actions/setup-go@v5
      
      - name: Cache
        uses: actions/cache@v4.1.0 # latest cache
      
      - name: Custom action
        uses: my-org/my-action@main

  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/upload-artifact@v4
`

	// Test extraction
	uses := extractActionUses(workflowContent, "test-repo", ".github/workflows/test.yml")

	// Verify we found all the uses
	expectedCount := 6
	if len(uses) != expectedCount {
		t.Errorf("Expected to find %d uses, but found %d", expectedCount, len(uses))
	}

	for _, use := range uses {
		t.Logf("Found %s @ %s", use.Action, use.Version)
	}

	// Verify that the comment is captured for actions/checkout with hash
	checkoutWithCommentFound := false
	for _, use := range uses {
		if use.Action == "actions/checkout" && strings.Contains(use.Version, "de0fac2e4500dabe0009e67214ff5f5447ce83dd") && strings.Contains(use.Version, "v6.0.2") {
			checkoutWithCommentFound = true
			break
		}
	}
	if !checkoutWithCommentFound {
		t.Error("Expected to find actions/checkout with version comment (hash # v6.0.2)")
	}
}

func TestRealWorkflowFiles(t *testing.T) {
	// Test with actual workflow files in the repository
	workflowFiles := []string{
		".github/workflows/build-go.yml",
		".github/workflows/codeql-go.yml",
		".github/workflows/release-go-github.yml",
	}

	totalUses := 0
	for _, file := range workflowFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Logf("Skipping %s (not found): %v", file, err)
			continue
		}

		uses := extractActionUses(string(content), "dotgithubindexer", file)
		t.Logf("File %s: found %d uses", file, len(uses))
		for _, use := range uses {
			t.Logf("  - %s @ %s", use.Action, use.Version)
		}
		totalUses += len(uses)
	}

	if totalUses == 0 {
		t.Error("Expected to find at least some action uses across all workflow files")
	}
}
