
dev:
	go tool air

start: templates assets
	go run main.go

templates:
	go tool templ generate 

assets:
	yarn install
	yarn build

db_migrate:
	go run ./cmd/db migrate --db data.db

db_seed:
	go run ./cmd/db seed --db data.db
