## criação das migrations em go 
migrate create -ext sql -dir ./postgres/migration/ -seq users_table

## para rodar o sqlc 
sqlc generate

importante rodar sem sudo.

## para jerar o jwt 
openssl rand --base64 32

o comando a cima vai gerar uma chave aleatoria no seu terminal.