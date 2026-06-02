.PHONY: lint release

lint:
	golangci-lint run ./...

release: ## Tag and push a release (usage: make release VERSION=v0.1.0)
	git tag $(VERSION)
	git push origin $(VERSION)