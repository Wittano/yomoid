package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/wittano/yomoid/gen/database"
	"math"
	"time"
)

func createPollFromMessage(_ *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Poll.Question.Text == "" || len(m.Poll.Answers) == 0 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := createDiscordHandlerLogger(ctx, *m.Message).
		With("pollName", m.Poll.Question.Text)
	LogDiscordMessage(ctx, logger, *m.Message)

	logger.InfoContext(ctx, "PollMessageCreate handler received a new poll")

	pollId, err := createPoll(ctx, *m)
	if err != nil {
		logger.Error("failed create a new poll", "error", err)
		return
	}

	logger.Info("poll created a new poll", "pollId", pollId)
}

func createPoll(ctx context.Context, msg discordgo.MessageCreate) (int64, error) {
	tx, err := Poll.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	query := database.New(tx)
	pollID, err := createPollEntity(ctx, query, msg)
	if err != nil {
		return 0, err
	}

	if err = assignPollOptions(ctx, query, pollID, msg); err != nil {
		return 0, err
	}

	return pollID, tx.Commit(ctx)
}

func assignPollOptions(ctx context.Context, q *database.Queries, pollID int64, msg discordgo.MessageCreate) error {
	for i, a := range msg.Poll.Answers {
		var question database.CreatePollOptionParams

		if a.Media.Text == "" {
			return fmt.Errorf("poll: answer #%d is empty", i+1)
		}

		if a.Media.Emoji != nil {
			question.Emoji = ParseString(a.Media.Emoji.Name)
		}
		question.Answer = a.Media.Text
		question.PollID = pollID

		if err := q.CreatePollOption(ctx, question); err != nil {
			return err
		}
	}

	return nil
}

func createPollEntity(ctx context.Context, q *database.Queries, msg discordgo.MessageCreate) (int64, error) {
	duration := math.Ceil(msg.Poll.Expiry.Sub(time.Now()).Hours())
	if duration == 0 {
		duration = 24
	}

	poll := database.CreatePollParams{
		GuildID:  msg.GuildID,
		Duration: int16(duration),
		IsMulti:  msg.Poll.AllowMultiselect,
		AuthorID: msg.Author.ID,
		Question: msg.Poll.Question.Text,
	}

	return q.CreatePoll(ctx, poll)
}
