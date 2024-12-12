package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

// ------------------------
// Section: Type Definitions
// ------------------------

// RepositoryManifest represents the manifest of all repositories processed.
type RepositoryManifest struct {
	Organization string   `yaml:"organization"`
	Repositories []string `yaml:"repositories"`
}

// ActionIndex maps repositories to the hash of the workflow file they use.
type ActionIndex struct {
	Repositories map[string]string `yaml:"repositories"` // RepoName: Hash
}

// WorkflowFile represents a GitHub Actions workflow file.
type WorkflowFile struct {
	RepoName string
	FilePath string
	Content  string
	Hash     string
}

// ------------------------
// Section: Global Variables
// ------------------------

var (
	org        string
	includePub bool
	includePrv bool
	token      string
	dbPath     string
)

// ------------------------
// Section: Main Function and CLI Setup
// ------------------------

func main() {

	// Define CLI flags using the flag package
	flag.StringVar(&org, "org", "", "GitHub Organization name (required)")
	flag.BoolVar(&includePub, "public", true, "Include public repositories; boolean")
	flag.BoolVar(&includePrv, "private", false, "Include private repositories; boolean")
	flag.StringVar(&token, "token", "", "GitHub API token (required)")
	flag.StringVar(&dbPath, "db", "./db", "Path to the database repository")

	flag.Parse()

	// Check required flags
	if org == "" || token == "" {
		fmt.Println("Usage: dotgithubindexer -org <organization> -token <token> [options]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Execute main audit logic
	startTime := time.Now()
	fmt.Println("Starting GitHub Actions Audit")

	err := auditGitHubActions(org, token, dbPath, includePub, includePrv)
	if err != nil {
		fmt.Printf("Audit failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Audit completed successfully in %v.\n", time.Since(startTime))
}

// ------------------------
// Section: GitHub Client Setup
// ------------------------

// getGitHubClient authenticates with GitHub using the provided token.
func getGitHubClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return client
}

// ------------------------
// Section: Fetch Repositories
// ------------------------

// fetchRepositories retrieves repositories based on visibility options.
func fetchRepositories(client *github.Client, org string, includePub, includePrv bool) ([]*github.Repository, error) {
	ctx := context.Background()
	var allRepos []*github.Repository
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			visibility := repo.GetVisibility()

			if includePub && visibility == "public" {
				allRepos = append(allRepos, repo)
			}
			if includePrv && visibility == "private" {
				allRepos = append(allRepos, repo)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage

		// Handle rate limiting
		if err := checkRateLimit(client); err != nil {
			return nil, err
		}
	}

	return allRepos, nil
}

// ------------------------
// Section: Fetch Workflow Files
// ------------------------

// fetchWorkflowFiles retrieves workflow files from a repository.
func fetchWorkflowFiles(client *github.Client, repo *github.Repository) ([]WorkflowFile, error) {
	ctx := context.Background()
	workflows := []WorkflowFile{}

	defaultBranch := getDefaultBranch(repo)
	fmt.Printf("Default branch for repository '%s' is '%s'\n", repo.GetName(), defaultBranch)

	// Check the .github/workflows directory
	_, workflowFiles, _, err := client.Repositories.GetContents(ctx, repo.GetOwner().GetLogin(), repo.GetName(), ".github/workflows", &github.RepositoryContentGetOptions{
		Ref: defaultBranch,
	})

	if err != nil {
		// If '.github/workflows' is not found, try 'workflows' directly under root
		if _, ok := err.(*github.ErrorResponse); ok && strings.Contains(err.Error(), "404") {
			fmt.Printf("No '.github/workflows' directory found in repository '%s'. Trying 'workflows' directory.\n", repo.GetName())
			_, workflowFiles, _, err = client.Repositories.GetContents(ctx, repo.GetOwner().GetLogin(), repo.GetName(), "workflows", &github.RepositoryContentGetOptions{
				Ref: defaultBranch,
			})
			if err != nil {
				// Repository might not have workflows
				fmt.Printf("No 'workflows' directory found in repository '%s'. Skipping.\n", repo.GetName())
				return workflows, nil
			}
		} else {
			fmt.Printf("Error accessing workflows directory in repository '%s': %v\n", repo.GetName(), err)
			return nil, err
		}
	}

	if workflowFiles == nil || len(workflowFiles) == 0 {
		fmt.Printf("No workflow files found in repository '%s'.\n", repo.GetName())
		return workflows, nil
	}

	// Iterate through the files in the directory
	for _, file := range workflowFiles {
		if file.GetType() == "file" {
			fmt.Printf("Found workflow file: %s in repository '%s'\n", file.GetPath(), repo.GetName())

			// Fetch the blob to get the content
			blob, _, err := client.Git.GetBlob(ctx, repo.GetOwner().GetLogin(), repo.GetName(), file.GetSHA())
			if err != nil {
				fmt.Printf("Error fetching blob for file '%s' in repository '%s': %v\n", file.GetPath(), repo.GetName(), err)
				continue
			}

			// Decode the content from base64
			contentBytes, err := base64.StdEncoding.DecodeString(blob.GetContent())
			if err != nil {
				fmt.Printf("Error decoding content for file '%s' in repository '%s': %v\n", file.GetPath(), repo.GetName(), err)
				continue
			}
			content := string(contentBytes)

			if content == "" {
				fmt.Printf("Empty content for file '%s' in repository '%s'\n", file.GetPath(), repo.GetName())
				continue
			}
			hash := computeHash([]byte(content))
			fmt.Printf("Hashing file '%s' in repository '%s': %s\n", file.GetPath(), repo.GetName(), hash)
			workflows = append(workflows, WorkflowFile{
				RepoName: repo.GetName(),
				FilePath: file.GetPath(),
				Content:  content,
				Hash:     hash,
			})
		}
	}

	return workflows, nil
}

// getDefaultBranch retrieves the default branch of a repository.
func getDefaultBranch(repo *github.Repository) string {
	if repo.GetDefaultBranch() != "" {
		return repo.GetDefaultBranch()
	}
	return "main" // Fallback to 'main' if not specified
}

// computeHash computes the SHA-256 hash of the given content.
func computeHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// ------------------------
// Section: Database Management
// ------------------------

// initializeDB sets up the database directory and initial manifests.
func initializeDB(dbPath string) error {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("Creating database directory at '%s'\n", dbPath)
		err := os.MkdirAll(dbPath, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("Database directory '%s' already exists.\n", dbPath)
	}

	// Initialize repositories.yaml
	reposManifestPath := filepath.Join(dbPath, "repositories.yaml")
	if _, err := os.Stat(reposManifestPath); os.IsNotExist(err) {
		fmt.Printf("Creating 'repositories.yaml' at '%s'\n", reposManifestPath)
		emptyManifest := RepositoryManifest{Organization: org, Repositories: []string{}}
		data, _ := yaml.Marshal(&emptyManifest)
		err = os.WriteFile(reposManifestPath, data, 0644)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("'repositories.yaml' already exists at '%s'\n", reposManifestPath)
	}

	// Initialize actions directory
	actionsPath := filepath.Join(dbPath, "workflows")
	if _, err := os.Stat(actionsPath); os.IsNotExist(err) {
		fmt.Printf("Creating 'actions' directory at '%s'\n", actionsPath)
		err = os.MkdirAll(actionsPath, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("'actions' directory already exists at '%s'\n", actionsPath)
	}

	return nil
}

// updateRepositoriesManifest adds a repository to the repositories.yaml manifest.
func updateRepositoriesManifest(dbPath string, repoName string) error {
	reposManifestPath := filepath.Join(dbPath, "repositories.yaml")
	var manifest RepositoryManifest

	data, err := os.ReadFile(reposManifestPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, &manifest)
	if err != nil {
		return err
	}

	// Add repo if not exists
	for _, r := range manifest.Repositories {
		if r == repoName {
			return nil
		}
	}
	manifest.Repositories = append(manifest.Repositories, repoName)

	// Sort repositories alphabetically
	sort.Strings(manifest.Repositories)

	updatedData, err := yaml.Marshal(&manifest)
	if err != nil {
		return err
	}
	err = os.WriteFile(reposManifestPath, updatedData, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Added repository '%s' to 'repositories.yaml'\n", repoName)
	return nil
}

// updateActionIndex maps a repository to a workflow file hash in the action's index.
func updateActionIndex(dbPath, actionName, repoName, hash string) error {
	actionPath := filepath.Join(dbPath, "workflows", actionName)
	if err := os.MkdirAll(actionPath, os.ModePerm); err != nil {
		return err
	}

	indexPath := filepath.Join(actionPath, "index.yaml")
	var index ActionIndex

	if _, err := os.Stat(indexPath); err == nil {
		data, err := os.ReadFile(indexPath)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(data, &index)
		if err != nil {
			return err
		}
	} else {
		index = ActionIndex{Repositories: make(map[string]string)}
	}

	index.Repositories[repoName] = hash

	// Sort repositories alphabetically by key
	sortedKeys := make([]string, 0, len(index.Repositories))
	for k := range index.Repositories {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	sortedRepositories := make(map[string]string)
	for _, k := range sortedKeys {
		sortedRepositories[k] = index.Repositories[k]
	}
	index.Repositories = sortedRepositories

	updatedData, err := yaml.Marshal(&index)
	if err != nil {
		return err
	}
	err = os.WriteFile(indexPath, updatedData, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Updated action index for action '%s' with repository '%s'\n", actionName, repoName)
	return nil
}

// storeActionVersion saves the workflow file content under its hash.
func storeActionVersion(dbPath, actionName, hash, content string) error {
	actionPath := filepath.Join(dbPath, "workflows", actionName)
	if err := os.MkdirAll(actionPath, os.ModePerm); err != nil {
		return err
	}

	filePath := filepath.Join(actionPath, fmt.Sprintf("%s", hash))
	// Check if file already exists to avoid unnecessary writes
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Storing workflow file '%s' under hash '%s'\n", actionName, hash)
		return os.WriteFile(filePath, []byte(content), 0644)
	}

	fmt.Printf("Workflow file with hash '%s' already exists. Skipping write.\n", hash)
	return nil
}

// ------------------------
// Section: Garbage Collection
// ------------------------

// garbageCollect removes unused workflow file versions from the database.
func garbageCollect(dbPath string) error {
	actionsPath := filepath.Join(dbPath, "workflows")
	if _, err := os.Stat(actionsPath); os.IsNotExist(err) {
		fmt.Printf("No 'workflows' directory found at '%s'. Skipping garbage collection.\n", actionsPath)
		return nil
	}

	dirs, err := os.ReadDir(actionsPath)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			actionName := dir.Name()
			indexPath := filepath.Join(actionsPath, actionName, "index.yaml")
			var index ActionIndex
			data, err := os.ReadFile(indexPath)
			if err != nil {
				fmt.Printf("No index found for action '%s'. Skipping.\n", actionName)
				continue
			}
			err = yaml.Unmarshal(data, &index)
			if err != nil {
				fmt.Printf("Error unmarshaling index for action '%s': %v\n", actionName, err)
				continue
			}

			// Collect all hashes in use
			hashesInUse := make(map[string]bool)
			for _, hash := range index.Repositories {
				hashesInUse[hash] = true
			}

			// Iterate over all files in action directory
			actionDirPath := filepath.Join(actionsPath, actionName)
			files, err := os.ReadDir(actionDirPath)
			if err != nil {
				fmt.Printf("Error reading action directory '%s': %v\n", actionDirPath, err)
				continue
			}

			for _, file := range files {
				if file.IsDir() || file.Name() == "index.yaml" {
					continue
				}
				hash := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
				if !hashesInUse[hash] {
					fmt.Printf("Removing unused workflow file '%s' from action '%s'\n", file.Name(), actionName)
					os.Remove(filepath.Join(actionDirPath, file.Name()))
				}
			}
		}
	}

	fmt.Println("Garbage collection completed.")
	return nil
}

// ------------------------
// Section: Rate Limiting
// ------------------------

// checkRateLimit monitors GitHub API rate limits and waits if necessary.
func checkRateLimit(client *github.Client) error {
	ctx := context.Background()
	rate, _, err := client.RateLimits(ctx)
	if err != nil {
		return err
	}

	core := rate.GetCore()
	if core.Remaining < 100 {
		waitDuration := time.Until(core.Reset.Time) + time.Minute
		fmt.Printf("Rate limit low (%d remaining). Waiting for %v...\n", core.Remaining, waitDuration)
		time.Sleep(waitDuration)
	}

	return nil
}

// ------------------------
// Section: Audit Function
// ------------------------

// auditGitHubActions orchestrates the entire audit process.
func auditGitHubActions(org, token, dbPath string, includePub, includePrv bool) error {
	client := getGitHubClient(token)

	// Initialize DB
	if err := initializeDB(dbPath); err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	// Fetch Repositories
	repos, err := fetchRepositories(client, org, includePub, includePrv)
	if err != nil {
		return fmt.Errorf("failed to fetch repositories: %v", err)
	}

	for _, repo := range repos {
		repoName := repo.GetName()
		fmt.Printf("Processing repository: %s\n", repoName)

		// Update repositories manifest
		if err := updateRepositoriesManifest(dbPath, repoName); err != nil {
			fmt.Printf("Error updating repositories manifest for %s: %v\n", repoName, err)
			continue
		}

		// Fetch workflow files
		workflows, err := fetchWorkflowFiles(client, repo)
		if err != nil {
			fmt.Printf("Error fetching workflow files for %s: %v\n", repoName, err)
			continue
		}

		if len(workflows) == 0 {
			fmt.Printf("No workflow files to process in repository '%s'.\n", repoName)
			continue
		}

		for _, wf := range workflows {
			actionName := filepath.Base(wf.FilePath)

			// Update action index
			if err := updateActionIndex(dbPath, actionName, wf.RepoName, wf.Hash); err != nil {
				fmt.Printf("Error updating action index for %s in %s: %v\n", actionName, repoName, err)
				continue
			}

			// Store action version
			if err := storeActionVersion(dbPath, actionName, wf.Hash, wf.Content); err != nil {
				fmt.Printf("Error storing action version for %s in %s: %v\n", actionName, repoName, err)
				continue
			}
		}

		// Handle rate limiting after processing each repository
		if err := checkRateLimit(client); err != nil {
			return fmt.Errorf("rate limit check failed: %v", err)
		}
	}

	// Perform garbage collection
	if err := garbageCollect(dbPath); err != nil {
		fmt.Printf("Error during garbage collection: %v\n", err)
	}

	// Generate README.md files
	if err := generateReadmeFiles(dbPath, org); err != nil {
		fmt.Printf("Error generating README.md files: %v\n", err)
	}

	return nil
}

// generateReadmeFiles creates README.md files in each action directory with links to workflow files.
func generateReadmeFiles(dbPath, org string) error {
	actionsPath := filepath.Join(dbPath, "workflows")
	dirs, err := os.ReadDir(actionsPath)
	if err != nil {
		return fmt.Errorf("failed to read actions directory: %v", err)
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			actionName := dir.Name()
			indexPath := filepath.Join(actionsPath, actionName, "index.yaml")
			var index ActionIndex

			data, err := os.ReadFile(indexPath)
			if err != nil {
				fmt.Printf("Skipping action '%s' due to missing index.yaml.\n", actionName)
				continue
			}

			err = yaml.Unmarshal(data, &index)
			if err != nil {
				fmt.Printf("Error parsing index.yaml for action '%s': %v\n", actionName, err)
				continue
			}

			// Reverse mapping from hash to repositories
			hashToRepos := make(map[string][]string)
			for repo, hash := range index.Repositories {
				hashToRepos[hash] = append(hashToRepos[hash], repo)
			}

			var markdownBuilder strings.Builder
			markdownBuilder.WriteString(fmt.Sprintf("# %s\n\n", actionName))
			for hash, repos := range hashToRepos {
				markdownBuilder.WriteString(fmt.Sprintf("## [%s](%s)\n\n", hash, hash))
				for _, repo := range repos {
					filePath := ".github/workflows/" + actionName
					url := fmt.Sprintf("https://github.com/%s/%s/blob/main/%s", org, repo, filePath)
					markdownBuilder.WriteString(fmt.Sprintf("- [%s](%s)\n", repo, url))
				}
				markdownBuilder.WriteString("\n")
			}

			readmePath := filepath.Join(actionsPath, actionName, "README.md")
			err = os.WriteFile(readmePath, []byte(markdownBuilder.String()), 0644)
			if err != nil {
				fmt.Printf("Error writing README.md for action '%s': %v\n", actionName, err)
				continue
			}

			fmt.Printf("Generated README.md for action '%s'\n", actionName)
		}
	}

	return nil
}
