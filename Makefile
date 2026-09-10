CURRENT_DIR=$(shell pwd)

APP=$(shell basename ${CURRENT_DIR})

APP_CMD_DIR=${CURRENT_DIR}/cmd

TAG=latest
ENV_TAG=latest
PROJECT_NAME=bazaar

run:
	go run ./cmd/api/main.go

build:
	CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -o ${CURRENT_DIR}/bin/${APP} ${APP_CMD_DIR}/main.go

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

up:
	docker compose up 

down:
	docker compose down

migrate-up:
	docker run --mount type=bind,source="${CURRENT_DIR}/migrations,target=/migrations" --network ${NETWORK_NAME} migrate/migrate \
		-path=/migrations/ -database=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable up

migrate-down:
	docker run --mount type=bind,source="${CURRENT_DIR}/migrations,target=/migrations" --network ${NETWORK_NAME} migrate/migrate \
		-path=/migrations/ -database=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable down
