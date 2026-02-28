# Indexed File Types

dotgithubindexer indexes three types of files from repositories. Each type is processed differently and stored in its own section of the output database.

## Workflow Files

**Source path:** `.github/workflows/`

The tool scans the `.github/workflows/` directory in each repository and indexes all files found there. Each workflow file is identified by its file name (e.g., `build.yml`, `release.yml`).

Workflow files with the same name across different repositories are grouped together. Each unique file content is identified by its SHA-256 hash. This makes it straightforward to see which repositories share identical workflow files and which have diverged.

The tool also extracts `uses` statements from workflow files to build an index of which GitHub Actions are used across the organization and at which versions.

## Dependabot Files

**Source path:** `.github/dependabot.yml`

The tool looks for a `dependabot.yml` file in the `.github/` directory of each repository. Since there is only one dependabot file per repository, all repositories are grouped together rather than by file name.

Dependabot files support category-based grouping. By adding a `# dotgithubindexer: <category>` comment to the file, repositories can be organized into logical groups. Files without this comment are placed in the `Default` category.

Example with a category comment:

```yaml
# dotgithubindexer: go-terraform
version: 2
updates:
  - package-ecosystem: gomod
    directory: "/"
    schedule:
      interval: weekly
  - package-ecosystem: terraform
    directory: "/"
    schedule:
      interval: weekly
```

## Root Files

**Source path:** Repository root (configurable)

Root files are dot files found in the root directory of repositories. Unlike workflow and dependabot files, root files are not indexed by default. You must specify which files to look for using the `-rootfiles` CLI flag.

Root files follow the same category-based grouping pattern as dependabot files. Adding a `# dotgithubindexer: <category>` comment to the file assigns it to a category. Files without this comment are placed in the `Default` category.

Common root files that can be indexed include:

- `.editorconfig` — Editor configuration
- `.prettierrc.json` — Prettier formatter configuration
- `.eslintrc.json` — ESLint configuration
- `.gitignore` — Git ignore rules
- `.gitattributes` — Git attributes

Example usage:

```bash
dotgithubindexer -org my-org -token ghp_xxxx -rootfiles .editorconfig,.gitattributes
```

The output structure mirrors the dependabot pattern, with an additional level of nesting for the file name. See [Database Structure](database-structure.md) for details.
