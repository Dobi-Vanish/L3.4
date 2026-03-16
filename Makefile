.PHONY: build run migrate

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

docker-build:
	docker-compose -f deployments/docker-compose.yml build

docker-up:
	docker-compose -f deployments/docker-compose.yml up -d

docker-down:
	docker-compose -f deployments/docker-compose.yml down

all: build docker-build docker-up