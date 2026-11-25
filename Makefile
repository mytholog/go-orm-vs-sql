.PHONY: deps generate up migrate seed test bench profile down
up:
	docker compose up -d --wait
migrate:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U demo -d demo < db/migrations/001.sql
seed:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U demo -d demo < db/seed.sql
deps:
	go get github.com/jackc/pgx/v5@latest github.com/jmoiron/sqlx@latest github.com/Masterminds/squirrel@latest gorm.io/gorm@latest gorm.io/driver/postgres@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	$(MAKE) generate
	go mod tidy
generate:
	$$(go env GOPATH)/bin/sqlc generate
test:
	go test ./...
bench:
	go test ./benchmarks -run '^$$' -bench . -benchmem -count=5
profile:
	go test ./benchmarks -run '^$$' -bench 'BenchmarkGetUser/gorm' -cpuprofile cpu.out -memprofile mem.out
demo-%:
	go run ./cmd/demo $*
plans:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U demo -d demo < db/plans.sql
down:
	docker compose down
