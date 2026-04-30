auth-migrations:
	migrate create -ext sql -dir migrations/auth -seq create_users_table