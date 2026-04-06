# Overview

dotgithubindexer is a command-line tool that indexes configuration files across repositories within a GitHub organization. It scans repositories and collects workflow files, dependabot configurations, and root-level dot files, storing them in a structured file-based database for comparison and auditing.

## Purpose

When managing many repositories in a GitHub organization, configuration files such as GitHub Actions workflows, dependabot configurations, and editor settings are often intended to be consistent across repositories. Over time these files can drift, making it difficult to identify which repositories have outdated or non-standard configurations.

dotgithubindexer solves this by:

- Scanning all repositories in an organization
- Collecting and hashing configuration files
- Grouping identical file versions together
- Producing a structured output that makes it easy to identify inconsistencies

## Design Philosophy

The tool operates on a simple principle: collect, hash, and compare. Each file is identified by its SHA-256 content hash. Repositories sharing the same hash for a given file have identical configurations. Repositories with different hashes have diverged.

The output is stored as plain text files (YAML and Markdown) rather than a traditional database. This design allows the output to be committed to a git repository, providing a version-controlled history of configuration changes over time.

## What Gets Indexed

dotgithubindexer indexes three types of files:

1. **Workflow Files** — GitHub Actions workflow YAML files found in `.github/workflows/`
2. **Dependabot Files** — The `.github/dependabot.yml` configuration file
3. **Root Files** — Configurable dot files in the repository root (e.g., `.editorconfig`, `.prettierrc.json`)

For details on each file type, see the [Indexed File Types](file-types.md) documentation.

## Next Steps

- [Installation & Usage](usage.md) — How to install and run the tool
- [Configuration](configuration.md) — Available options and settings
- [Database Structure](database-structure.md) — Understanding the output format
