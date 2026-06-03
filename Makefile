.PHONY: lint test coverage clean release

lint:
	golangci-lint run ./...

test:
	go test -race -covermode=atomic -coverprofile=coverage.out ./cmd/... ./internal/...
	@COVERAGE=$$(go tool cover -func=coverage.out | grep '^total:' | awk '{print $$3}' | sed 's/%//'); \
	if [ -z "$${COVERAGE}" ]; then echo "❌ Could not extract coverage"; exit 1; fi; \
	echo "Coverage: $${COVERAGE}%"; \
	if awk "BEGIN {exit ($${COVERAGE} >= 70)}"; then \
	  echo "❌ Coverage $${COVERAGE}% is below 70% threshold"; \
	  exit 1; \
	fi; \
	echo "✅ Coverage $${COVERAGE}% meets threshold"

coverage: test
	go tool cover -html=coverage.out

clean:
	rm -f coverage.out

test-integration: ## Run subprocess integration tests (compiles devlog binary first)
	go test -tags=integration -race -timeout=2m -v ./tests/integration/...

test-regression: ## Run golden-file regression tests
	go test -tags=regression -timeout=5m -v ./tests/regression/...

update-golden: ## Regenerate golden files after intentional output changes
	DEVLOG_GOLDEN_UPDATE=1 go test -tags=regression -timeout=5m -run TestRegression ./tests/regression/...

test-all: test test-integration test-regression ## Run unit + integration + regression suites
	@echo "All test suites completed"

release: ## Tag and push a release (usage: make release VERSION=v0.1.0)
	git tag $(VERSION)
	git push origin $(VERSION)
