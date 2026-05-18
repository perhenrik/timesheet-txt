BINARY ?= timesheet
VERSION ?= 0.2.0
PLATFORMS := darwin linux windows
ARCH := amd64 arm64
AIR ?= air

.PHONY: fmt vet test lint build clean release dev-tui

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint: fmt vet

build:
	mkdir -p bin
	go build -o bin/$(BINARY) .

clean:
	rm -rf bin release coverage.out timesheet

release:
	mkdir -p release
	for os in $(PLATFORMS); do \
		for arch in $(ARCH); do \
			GOOS=$$os GOARCH=$$arch go build -o release/$(BINARY)-$(VERSION)-$$os-$$arch . ; \
		done ; \
	done

dev-tui:
	@command -v $(AIR) >/dev/null 2>&1 || { echo "air is not installed. Install with: brew install air"; exit 1; }
	$(AIR) --build.cmd "go build -o ./tmp/timesheet ." --build.bin "./tmp/timesheet tui"
