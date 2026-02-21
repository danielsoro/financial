.PHONY: db db-down db-reset run dev migrate frontend

db:
	docker compose up -d

db-down:
	docker compose down

db-reset:
	docker compose down -v
	docker compose up -d

run:
	cd backend && go run ./cmd/api

dev:
	cd backend && air

migrate:
	cd backend && go run ./cmd/api --migrate

frontend:
	cd frontend && npm run dev
