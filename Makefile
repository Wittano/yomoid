staticcheck:
	go mod tidy
	go tool govulncheck ./...
	go tool staticcheck ./...
	go vet ./...

sqlc:
	go tool sqlc -f ./database/sqlc.yml generate

build: tests
	go build -o bin/yomoid cmd/yomoid/main.go

tests: sqlc staticcheck
	go test ./...

run: tests
	go run cmd/yomoid/main.go --level DEBUG