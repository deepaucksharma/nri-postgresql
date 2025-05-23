export PATH := $(PATH):$(GOPATH)/bin
NATIVEOS    := $(shell go version | awk -F '[ /]' '{print $$4}')
NATIVEARCH  := $(shell go version | awk -F '[ /]' '{print $$5}')
INTEGRATION     := postgresql
GOFLAGS          = -mod=readonly # ignore the vendor directory and to report an error if go.mod needs to be updated.
BINARY_NAME      = nri-$(INTEGRATION)
INTEGRATIONS_DIR = /var/db/newrelic-infra/newrelic-integrations/
CONFIG_DIR       = /etc/newrelic-infra/integrations.d
GO_FILES        := ./src/
GO_VERSION 		?= $(shell grep '^go ' go.mod | awk '{print $$2}')
BUILDER_IMAGE 	?= "ghcr.io/newrelic/coreint-automation:latest-go$(GO_VERSION)-ubuntu16.04"

all: build

build: clean test compile

build-container:
	docker build -t nri-postgresql .

clean:
	@echo "=== $(INTEGRATION) === [ clean ]: Removing binaries and coverage file..."
	@rm -rfv bin coverage.xml

compile:
	@echo "=== $(INTEGRATION) === [ compile ]: Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) $(GO_FILES)

test:
	@echo "=== $(INTEGRATION) === [ test ]: running unit tests..."
	@go test -race ./... -count=1


integration-test:
	@echo "=== $(INTEGRATION) === [ test ]: running integration tests..."
	@docker compose -f tests/docker-compose.yml up -d
	# Sleep added to allow postgres with test data and extensions to start up
	@sleep 10
	@go test -v -tags=integration -count 1 ./tests/postgresql_test.go -timeout 300s || (ret=$$?; docker compose -f tests/docker-compose.yml down -v && exit $$ret)
	@docker compose -f tests/docker-compose.yml down -v
	@echo "=== $(INTEGRATION) === [ test ]: running integration tests for query performance monitoring..."
	@echo "Starting containers for performance tests..."
	@docker compose -f tests/docker-compose-performance.yml up -d
	# Sleep added to allow postgres with test data and extensions to start up
	@sleep 30
	@go test -v -tags=query_performance ./tests/postgresqlperf_test.go -timeout 600s || (ret=$$?; docker compose -f tests/docker-compose-performance.yml down -v && exit $$ret)
	@echo "Stopping performance test containers..."
	@docker compose -f tests/docker-compose-performance.yml down -v

install: compile
	@echo "=== $(INTEGRATION) === [ install ]: installing bin/$(BINARY_NAME)..."
	@sudo install -D --mode=755 --owner=root --strip $(ROOT)bin/$(BINARY_NAME) $(INTEGRATIONS_DIR)/bin/$(BINARY_NAME)
	@sudo install -D --mode=644 --owner=root $(ROOT)$(INTEGRATION)-config.yml.sample $(CONFIG_DIR)/$(INTEGRATION)-config.yml.sample

# rt-update-changelog runs the release-toolkit run.sh script by piping it into bash to update the CHANGELOG.md.
# It also passes down to the script all the flags added to the make target. To check all the accepted flags,
# see: https://github.com/newrelic/release-toolkit/blob/main/contrib/ohi-release-notes/run.sh
#  e.g. `make rt-update-changelog -- -v`
rt-update-changelog:
	curl "https://raw.githubusercontent.com/newrelic/release-toolkit/v1/contrib/ohi-release-notes/run.sh" | bash -s -- $(filter-out $@,$(MAKECMDGOALS))

# Include thematic Makefiles
include $(CURDIR)/build/ci.mk
include $(CURDIR)/build/release.mk

.PHONY: all build clean compile test integration-test install rt-update-changelog
