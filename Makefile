.PHONY: test
test:
	go test ./...

.PHONY: test-verbose
test-verbose:
	go test -v ./...

.PHONY: test-coverage
test-coverage:
	go test -cover ./...

.PHONY: build
build:
	go build -o pudd .

.PHONY: install
install:
	go install

.PHONY: clean
clean:
	rm -f pudd

.PHONY: tools
tools: ## Install development tools
	go install github.com/caarlos0/svu@latest

.PHONY: next-version
next-version: ## Calculate next semantic version
	@go run github.com/caarlos0/svu next

.PHONY: create-tag
create-tag: ## Create a release tag based on conventional commits (doesn't push)
	@VERSION=$$(go run github.com/caarlos0/svu next); \
	echo "Creating tag $$VERSION"; \
	git tag -a $$VERSION -m "Release $$VERSION"

.PHONY: push-tag
push-tag: ## Push the latest tag to origin
	@TAG=$$(git describe --tags --abbrev=0); \
	echo "Pushing tag $$TAG"; \
	git push origin $$TAG

.PHONY: publish-release
publish-release: ## Run GoReleaser to create and publish the release
	goreleaser release --clean

.PHONY: test-release-locally
test-release-locally: ## Test the release process locally without publishing
	goreleaser release --snapshot --clean
