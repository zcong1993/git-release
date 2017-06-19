COMMIT = $$(git describe --always)

generate:
	@go generate ./...

build: generate
	@echo "====> Build rls"
	@sh -c ./build.sh
