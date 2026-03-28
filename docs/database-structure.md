# Database Structure

dotgithubindexer does not use a traditional database. Instead, it outputs files to a directory structure that can be committed to a git repository. This provides version-controlled tracking of configuration changes over time.

## Directory Layout

```text
db/
├── README.md
├── USES.md
├── repositories.yaml
├── workflows/
│   └── <workflow-name>/
│       ├── README.md
│       ├── index.yaml
│       └── <hash>
├── dependabot/
│   └── <category>/
│       ├── README.md
│       ├── index.yaml
│       └── <hash>
└── rootfiles/
    └── <filename>/
        └── <category>/
            ├── README.md
            ├── index.yaml
            └── <hash>
```

## File Descriptions

### `repositories.yaml`

Lists all repositories that were indexed during the scan.

```yaml
organization: my-org
repositories:
    - repo-a
    - repo-b
    - repo-c
```

### `README.md` (root)

A summary table showing all indexed workflows, dependabot categories, and root file categories with counts of unique versions and total uses.

### `USES.md`

An index of all GitHub Actions used across workflow files, organized by action name and version, with links to the workflow files that use each version.

### Workflow Index (`workflows/<name>/index.yaml`)

Maps each repository to the content hash of its workflow file.

```yaml
repositories:
    repo-a: 559aead08264d5795d3909718cdd05abd49572e84fe55590eef31a88a08fdffd
    repo-b: df7e70e5021544f4834bbee64a9e3789febc4be81470df629cad6ddb03320a5c
```

### Hash Files

Each unique file version is stored as a plain text file named by its SHA-256 hash. This allows direct comparison of file contents.

### Dependabot Index (`dependabot/<category>/index.yaml`)

Maps each repository to the content hash of its dependabot file within a category.

### Root Files Index (`rootfiles/<filename>/<category>/index.yaml`)

Maps each repository to the content hash of the root file within a given file name and category combination.

## Garbage Collection

After each run, the tool performs garbage collection to remove hash files that are no longer referenced by any index. This keeps the database clean when file contents change between runs.

## Generated Markdown

README.md files are generated in each workflow, dependabot category, and root file category directory. These files provide a human-readable view with links to the source files on GitHub, grouped by content hash. Repositories sharing the same hash appear together, making it easy to identify groups of identical configurations.
