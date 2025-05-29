staticcheck:
	go tool govulncheck ./...
	go tool staticcheck ./...
	go vet ./...

sqlc:
	go tool sqlc -f ./database/sqlc.yml generate

build: sqlc staticcheck
	go build -o bin/yomoid main.go

tests: sqlc staticcheck
	go test ./...

run: sqlc staticcheck
	go run main.go --verbose --level DEBUG