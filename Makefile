# Based on https://sahilm.com/makefiles-for-golang/

PKGS := $(shell go list ./... | grep -v /vendor)

.PHONY: test
test:
	go test -coverprofile=coverage.out $(PKGS)
	go tool cover -func=coverage.out
	#go tool cover -html=coverage.out

BIN_DIR := $(GOPATH)/bin
GOMETALINTER := $(BIN_DIR)/gometalinter

$(GOMETALINTER):
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install &> /dev/null

.PHONY: lint
lint: $(GOMETALINTER)
	gometalinter ./... --vendor

BINARY := timesheet
VERSION ?= 0.1
PLATFORMS := windows linux darwin
os = $(word 1, $@)

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	mkdir -p release
	GOOS=$(os) GOARCH=amd64 go build -o release/$(BINARY)-$(VERSION)-$(os)-amd64

.PHONY: release
release: windows linux darwin
