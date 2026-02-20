.PHONY: build run test clean migrate-up migrate-down docker-up docker-down

APP_NAME=movie-recommend
MAIN_PATH=./cmd/api

build:
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)

test:
	go test -v -cover ./...

clean:
	rm -rf bin/

lint:
	golangci-lint run ./...

migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/movie_recommend?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/movie_recommend?sslmode=disable" down

docker-up:
	docker-compose -f docker/docker-compose.yml up --build -d

docker-down:
	docker-compose -f docker/docker-compose.yml down
