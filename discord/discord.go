package discord

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/wittano/yomoid/logger"
	"github.com/wittano/yomoid/poll"
	"log/slog"
	"strings"
	"time"
)

type SlashCommandHandler interface {
	HandleSlashCommand(ctx context.Context, l *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error)
}

type MessageErr struct {
	error
	CommandName string
	Msg         string
}

func (e MessageErr) Error() string {
	var msg string
	if e.error != nil {
		msg = e.error.Error()
	} else {
		msg = strings.ToLower(e.Msg)
	}

	return fmt.Sprintf("discord slashCommand %s: %s", e.CommandName, msg)
}

var subCommandMap map[string]SlashCommandHandler

func InitSlashCommandList(db poll.Queries, handler *poll.MessageCreateHandler) {
	subCommandMap = map[string]SlashCommandHandler{
		"poll": NewPollCommand(db, handler),
	}
}

func HandleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	const slashCommandTimeout = time.Duration(float64(time.Second) * 2)
	ctx, cancel := context.WithTimeout(context.Background(), slashCommandTimeout)
	defer cancel()

	l := logger.NewLoggerFromInteraction(ctx, s, *i.Interaction).
		With("commandName", i.ApplicationCommandData().Name)

	handler, ok := subCommandMap[i.ApplicationCommandData().Name]
	if !ok {
		l.WarnContext(ctx, "unknown slash command")
		return
	}

	l.InfoContext(ctx, "slash command handler received a new command")

	res, err := handler.HandleSlashCommand(ctx, l, s, i)
	if err != nil {
		var (
			discordErr MessageErr
			content    = "Unexpected internal error. Try again later"
		)
		if errors.As(err, &discordErr) {
			content = discordErr.Msg
		}

		l.ErrorContext(ctx, "unexpected failed handle slash command", "error", err)
		res = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}
	}

	if err = s.InteractionRespond(i.Interaction, res); err != nil {
		l.ErrorContext(ctx, "failed send interaction respond to slash command", "error", err)
	} else {
		l.Info("response for interaction")
	}
}

func parseInteractionInput(i discordgo.Interaction) (in map[string]any) {
	if len(i.ApplicationCommandData().Options) == 0 {
		return nil
	}

	in = make(map[string]any, len(i.ApplicationCommandData().Options[0].Options))

	for _, data := range i.ApplicationCommandData().Options[0].Options {
		if data == nil {
			continue
		}

		in[data.Name] = data.Value
	}

	return
}

func CreateSimpleDiscordResponse(msg string) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}
}
