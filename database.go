package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wittano/yomoid/gen/database"
	"os"
)

type Poll struct {
	ID        int64
	Question  string
	GuildID   string
	AuthorID  string
	IsMulti   bool
	Duration  int16
	CreatedAt pgtype.Timestamptz
	Options   []string
}

type DatabaseQueries interface {
	FindPoll(ctx context.Context, guildID string, id int64, title string) (Poll, error)
	CreatePoll(ctx context.Context, params CreatePollParams) (int64, error)
}

type Database struct {
	poll *pgxpool.Pool
}

func (d Database) FindPoll(ctx context.Context, guildID string, id int64, title string) (poll Poll, err error) {
	if id > 0 && title != "" {
		var p database.FindPollByIdAndQuestionRow
		p, err = database.New(d.poll).FindPollByIdAndQuestion(ctx, database.FindPollByIdAndQuestionParams{Column1: title, ID: id, GuildID: guildID})
		if err != nil {
			return
		}
		poll = Poll{
			ID:        p.ID,
			Question:  p.Question,
			GuildID:   p.GuildID,
			AuthorID:  p.AuthorID,
			IsMulti:   p.IsMulti,
			Duration:  p.Duration,
			CreatedAt: p.CreatedAt,
			Options:   p.Options,
		}
	} else if id > 0 {
		var p database.FindPollByIDRow
		p, err = database.New(d.poll).FindPollByID(ctx, database.FindPollByIDParams{ID: id, GuildID: guildID})
		if err != nil {
			return
		}
		poll = Poll{
			ID:        p.ID,
			Question:  p.Question,
			GuildID:   p.GuildID,
			AuthorID:  p.AuthorID,
			IsMulti:   p.IsMulti,
			Duration:  p.Duration,
			CreatedAt: p.CreatedAt,
			Options:   p.Options,
		}
	} else if title != "" {
		var p database.FindPollByQuestionRow
		p, err = database.New(d.poll).FindPollByQuestion(ctx, database.FindPollByQuestionParams{Column1: title, GuildID: guildID})
		if err != nil {
			return
		}
		poll = Poll{
			ID:        p.ID,
			Question:  p.Question,
			GuildID:   p.GuildID,
			AuthorID:  p.AuthorID,
			IsMulti:   p.IsMulti,
			Duration:  p.Duration,
			CreatedAt: p.CreatedAt,
			Options:   p.Options,
		}
	} else {
		return poll, errors.New("database: Missing required poll id or title argument")
	}

	return
}

type AnswerParams struct {
	Text  string
	Emoji string
}

type CreatePollParams struct {
	Question string
	GuildID  string
	AuthorID string
	Duration int16
	IsMulti  bool
	Answers  []AnswerParams
}

func (d Database) CreatePoll(ctx context.Context, params CreatePollParams) (int64, error) {
	tx, err := d.poll.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback(ctx))
		} else {
			err = errors.Join(err, tx.Commit(ctx))
		}
	}()

	q := database.New(tx)
	pollID, err := q.CreatePoll(ctx, database.CreatePollParams{
		Question: params.Question,
		GuildID:  params.GuildID,
		AuthorID: params.AuthorID,
		Duration: params.Duration,
		IsMulti:  params.IsMulti,
	})
	if err != nil {
		return 0, err
	}

	for i, a := range params.Answers {
		if a.Text == "" {
			return 0, fmt.Errorf("database: answer %d is empty string", i)
		}

		answer := database.CreatePollOptionParams{
			Answer: a.Text,
			Emoji:  ParseString(a.Emoji),
			PollID: pollID,
		}

		if err = q.CreatePollOption(ctx, answer); err != nil {
			return 0, err
		}
	}

	return pollID, nil
}

func NewDatabase(ctx context.Context) (*Database, error) {
	url, ok := os.LookupEnv("DATABASE_URL")
	if url != "" && !ok {
		return nil, errors.New("database: Database URL is not set. Please set DATABASE_URL environment variable.")
	}

	var (
		err error
		db  = new(Database)
	)
	db.poll, err = pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}

	if err = db.poll.Ping(ctx); err != nil {
		err = errors.Join(err, errors.New("database: failed to ping database"))
	}

	return db, err
}

func ParseString(s string) (p pgtype.Text) {
	p.Valid = len(s) > 0
	p.String = s
	return
}
