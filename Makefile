THRESHOLD ?= 50
RAW ?= coverage.raw.out
OUT ?= coverage.out

.PHONY: test coverage coverage-check clean

test:
	go test .

coverage:
	go test . -covermode=atomic -coverprofile=$(RAW)
	awk 'NR == 1 || ($$0 !~ /\/mocks?\// && $$0 !~ /_mock\.go/ && $$0 !~ /mock_.*\.go/ && $$0 !~ /\.mock\.go/ && $$0 !~ /place_store_seed\.go/)' $(RAW) > $(OUT)
	go tool cover -func=$(OUT)

clean:
	rm -f $(RAW) $(OUT)
