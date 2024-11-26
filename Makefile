
include .env

MIGRATIONS_DIR=internal/infra/db/schema/migrations

.PHONY: sqlc
sqlc:
	cd internal/infra/db && sqlc generate

.PHONY: migrate-up
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down

.PHONY: migrate-create
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name

.PHONY: migrate-force
migrate-force:
	@read -p "Enter version number: " version; \
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $$version

.PHONY: migrate-version
migrate-version:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

.PHONY: create-queue
create-queue:
	aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name todo-queue 

.PHONY: create-bucket
create-bucket:
	aws --endpoint-url=http://localhost:4566 s3api create-bucket --bucket todo-bucket --region us-east-1  

.PHONY: run
run:
	@echo "Starting services..."
	docker-compose up -d

	@echo "Waiting for services to be ready..."
	sleep 10

	@echo "Creating SQS queue..."
	docker-compose exec localstack aws --endpoint-url=http://localhost:4566 \
		sqs create-queue --queue-name todo-queue || true

	@echo "Creating S3 bucket..."
	docker-compose exec localstack aws --endpoint-url=http://localhost:4566 \
		s3api create-bucket --bucket todo-bucket --region us-east-1 || true

	@echo "Running migrations..."
	docker-compose exec app migrate -path /app/internal/infra/db/schema/migrations \
		-database "$(DATABASE_URL)" up

	@echo "Services are ready!"
	docker-compose logs -f app

.PHONY: stop
stop:
	docker-compose down -v

.PHONY: logs
logs:
	docker-compose logs -f

.PHONY: test
test:
	go test -v ./...