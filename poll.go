package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/wittano/yomoid/gen/database"
	"log/slog"
	"math"
	"time"
)

type pollMediaObject struct {
	Text  string `json:"text"`
	Emoji *struct {
		Name string `json:"name"`
		Id   string `json:"id"`
	} `json:"emoji"`
}

type discordPoll struct {
	Question   pollMediaObject `json:"question"`
	LayoutType int             `json:"layout_type"`
	Expiry     time.Time       `json:"expiry"`
	Answers    []struct {
		PollMedia pollMediaObject `json:"poll_media"`
		AnswerId  int             `json:"answer_id"`
	} `json:"answers"`
	AllowMultiselect bool `json:"allow_multiselect"`
}

func (p discordPoll) IsValid() bool {
	return p.Question.Text != "" && len(p.Answers) > 0
}

type pollCreateMessage struct {
	GuildID   string         `json:"guild_id"`
	Author    discordgo.User `json:"author"`
	ChannelId string         `json:"channel_id"`
	Poll      discordPoll    `json:"poll"`
	ID        string         `json:"id"`
}

var (
	errInvalidPollMessageType = errors.New("poll: event type is not MESSAGE_CREATE")
	errEventMissing           = errors.New("poll: event is nil")
	errPollMising             = errors.New("poll: discord poll data is nil")
)

func createPollFromMessage(_ *discordgo.Session, e *discordgo.Event) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msg, err := parsePollMessage(e)
	if errors.Is(err, errEventMissing) || errors.Is(err, errInvalidPollMessageType) || errors.Is(err, errPollMising) {
		return
	} else if err != nil {
		slog.Error("failed parse poll message", "error", err)
		return
	}

	logger := createDiscordHandlerLogger(ctx, discordgo.Message{
		GuildID:   msg.GuildID,
		ChannelID: msg.ChannelId,
		Author:    &msg.Author,
		ID:        msg.ID,
	})

	pollId, err := createPoll(ctx, msg)
	if err != nil {
		logger.Error("failed create a new poll",
			"pollName", msg.Poll.Question.Text,
			"error", err)
		return
	}

	logger.Info("poll created a new poll", "pollName", msg.Poll.Question.Text, "pollId", pollId)
}

func createPoll(ctx context.Context, msg pollCreateMessage) (int64, error) {
	duration := math.Ceil(msg.Poll.Expiry.Sub(time.Now()).Hours())
	if duration == 0 {
		duration = 24
	}

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

	poll := database.CreatePollParams{
		GuildID:  msg.GuildID,
		Duration: int16(duration),
		IsMulti:  msg.Poll.AllowMultiselect,
		AuthorID: msg.Author.ID,
		Question: msg.Poll.Question.Text,
	}

	pollId, err := query.CreatePoll(ctx, poll)
	if err != nil {
		err = errors.Join(err, tx.Rollback(ctx))
		return 0, err
	}

	questions := make([]database.CreatePollOptionParams, len(msg.Poll.Answers))
	for i, a := range msg.Poll.Answers {
		if a.PollMedia.Text == "" {
			return 0, fmt.Errorf("poll: answer #%d is empty", i+1)
		}

		if a.PollMedia.Emoji != nil {
			questions[i].Emoji = ParseString(a.PollMedia.Emoji.Name)
		}
		questions[i].Answer = a.PollMedia.Text
		questions[i].PollID = pollId

		if err = query.CreatePollOption(ctx, questions[i]); err != nil {
			err = errors.Join(err, tx.Rollback(ctx))
			return 0, err
		}
	}

	return pollId, tx.Commit(ctx)
}

func parsePollMessage(e *discordgo.Event) (msg pollCreateMessage, err error) {
	if e == nil {
		err = errEventMissing
		return
	}
	if e.Type != "MESSAGE_CREATE" {
		err = errInvalidPollMessageType
		return
	}

	err = json.Unmarshal(e.RawData, &msg)
	if !msg.Poll.IsValid() && err == nil {
		err = errPollMising
		return
	}

	return
}
