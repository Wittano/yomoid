package main

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

func createDiscordHandlerLogger(ctx context.Context, m discordgo.Message) *slog.Logger {
	username := m.Author.Username
	if username == "" {
		username = "Unknown"
	}

	logger := slog.Default().
		With(slog.String("messageID", m.ID)).
		With(slog.String("authorID", m.Author.ID)).
		With(slog.String("authorName", username)).
		With(slog.String("channelID", m.ChannelID)).
		With(slog.String("guildID", m.GuildID))

	if !IsVerbose() {
		return logger
	}

	channel, err := bot.Channel(m.ChannelID, discordgo.WithContext(ctx))
	if err != nil {
		logger.DebugContext(ctx, "failed fetch channel", "error", err)
	} else {
		logger = logger.With(slog.String("channelName", channel.Name))
	}

	guild, err := bot.Guild(m.GuildID, discordgo.WithContext(ctx))
	if err != nil {
		logger.DebugContext(ctx, "failed fetch guild", "error", err)
	} else {
		logger = logger.With(slog.String("guildName", guild.Name))
	}

	return logger
}

func LogDiscordMessage(ctx context.Context, l *slog.Logger, m discordgo.Message) {
	if l == nil {
		panic("logger is nil")
	}

	l.DebugContext(ctx, "message received", "msg", m)
}
