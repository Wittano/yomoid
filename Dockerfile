from golang as builder

copy . .

run go build -o /app/yomoid cmd/yomoid/main.go

from busybox

workdir /app

copy --from=builder /app/* /app

entrypoint ["./app/yomoid"]