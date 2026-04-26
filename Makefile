SHELL := /bin/bash

THRESHOLD ?= 60
RAW ?= coverage.raw.out
OUT ?= coverage.out
TOOLCHAIN ?= go1.25.0
GO_CACHE_DIR ?= $(CURDIR)/.cache/go-build
GO_TMP_DIR ?= $(CURDIR)/.cache/gotmp
DATABASE_URL ?= postgres://cityhawk:cityhawk@localhost:5432/cityhawk?sslmode=disable

.PHONY: test coverage coverage-check clean db-schema db-seed db-reset

define GO_ENV
export GOCACHE="$(GO_CACHE_DIR)" GOTMPDIR="$(GO_TMP_DIR)" GOTOOLCHAIN="$(TOOLCHAIN)";
endef

test:
	@mkdir -p "$(GO_CACHE_DIR)" "$(GO_TMP_DIR)"
	@$(GO_ENV) go test ./...

coverage:
	@mkdir -p "$(GO_CACHE_DIR)" "$(GO_TMP_DIR)"
	@$(GO_ENV) coverpkgs=$$(go list ./... | grep -v '/cmd$$' | grep -v '/internal/mocks$$' | paste -sd ',' -); \
	go test -count=1 ./... -covermode=atomic -coverpkg="$$coverpkgs" -coverprofile="$(RAW)"; \
	cp "$(RAW)" "$(OUT)"; \
	go tool cover -func="$(OUT)"

coverage-check:
	@mkdir -p "$(GO_CACHE_DIR)" "$(GO_TMP_DIR)"
	@$(GO_ENV) coverpkgs=$$(go list ./... | grep -v '/cmd$$' | grep -v '/internal/mocks$$' | paste -sd ',' -); \
	go test -count=1 ./... -covermode=atomic -coverpkg="$$coverpkgs" -coverprofile="$(RAW)" >/dev/null; \
	cp "$(RAW)" "$(OUT)"; \
	total=$$(go tool cover -func="$(OUT)" | awk '/^total:/ {gsub("%","",$$3); print $$3}'); \
	echo "Total coverage: $$total%"; \
	awk -v total="$$total" -v threshold="$(THRESHOLD)" 'BEGIN { if (total + 0 < threshold + 0) exit 1 }'

db-schema:
	psql "$(DATABASE_URL)" -f db/migrations/0001_init.up.sql
	psql "$(DATABASE_URL)" -f db/migrations/0003_search_trgm.up.sql
	psql "$(DATABASE_URL)" -f db/migrations/0004_media_paths.up.sql
	psql "$(DATABASE_URL)" -f db/migrations/0005_support_tickets.up.sql
	psql "$(DATABASE_URL)" -f db/migrations/0006_support_ticket_messages.up.sql
	psql "$(DATABASE_URL)" -f db/migrations/0007_user_roles.up.sql

db-seed:
	psql "$(DATABASE_URL)" -f db/migrations/0002_seed.up.sql

db-reset:
	psql "$(DATABASE_URL)" -f db/migrations/0002_seed.down.sql
	psql "$(DATABASE_URL)" -f db/migrations/0007_user_roles.down.sql
	psql "$(DATABASE_URL)" -f db/migrations/0006_support_ticket_messages.down.sql
	psql "$(DATABASE_URL)" -f db/migrations/0005_support_tickets.down.sql
	psql "$(DATABASE_URL)" -f db/migrations/0004_media_paths.down.sql
	psql "$(DATABASE_URL)" -f db/migrations/0003_search_trgm.down.sql
	psql "$(DATABASE_URL)" -f db/migrations/0001_init.down.sql

clean:
	rm -f $(RAW) $(OUT)
