package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadDotfilesConfig(t *testing.T) {
	dbPath := t.TempDir()
	configPath := filepath.Join(dbPath, "dotfiles.yaml")

	if err := os.WriteFile(configPath, []byte("dotfiles:\n  - .gitignore\n  - .repver\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := loadDotfilesConfig(dbPath)
	if err != nil {
		t.Fatalf("loadDotfilesConfig() error = %v", err)
	}

	want := []string{".gitignore", ".repver"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("loadDotfilesConfig() = %v, want %v", got, want)
	}
}

func TestResolveRootFilesMergesConfigAndFlag(t *testing.T) {
	dbPath := t.TempDir()
	configPath := filepath.Join(dbPath, "dotfiles.yaml")

	if err := os.WriteFile(configPath, []byte("dotfiles:\n  - .gitignore\n  - .repver\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := resolveRootFiles(dbPath, ".repver, .clip4llm")
	if err != nil {
		t.Fatalf("resolveRootFiles() error = %v", err)
	}

	want := []string{".gitignore", ".repver", ".clip4llm"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resolveRootFiles() = %v, want %v", got, want)
	}
}

func TestInitializeDBPreservesDotfilesConfig(t *testing.T) {
	dbPath := t.TempDir()
	configPath := filepath.Join(dbPath, "dotfiles.yaml")

	if err := os.WriteFile(configPath, []byte("dotfiles:\n  - .gitignore\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	previousOrg := org
	t.Cleanup(func() {
		org = previousOrg
	})
	org = "UnitVectorY-Labs"

	if err := initializeDB(dbPath); err != nil {
		t.Fatalf("initializeDB() error = %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if string(data) != "dotfiles:\n  - .gitignore\n" {
		t.Fatalf("dotfiles.yaml changed unexpectedly: %q", string(data))
	}
}
