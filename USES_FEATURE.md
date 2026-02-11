# USES.md Feature Implementation

## Overview

This implementation adds automatic generation of a `USES.md` file that indexes all GitHub Actions `uses` statements across workflow files in an organization.

## How It Works

1. **Parsing**: The application parses each workflow YAML file and extracts all `uses` statements from steps
2. **Version Extraction**: For each use, it captures:
   - The action name (e.g., `actions/checkout`)
   - The version or reference (e.g., `v6`, `de0fac2e4500dabe0009e67214ff5f5447ce83dd`)
   - Any inline comments (e.g., `# v6.0.2`)
3. **Aggregation**: Data is aggregated by:
   - Action name (alphabetically sorted)
   - Version within each action (alphabetically sorted)
   - List of workflow files using each version
4. **Generation**: A markdown file is generated with:
   - Action sections with total usage counts
   - Version subsections with usage counts
   - Collapsible details sections listing all workflow files
   - Direct links to workflow files on GitHub

## Output Location

The `USES.md` file is generated in the `db` folder alongside:
- `README.md` (workflow summary)
- `repositories.yaml` (repository manifest)
- `workflows/` (workflow file storage)
- `dependabot/` (dependabot configuration storage)

## Example

See `USES_EXAMPLE.md` in the repository root for a sample of the generated output.

## Key Features

- **Comprehensive**: Captures all action uses across all workflows
- **Version Aware**: Includes both references and inline comments
- **Organized**: Alphabetically sorted for easy navigation
- **Minimal Noise**: Collapsible sections prevent overwhelming detail
- **Actionable**: Direct links to source files for investigation
- **Automated**: Generated automatically during each audit run

## Use Cases

1. **Standardization**: Identify which versions of actions are in use across the organization
2. **Updates**: Find all workflows using specific action versions for coordinated updates
3. **Security**: Quickly locate workflows using deprecated or vulnerable action versions
4. **Compliance**: Track action usage for compliance and governance purposes
5. **Optimization**: Identify opportunities to consolidate on fewer action versions
