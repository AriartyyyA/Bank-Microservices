auth-migrations:
	migrate create -ext sql -dir migrations/auth -seq create_users_table

migrate_auth:
	migrate \
		-path migrations/auth \
		-database "postgres://gobank:secret@localhost:5432/gobank_auth?sslmode=disable" \
		up

migrate_wallet:
	migrate \
  		-path migrations/wallet \
  		-database "postgres://gobank:secret@localhost:5433/gobank_wallet?sslmode=disable" \
  		up 1

migrate_wallet_down:
	migrate \
  		-path migrations/wallet \
  		-database "postgres://gobank:secret@localhost:5433/gobank_wallet?sslmode=disable" \
  		down 

migrate_wallet_force:
	migrate \
  		-path migrations/wallet \
  		-database "postgres://gobank:secret@localhost:5433/gobank_wallet?sslmode=disable" \
  	force 0

wallet_migrations:
	migrate create -ext sql -dir migrations/wallet -seq create_wallet_tables

run-auth:
	go run cmd/auth/main.go