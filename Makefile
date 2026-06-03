.PHONY: lint test coverage release

lint:
	golangci-lint run ./...

test:
	go test -race -covermode=atomic -coverprofile=coverage.out ./...
	@COVERAGE=$$(go tool cover -func=coverage.out | grep '^total:' | awk '{print $$3}' | sed 's/%//'); \
	echo "Coverage: $${COVERAGE}%"

coverage: test
	go tool cover -html=coverage.out

release: ## Tag and push a release (usage: make release VERSION=v0.1.0)
	git tag $(VERSION)
	git push origin $(VERSION)