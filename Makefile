.PHONY: test test-e2e test-hardware build up down sbom

test:
	./scripts/test.sh

test-e2e:
	./scripts/test-e2e.sh

test-hardware:
	./scripts/test-hardware.sh

build:
	docker compose build

up:
	docker compose up -d --build

down:
	docker compose down

sbom:
	./scripts/sbom.sh
