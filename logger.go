package main

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

func createDiscordHandlerLogger(ctx context.Context, m discordgo.Message) *slog.Logger {
	logger := slog.Default().
		With(slog.String("messageId", m.ID)).
		With(slog.String("authorId", m.Author.ID)).
		With(slog.String("authorName", m.Author.GlobalName))

	guild, err := bot.Guild(m.GuildID, discordgo.WithContext(ctx))
	if err != nil {
		logger.Error("failed fetch guild", "error", err)
	} else {
		logger = logger.With(slog.String("guildName", guild.Name)).
			With(slog.String("guildName", m.GuildID))
	}

	channel, err := bot.Channel(m.ChannelID)
	if err != nil {
		logger.Error("failed fetch channel", "error", err)
	} else {
		logger = logger.With(slog.String("channelName", channel.Name)).
			With(slog.String("channelId", m.ChannelID))
	}

	return logger
}
