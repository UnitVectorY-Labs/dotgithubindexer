package main

import (
"fmt"
"os"
)

func testParsing() {
// Read the workflow file
content, err := os.ReadFile(".github/workflows/build-go.yml")
if err != nil {
fmt.Printf("Error reading file: %v\n", err)
os.Exit(1)
}

// Parse it
uses := extractActionUses(string(content), "dotgithubindexer", ".github/workflows/build-go.yml")

fmt.Printf("Found %d uses:\n", len(uses))
for _, use := range uses {
fmt.Printf("  - Action: %s\n", use.Action)
fmt.Printf("    Version: %s\n", use.Version)
fmt.Printf("    File: %s/%s\n", use.RepoName, use.FilePath)
fmt.Println()
}
}
