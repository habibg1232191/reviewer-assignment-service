APP_NAME = reviewer-service
DOCKER_COMPOSE_FILE = docker-compose.yml

build:
	go build -o $(APP_NAME) ./cmd

run: build
	./$(APP_NAME)

up:
	docker-compose -f $(DOCKER_COMPOSE_FILE) up --build

down:
	docker-compose -f $(DOCKER_COMPOSE_FILE) down -v

clean:
	docker-compose down -v

all: clean build test up
