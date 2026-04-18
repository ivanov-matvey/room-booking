.PHONY: up down seed test test-unit test-e2e coverage mock swagger lint

up:
	docker-compose up -d --build

down:
	docker-compose down -v

seed:
	cat internal/db/seed_realistic.sql | docker-compose exec -T postgres psql -U postgres -d roomdb

mock:
	go run github.com/vektra/mockery/v2@v2.53.6

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/main.go -o docs --pdl 1

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run

test:
	go test ./...

test-unit:
	go test ./internal/...

coverage:
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -func=coverage.out

test-e2e:
	docker-compose -f docker-compose.e2e.yaml up -d --build
	@echo "Waiting for server to be ready..."
	@until docker-compose -f docker-compose.e2e.yaml exec -T app_e2e wget -qO- http://localhost:8080/_info > /dev/null 2>&1; do sleep 2; done
	TEST_SERVER_URL=http://localhost:8081 go test -v -tags=integration ./tests/... || true
	docker-compose -f docker-compose.e2e.yaml down -v
