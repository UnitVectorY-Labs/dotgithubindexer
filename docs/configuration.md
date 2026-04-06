# Configuration

## CLI Flags

All configuration is provided through command-line flags. There are no configuration files.

### Required Flags

- **`-org`** — The GitHub organization name. The tool will scan repositories belonging to this organization.
- **`-token`** — A GitHub personal access token. The token needs read access to repository contents for the repositories you want to index.

### Optional Flags

- **`-db`** — The output directory path. Defaults to `./db`. The directory is created if it does not exist.
- **`-public`** — Whether to include public repositories. Defaults to `true`.
- **`-private`** — Whether to include private repositories. Defaults to `false`.
- **`-rootfiles`** — A comma-separated list of file names to index from the root of each repository. This is intended for dot files such as `.editorconfig` or `.prettierrc.json`.

## Root Files Configuration

The `-rootfiles` flag accepts a comma-separated list of file names. Each file name is looked up in the root directory of every repository. If the file exists, it is indexed using the same category-based system as dependabot files.

Example:

```bash
-rootfiles .editorconfig,.prettierrc.json,.eslintrc.json
```

This will look for `.editorconfig`, `.prettierrc.json`, and `.eslintrc.json` in the root of each repository.

## Category Comments

Both dependabot files and root files support category-based grouping through a special comment format. Adding the following comment anywhere in a file assigns it to a category:

```
# dotgithubindexer: <category>
```

For example, a `.editorconfig` file might include:

```
# dotgithubindexer: java-standard
root = true

[*]
indent_style = space
indent_size = 4
```

This assigns the file to the `java-standard` category. Files without this comment are assigned to the `Default` category.

Categories are useful when the same file type is intentionally different across groups of repositories. For example, Java projects may use different editor settings than JavaScript projects. By assigning categories, you can compare files within the same category rather than across all repositories.

## Token Permissions

The GitHub token requires the following permissions:

- **`repo`** scope for private repositories
- **`public_repo`** scope for public repositories only
