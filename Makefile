THRESHOLD ?= 50
RAW ?= coverage.raw.out
OUT ?= coverage.out
DATABASE_URL ?= postgres://cityhawk:cityhawk@localhost:5432/cityhawk?sslmode=disable

.PHONY: test coverage coverage-check clean db-schema db-seed db-reset

test:
	go test .

coverage:
	go test . -covermode=atomic -coverprofile=$(RAW)
	awk 'NR == 1 || ($$0 !~ /\/mocks?\// && $$0 !~ /_mock\.go/ && $$0 !~ /mock_.*\.go/ && $$0 !~ /\.mock\.go/ && $$0 !~ /place_store_seed\.go/)' $(RAW) > $(OUT)
	go tool cover -func=$(OUT)

db-schema:
	psql "$(DATABASE_URL)" -f db/migrations/0001_init.up.sql
	psql "$(DATABASE_URL)" -f db/migrations/0003_search_trgm.up.sql

db-seed:
	psql "$(DATABASE_URL)" -f db/migrations/0002_seed.up.sql

db-reset:
	psql "$(DATABASE_URL)" -f db/migrations/0002_seed.down.sql
	psql "$(DATABASE_URL)" -f db/migrations/0003_search_trgm.down.sql
	psql "$(DATABASE_URL)" -f db/migrations/0001_init.down.sql

clean:
	rm -f $(RAW) $(OUT)
