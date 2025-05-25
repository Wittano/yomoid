package main

import (
	"context"
	"github.com/go-faster/errors"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

var Poll *pgxpool.Pool

func InitDb(ctx context.Context) (err error) {
	url, ok := os.LookupEnv("DATABASE_URL")
	if url != "" && !ok {
		return errors.New("database: Database URL is not set. Please set DATABASE_URL environment variable.")
	}

	Poll, err = pgxpool.New(ctx, url)
	return
}

func ParseString(s string) (p pgtype.Text) {
	p.Valid = len(s) > 0
	p.String = s
	return
}
