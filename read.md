## criação das migrations em go 
migrate create -ext sql -dir ./postgres/migration/ -seq users_table

## para rodar o sqlc 
sqlc generate

importante rodar sem sudo.