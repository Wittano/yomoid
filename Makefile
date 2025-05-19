staticcheck:
	go tool staticcheck ./...
	go vet ./...

sqlc:
	go tool sqlc -f ./database/sqlc.yml generate

run: sqlc staticcheck
	go run main.go