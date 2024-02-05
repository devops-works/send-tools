BUILD_DATE := $(shell date -u '+%Y%m%d.%H%M%S')
VERSION := $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
                        cat $(CURDIR)/.version 2> /dev/null || echo v0)
FLAGS := -X main.buildDate=$(BUILD_DATE) -X main.version=$(VERSION)
STATIC := -a -ldflags "-extldflags '-static' $(FLAGS)"

IMAGE_NAME = devopsworks/phpsecscan:${VERSION}

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: no-dirty
no-dirty:
	git diff --exit-code

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## audit: run quality control checks
.PHONY: audit
audit:
	go mod verify
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	go test -race -buildvcs -vet=off ./...

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

## build: build the application
.PHONY: build
build:
	go build -o send-cleanup -ldflags "$(FLAGS)" ./cmd/send-cleanup/

## static: statically build the application
.PHONY: static
static:
	CGO_ENABLED=0 GOOS=linux go build -o send-cleanup $(STATIC) ./cmd/send-cleanup/
