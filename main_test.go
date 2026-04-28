package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeDotfilePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "root dotfile", input: ".gitignore", want: ".gitignore"},
		{name: "trim current directory", input: "./.gitignore", want: ".gitignore"},
		{name: "nested path", input: ".config/example.yml", want: ".config/example.yml"},
		{name: "empty", input: "", wantErr: true},
		{name: "absolute", input: "/tmp/.env", wantErr: true},
		{name: "parent traversal", input: "../.env", wantErr: true},
		{name: "github path", input: ".github/dependabot.yml", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeDotfilePath(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeDotfilePath(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("normalizeDotfilePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLoadDotfilesConfigSortsAndDeduplicates(t *testing.T) {
	t.Parallel()

	dbPath := t.TempDir()
	configPath := filepath.Join(dbPath, "dotfiles.yaml")
	content := "dotfiles:\n  - .gitignore\n  - .config/example.yml\n  - ./.gitignore\n"
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	config, err := loadDotfilesConfig(dbPath)
	if err != nil {
		t.Fatalf("loadDotfilesConfig returned error: %v", err)
	}
	if config == nil {
		t.Fatal("expected config to be loaded")
	}

	want := []string{".config/example.yml", ".gitignore"}
	if len(config.Dotfiles) != len(want) {
		t.Fatalf("got %d dotfiles, want %d", len(config.Dotfiles), len(want))
	}
	for i := range want {
		if config.Dotfiles[i] != want[i] {
			t.Fatalf("config.Dotfiles[%d] = %q, want %q", i, config.Dotfiles[i], want[i])
		}
	}
}

func TestGenerateDotfileReadmeFilesWithoutCategories(t *testing.T) {
	t.Parallel()

	dbPath := t.TempDir()
	if err := updateDotfileIndex(dbPath, ".gitignore", "repo-a", "hash-one", "Default"); err != nil {
		t.Fatalf("updateDotfileIndex returned error: %v", err)
	}
	if err := updateDotfileIndex(dbPath, ".gitignore", "repo-b", "hash-one", "Default"); err != nil {
		t.Fatalf("updateDotfileIndex returned error: %v", err)
	}

	if err := generateDotfileReadmeFiles(dbPath, "UnitVectorY-Labs"); err != nil {
		t.Fatalf("generateDotfileReadmeFiles returned error: %v", err)
	}

	readmePath := filepath.Join(dbPath, "dotfiles", ".gitignore", "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("failed to read generated README: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "## [hash-one](hash-one)") {
		t.Fatalf("expected hash section in README, got:\n%s", content)
	}
	if strings.Contains(content, "## Default") {
		t.Fatalf("did not expect category section in README, got:\n%s", content)
	}
}

func TestGenerateDBSummaryIncludesDotfileCategories(t *testing.T) {
	t.Parallel()

	dbPath := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dbPath, "workflows", "build.yml"), 0755); err != nil {
		t.Fatalf("failed to create workflows directory: %v", err)
	}
	if err := updateActionIndex(dbPath, "build.yml", "repo-a", "workflow-hash"); err != nil {
		t.Fatalf("updateActionIndex returned error: %v", err)
	}
	if err := updateDotfileIndex(dbPath, ".gitignore", "repo-a", "dotfile-hash", "Base"); err != nil {
		t.Fatalf("updateDotfileIndex returned error: %v", err)
	}

	if err := generateDBSummary(dbPath); err != nil {
		t.Fatalf("generateDBSummary returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dbPath, "README.md"))
	if err != nil {
		t.Fatalf("failed to read summary README: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "## Dotfile Summary") {
		t.Fatalf("expected dotfile summary section, got:\n%s", content)
	}
	if !strings.Contains(content, "| Dotfile Path | Categories | Unique Versions | Total Uses |") {
		t.Fatalf("expected categorized dotfile table header, got:\n%s", content)
	}
	if !strings.Contains(content, "[.gitignore](dotfiles/.gitignore/README.md) | Base | 1 | 1 |") {
		t.Fatalf("expected categorized dotfile row, got:\n%s", content)
	}
}

func TestClearDotfilesOutputRemovesGeneratedDotfiles(t *testing.T) {
	t.Parallel()

	dbPath := t.TempDir()
	configPath := filepath.Join(dbPath, "dotfiles.yaml")
	if err := os.WriteFile(configPath, []byte("dotfiles:\n  - .gitignore\n"), 0644); err != nil {
		t.Fatalf("failed to write dotfiles config: %v", err)
	}
	if err := updateDotfileIndex(dbPath, ".gitignore", "repo-a", "hash-one", "Default"); err != nil {
		t.Fatalf("updateDotfileIndex returned error: %v", err)
	}

	if err := clearDotfilesOutput(dbPath); err != nil {
		t.Fatalf("clearDotfilesOutput returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dbPath, "dotfiles")); !os.IsNotExist(err) {
		t.Fatalf("expected dotfiles directory to be removed, got err=%v", err)
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected dotfiles.yaml to be preserved, got err=%v", err)
	}
}
