COMMIT = $$(git describe --always)

generate:
	@go generate ./...

build: generate
	@echo "====> Build rls"
	go build -ldflags "-X main.GitCommit=\"$(COMMIT)\""