.PHONY: server DatabaseInteractive migrationsUp migrationsDown

DB_URI := postgres://fraud_db:fraud_db@localhost:5433/fraud_db?sslmode=disable

server:
	@go build -o fraud-detection-lite ./cmd && ./fraud-detection-lite

databaseInteractive:
	@docker exec -it fraud-detection-lite-postgres-1 psql -U fraud_db -d fraud_db

migrationsUp:
	@goose -dir internal/db/migrations postgres $(DB_URI) up

migrationsDown:
	@goose -dir internal/db/migrations postgres $(DB_URI) down

redisInteractive:
	@docker exec -it fraud-detection-lite-redis-1 redis-cli