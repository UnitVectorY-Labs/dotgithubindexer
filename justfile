
# Commands for dotgithubindexer
default:
  @just --list
# Build dotgithubindexer with Go
build:
  go build ./...

# Run tests for dotgithubindexer with Go
test:
  go clean -testcache
  go test ./...