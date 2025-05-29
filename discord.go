package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"strings"
	"time"
)

type DiscordSlashCommandHandler interface {
	HandleSlashCommand(ctx context.Context, l *slog.Logger, s *discordgo.Session, m *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error)
}

type DiscordMessageErr struct {
	error
	CommandName string
	Msg         string
}

func (e DiscordMessageErr) Error() string {
	var msg string
	if e.error != nil {
		msg = e.error.Error()
	} else {
		msg = strings.ToLower(e.Msg)
	}

	return fmt.Sprintf("discord slashCommand %s: %s", e.CommandName, msg)
}

var subCommandMap map[string]DiscordSlashCommandHandler

func initSlashCommandList(db DatabaseQueries) {
	subCommandMap = map[string]DiscordSlashCommandHandler{
		"poll": NewPollCommand(db),
	}
}

func handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	const slashCommandTimeout = time.Duration(float64(time.Second) * 2)
	ctx, cancel := context.WithTimeout(context.Background(), slashCommandTimeout)
	defer cancel()

	logger := createLoggerFromInteraction(ctx, *i.Interaction).
		With("commandName", i.ApplicationCommandData().Name)

	handler, ok := subCommandMap[i.ApplicationCommandData().Name]
	if !ok {
		logger.WarnContext(ctx, "unknown slash command")
		return
	}

	logger.InfoContext(ctx, "slash command handler received a new command")

	res, err := handler.HandleSlashCommand(ctx, logger, s, i)
	if err != nil {
		var (
			discordErr DiscordMessageErr
			content    = "Unexpected internal error. Try again later"
		)
		if errors.As(err, &discordErr) {
			content = discordErr.Msg
		}

		logger.ErrorContext(ctx, "unexpected failed handle slash command", "error", err)
		res = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}
	}

	if err = s.InteractionRespond(i.Interaction, res); err != nil {
		logger.ErrorContext(ctx, "failed send interaction respond to slash command", "error", err)
	} else {
		logger.Info("response for interaction")
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
