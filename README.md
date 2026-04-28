[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![Work In Progress](https://img.shields.io/badge/Status-Work%20In%20Progress-yellow)](https://guide.unitvectorylabs.com/bestpractices/status/#work-in-progress) 

# dotgithubindexer

A tool for indexing the different GitHub Actions workflows files across repositories within an organization.

## Overview

This tool is designed to help index all of the different GitHub Actions workflows files across repositories within an organization. This can be useful for understanding the different variants of workflows that are being used across repositories.

While this tool is not opinionated, its intended use case is when the different workflows with the same name are intended to be identical between repositories. This tool can help identify when this is not the case and make it easier to resolve those conflicts.

In addition to workflows and `.github/dependabot.yml`, the tool can optionally index specific dotfiles that live outside of `.github` when they are configured in `db/dotfiles.yaml`.

## Use

```text
Usage: dotgithubindexer -org <organization> -token <token> [options]
  -db string
    	Path to the database repository (default "./db")
  -org string
    	GitHub Organization name (required)
  -private
    	Include private repositories; boolean
  -public
    	Include public repositories; boolean (default true)
  -token string
    	GitHub API token (required)
```

## Optional Dotfile Indexing

Additional dotfiles are only indexed when `dotfiles.yaml` exists in the configured database folder. If that file is missing, the existing behavior is unchanged.

```yaml
dotfiles:
  - .gitignore
  - .gitattributes
  - .config/example.yml
```

Each configured path is treated as a repository-relative path outside of `.github`. If the file exists in a repository, it is indexed under `db/dotfiles/`. These files use the same optional `# dotgithubindexer: <category>` comment convention as dependabot files. Categories are only reflected in the generated dotfile output when at least one indexed dotfile uses a non-default category; otherwise dotfiles are grouped by file path like workflows.

## Archived Repositories

Archived repositories are automatically excluded from indexing because they cannot be modified. When fetching repositories from the GitHub API, archived repositories are filtered out and will not be indexed.

## Folder Structure

This application does not utilize a database, instead the content is output to text files and is intended to be committed to a git repository. The folder structure is as follows:

```text
.
└── db
    ├── workflows
    │   ├── build.yml
    │   │   ├── 559aead08264d5795d3909718cdd05abd49572e84fe55590eef31a88a08fdffd
    │   │   ├── df7e70e5021544f4834bbee64a9e3789febc4be81470df629cad6ddb03320a5c
    │   │   └── index.yaml
    │   └── release.yml
    │       ├── 6b23c0d5f35d1b11f9b683f0b0a617355deb11277d91ae091d399c655b87940d
    │       └── index.yaml
    ├── dotfiles
    │   └── .gitignore
    │       ├── 559aead08264d5795d3909718cdd05abd49572e84fe55590eef31a88a08fdffd
    │       ├── index.yaml
    │       └── README.md
    └── repositories.yaml
```

The `repositories.yaml` file contains the index of 

```yaml
organization: UnitVectorY-Labs
repositories:
    - repository-a
    - repository-b
```

The folder structure within the `workflows` folder represents each workflow file that was identified. In that folder there is a file for each unique version of the workflow file whose name is the hash of the file content to ensure uniqueness. The `index.yaml` file contains the index mapping each repository to the file hash.

```yaml
repositories:
    repository-a: 559aead08264d5795d3909718cdd05abd49572e84fe55590eef31a88a08fdffd
    repository-b: df7e70e5021544f4834bbee64a9e3789febc4be81470df629cad6ddb03320a5c
```

A `README.md` file is generated for each workflow file that links to that file on GitHub for easy reference.

Configured dotfiles follow the same pattern under `db/dotfiles/<path>/`, and also generate `README.md` files for easy review.
