SQLC_VERSION := v1.31.1
PSQL = docker compose exec -T postgres psql -U demo -d demo

.PHONY: deps generate up migrate seed test test-integration bench profile plans down
up:
	docker compose up -d --wait
deps:
	go mod download
generate:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) generate
migrate:
	$(PSQL) -v ON_ERROR_STOP=1 --single-transaction -f - < db/migrations/001.sql
seed:
	$(PSQL) -v ON_ERROR_STOP=1 --single-transaction -f - < db/seed.sql
test:
	go test ./...
test-integration:
	go test ./... -tags integration -count=1
bench:
	go test ./benchmarks -run '^$$' -bench . -benchmem -count=5
profile:
	go test ./benchmarks -run '^$$' -bench 'BenchmarkGetUser/gorm$$' -cpuprofile cpu.out -memprofile mem.out
demo-%:
	go run ./cmd/demo $*
plans:
	$(PSQL) -v ON_ERROR_STOP=1 --single-transaction -f - < db/plans.sql
down:
	docker compose down
