.PHONY: db run dev migrate frontend

db:
	docker compose up -d

db-down:
	docker compose down

run:
	cd backend && go run ./cmd/api

dev:
	cd backend && air

migrate:
	cd backend && go run ./cmd/api --migrate

frontend:
	cd frontend && npm run dev
