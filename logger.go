package main

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

var (
	channelNameCache = map[string]string{}
	guildNameCache   = map[string]string{}
)

func createLoggerFromInteraction(ctx context.Context, i discordgo.Interaction) *slog.Logger {
	var (
		user             = i.User
		userID, username = "unknown", "unknown"
	)
	if user != nil {
		userID = user.ID
		username = user.GlobalName
	}

	logger := slog.Default().
		With(slog.String("messageID", i.ID)).
		With(slog.String("authorID", userID)).
		With(slog.String("authorName", username)).
		With(slog.String("channelID", i.ChannelID)).
		With(slog.String("guildID", i.GuildID))

	if !IsVerbose() {
		return logger
	}

	appendGuildNameAttr(ctx, logger, i.GuildID)
	appendChannelNameAttr(ctx, logger, i.ChannelID)

	return logger
}

func appendChannelNameAttr(ctx context.Context, l *slog.Logger, channelID string) {
	if name, ok := channelNameCache[channelID]; ok && name != "" {
		l = l.With(slog.String("channelName", name))
		return
	}

	guild, err := bot.Guild(channelID, discordgo.WithContext(ctx))
	if err != nil {
		l.DebugContext(ctx, "failed fetch guild", "error", err)
	} else {
		l = l.With(slog.String("channelName", guild.Name))
		channelNameCache[channelID] = guild.Name
	}
}

func appendGuildNameAttr(ctx context.Context, l *slog.Logger, guidID string) {
	if name, ok := guildNameCache[guidID]; ok && name != "" {
		l = l.With(slog.String("guildName", name))
		return
	}

	guild, err := bot.Guild(guidID, discordgo.WithContext(ctx))
	if err != nil {
		l.DebugContext(ctx, "failed fetch guild", "error", err)
	} else {
		l = l.With(slog.String("guildName", guild.Name))
		channelNameCache[guidID] = guild.Name
	}
}

func createLoggerFromMessage(ctx context.Context, m discordgo.Message) *slog.Logger {
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
