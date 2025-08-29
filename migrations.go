package yomoid

import (
	"database/sql"
	"embed"
	"os"

	"github.com/pressly/goose/v3"
	"github.com/wittano/yomoid/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed database/migrations/*.sql
var migrations embed.FS

func MigrateDatabase(dbURL string) (err error) {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return err
	}
	defer logger.LogCloser(db)

	if os.Getenv("GOOSE_DBSTRING") == "" {
		err = os.Setenv("GOOSE_DBSTRING", dbURL)
	}
	if err != nil {
		return err
	}

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, "database/migrations")
}
