# Variables
DOCKER_IMAGE_NAME=your-go-app
DOCKER_COMPOSE_FILE=docker-compose.yml

# Default target
.PHONY: all
all: build

# Build the Docker image
.PHONY: build
build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE_NAME) .

# Run the application using Docker Compose
.PHONY: run
run: build
	@echo "Starting application with Docker Compose..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) up --build -d

# Stop the application
.PHONY: stop
stop:
	@echo "Stopping application..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

# Clean up containers, networks, and volumes
.PHONY: clean
clean:
	@echo "Cleaning up..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) down -v --remove-orphans

# Tail logs from the application
.PHONY: logs
logs:
	@echo "Tailing logs..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

# Rebuild and restart the application
.PHONY: rebuild
rebuild: clean run

# Show status of running containers
.PHONY: status
status:
	@echo "Showing container status..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) ps
