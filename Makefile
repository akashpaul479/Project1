# Load environment variables
include .env

#Project Name 
APP_NAME=college-management

#Docker Compose Command
DC=docker compose

#Default target
.PHONY: help
help:
	@echo "Available commands:"
	@echo " make up           -> start containers "
	@echo " make down         -> Stop containers "
	@echo " make rebuild      -> Rebuild & start "
	@echo " make restart      -> Restart app "
	@echo " make logs         -> Show app logs "
	@echo " make mysql        -> Open MySQL shell "
	@echo " make mongo        -> Open Mongo shell "
	@echo " make redis        -> Open Redis CLI  "
	@echo " make test         -> Run Go tests"
	@echo " make clean        -> Remove volumes  "

# Start containers
.PHONY: up
up:
	$(DC) up -d

# Stop containers
.PHONY: down
down:
	$(DC) down

# Rebuild and start
.PHONY: rebuild
rebuild:
	$(DC) up --build -d

# Restart app
.PHONY: restart
restart:
	$(DC) restart app


# View logs
.PHONY: logs
logs:
	$(DC) logs -f app


# MySQL Shell
.PHONY: mysql
mysql:
	docker exec -it college_mysql mysql -uroot -p$(MYSQL_ROOT_PASSWORD)


# Mongo Shell 
.PHONY: mongo
mongo:
	docker exec -it college_mongo mongosh

# Redis Shell 
.PHONY: redis
redis:
	docker exec -it college_redis redis-cli


# Run tests
.PHONY: test
test:
	go test ./...

# Clean everything (Warning: deletes DB)
.PHONY: clean
clean:
	$(DC) down -v










