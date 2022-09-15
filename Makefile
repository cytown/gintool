GOFMT ?= gofmt "-s"
PACKAGES ?= $(shell go list ./... | grep -v /vendor/)
GOFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*")
GONOTESTFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*"|grep -v _test.go)
GONODOCFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*"|grep -v doc.go)
GOTESTFILES := $(shell find . -name "*_test.go" -type f -not -path "./vendor/*")
BUILD_VER ?= $(shell ver=$$(git describe --always --abbrev=0 --tags); \
             	vtag=$$(git rev-parse $$ver | head -c 8); \
             	tag=$$(git rev-parse HEAD | head -c 8); \
             	if [ "$$vtag" != "$$tag" ]; then \
             		ver=$$ver-$$tag; \
				else \
					ver=$$vtag; \
             	fi ;\echo $$ver)

.PHONY: all
all: alltest

.PHONY: alltest
alltest: fmt vet misspell tidy test lint

.PHONY: allcheck
allcheck: fmt-check vet misspell-check test lint

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	@if [ ! -d "./sonar" ]; then \
		mkdir sonar; \
	fi;
	go test -covermode=atomic -coverprofile=sonar/cover.out $(PACKAGES)
	go tool cover -func=sonar/cover.out

.PHONY: testout
testout:test
	go tool cover -html=sonar/cover.out

.PHONY: fmt
fmt:
	$(GOFMT) -w $(GOFILES)

.PHONY: fmt-check
fmt-check:
	# get all go files and run go fmt on them
	@diff=$$($(GOFMT) -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

vet:
	go vet $(PACKAGES)

.PHONY: misspell-check
misspell-check:
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go get -u github.com/client9/misspell/cmd/misspell; \
	fi
	misspell -error $(GOFILES)

.PHONY: misspell
misspell:
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go get -u github.com/client9/misspell/cmd/misspell; \
	fi
	misspell -w $(GOFILES)

.PHONY: lint
lint:
	@hash golangci-lint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go get github.com/golangci/golangci-lint/cmd/golangci-lint; \
	fi
	# run golanci-lint
	golangci-lint run

