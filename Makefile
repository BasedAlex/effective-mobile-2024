all: compile

docker_up:
	@echo Starting Docker images..
	docker-compose up -d 
	@echo Docker images started

## down: stop docker compose
docker_down:
	@echo Stopping docker compose 
	docker-compose down

compile:
	echo "Compiling for linux"
	cd cmd/api && GOOS=linux GOARCH=amd64 go build -o ../../cars.elf .

run:
	./cars.elf

up: compile run

run_migrations:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	cd internal/migrations && goose postgres "user=postgres password=password host=localhost port=5439 dbname=efmobile sslmode=disable" up

down_migrations: 
	cd internal/migrations && goose postgres "user=postgres password=password host=localhost port=5439 dbname=efmobile sslmode=disable" down

first_run: docker_up run_migrations up