PLUGIN_DATA_DIR = $(HOME)/.claude/plugins/data/claude-code-notification-inline
VERSION = $(shell jq -r '.version' plugin/.claude-plugin/plugin.json)

.PHONY: build install test coverage clean

build:
	go build -ldflags "-X main.version=$(VERSION)" -o notification ./cmd/claude-code-notification

install: build
	mkdir -p $(PLUGIN_DATA_DIR)/bin
	cp notification $(PLUGIN_DATA_DIR)/bin/notification
	echo "Installed claude-code-notification version $(VERSION) to $(PLUGIN_DATA_DIR)/bin/notification"

unit-test:
	go clean -testcache
	go test ./internal/...

integration-test:
	go clean -testcache
	go test ./test/integration/

coverage:
	@rm -rf .coverdata && mkdir -p .coverdata/unit .coverdata/integration
	go clean -testcache
	@echo ""
	@echo "--- Running unit tests ---"
	go test -coverprofile=.coverdata/unit.txt ./internal/...
	@echo ""
	@echo "--- Running integration tests ---"
	INTEGRATION_COVER_DIR=$(CURDIR)/.coverdata/integration go test ./test/integration/
	go tool covdata textfmt -i=.coverdata/integration -o=.coverdata/integration.txt
	go tool cover -func=.coverdata/unit.txt
	go tool cover -html=.coverdata/unit.txt -o .coverdata/unit.html
	@echo ""
	@echo "--- Integration coverage (binary) ---"
	go tool cover -func=.coverdata/integration.txt
	go tool cover -html=.coverdata/integration.txt -o .coverdata/integration.html
	@echo ""
	@echo "HTML reports: .coverdata/unit.html  .coverdata/integration.html"

clean:
	rm -f notification
