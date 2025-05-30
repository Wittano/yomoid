package poll

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/wittano/yomoid/logger"
	"math"
	"time"
)

type MessageCreateHandler struct {
	Db Queries
}

func (p MessageCreateHandler) Handler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Poll == nil || m.Poll.Question.Text == "" || len(m.Poll.Answers) == 0 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if p.Db.Exists(ctx, m.Poll.Question.Text, m.GuildID) {
		return
	}

	l := logger.CreateLoggerFromMessage(ctx, s, *m.Message).
		With("pollName", m.Poll.Question.Text)

	l.InfoContext(ctx, "PollMessageCreate handler received a new poll")

	pollId, err := p.Db.CreatePoll(ctx, createPoll(*m))
	if err != nil {
		l.Error("failed create a new poll", "error", err)
		return
	}

	l.Info("poll created a new poll", "pollId", pollId)

	if _, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Flags: discordgo.MessageFlagsEphemeral,
		Reference: &discordgo.MessageReference{
			Type:      discordgo.MessageReferenceTypeDefault,
			MessageID: m.ID,
			GuildID:   m.GuildID,
			ChannelID: m.ChannelID,
		},
		Content: fmt.Sprintf("I saved your poll %s. Model's id is `%d`", m.Poll.Question.Text, pollId),
	}); err != nil {
		l.ErrorContext(ctx, "failed response for creating a new poll in database", "error", err)
	}
}

func createPoll(msg discordgo.MessageCreate) (poll CreatePollParams) {
	duration := math.Ceil(time.Until(*msg.Poll.Expiry).Hours())
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
