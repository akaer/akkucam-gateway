.PHONY: test lint test-e2e test-hardware build up down sbom

COMMIT := $(shell git rev-parse HEAD 2>/dev/null)

test:
	./scripts/test.sh

lint:
	./scripts/lint.sh

test-e2e:
	./scripts/test-e2e.sh

test-hardware:
	./scripts/test-hardware.sh

build:
	COMMIT="$(COMMIT)" docker compose build

up:
	COMMIT="$(COMMIT)" docker compose up -d --build --wait --wait-timeout 60

down:
	docker compose down

sbom:
	./scripts/sbom.sh
