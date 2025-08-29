from golang as builder

copy . .

run go tool sqlc -f ./database/sqlc.yml generate

run go build -o /app/yomoid cmd/yomoid/main.go

from busybox

workdir /app

copy --from=builder /app/* /app

copy --from=builder /etc/ssl/certs /etc/ssl/certs

env GOOSE_DRIVER=postgres

entrypoint ["/app/yomoid"]
