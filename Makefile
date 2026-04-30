auth-migrations:
	migrate create -ext sql -dir migrations/auth -seq create_users_table

migrate_auth:
	migrate \
		-path migrations/auth \
		-database "postgres://gobank:secret@localhost:5432/gobank_auth?sslmode=disable" \
		up

run-auth:
	go run cmd/auth/main.go