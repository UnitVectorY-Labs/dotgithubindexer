# Installation & Usage

## Prerequisites

- Go 1.21 or later
- A GitHub personal access token with appropriate repository read permissions

## Installation

Build from source:

```bash
go build -o dotgithubindexer .
```

Or install directly:

```bash
go install github.com/UnitVectorY-Labs/dotgithubindexer@latest
```

## Basic Usage

```bash
dotgithubindexer -org <organization> -token <token>
```

This scans all public repositories in the specified organization and writes the results to a `./db` directory.

## Command-Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-org` | *(required)* | GitHub organization name to scan |
| `-token` | *(required)* | GitHub API personal access token |
| `-db` | `./db` | Path to the output database directory |
| `-public` | `true` | Include public repositories |
| `-private` | `false` | Include private repositories |
| `-rootfiles` | *(empty)* | Comma-separated list of root dot files to index |
| `-version` | | Print the version and exit |

## Examples

Index only public repositories:

```bash
dotgithubindexer -org my-org -token ghp_xxxx
```

Include private repositories:

```bash
dotgithubindexer -org my-org -token ghp_xxxx -private
```

Index root dot files in addition to workflows and dependabot:

```bash
dotgithubindexer -org my-org -token ghp_xxxx -rootfiles .editorconfig,.prettierrc.json
```

Write output to a custom directory:

```bash
dotgithubindexer -org my-org -token ghp_xxxx -db /path/to/output
```

## Rate Limiting

The tool monitors GitHub API rate limits during execution. When the remaining rate limit falls below 100 requests, the tool automatically pauses and waits for the rate limit to reset before continuing. This ensures the tool can process large organizations without hitting API limits.

## Archived Repositories

Archived repositories are automatically excluded from indexing. Since archived repositories cannot be modified, they are filtered out during the repository scanning phase.
