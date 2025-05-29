package main

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"math"
	"time"
)

type PollMessageCreateHandler struct {
	db DatabaseQueries
}

func (p PollMessageCreateHandler) Handler(_ *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Poll == nil || m.Poll.Question.Text == "" || len(m.Poll.Answers) == 0 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := createLoggerFromMessage(ctx, *m.Message).
		With("pollName", m.Poll.Question.Text)

	logger.InfoContext(ctx, "PollMessageCreate handler received a new poll")

	pollId, err := p.db.CreatePoll(ctx, createPoll(*m))
	if err != nil {
		logger.Error("failed create a new poll", "error", err)
		return
	}

	logger.Info("poll created a new poll", "pollId", pollId)
}

func createPoll(msg discordgo.MessageCreate) (poll CreatePollParams) {
	duration := math.Ceil(msg.Poll.Expiry.Sub(time.Now()).Hours())
	if duration == 0 {
		duration = 24
	}

	poll.GuildID = msg.GuildID
	poll.Duration = int16(duration)
	poll.IsMulti = msg.Poll.AllowMultiselect
	poll.AuthorID = msg.Author.ID
	poll.Question = msg.Poll.Question.Text
	poll.Answers = make([]AnswerParams, len(msg.Poll.Answers))

	for i, a := range msg.Poll.Answers {
		if a.Media.Emoji != nil {
			poll.Answers[i].Emoji = a.Media.Emoji.Name
		}
		poll.Answers[i].Text = a.Media.Text
	}

	return
}
