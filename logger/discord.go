package logger

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

var (
	channelNameCache = map[string]string{}
	guildNameCache   = map[string]string{}
)

func NewLoggerFromInteraction(ctx context.Context, s *discordgo.Session, i discordgo.Interaction) *slog.Logger {
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

	logger = appendGuildNameAttr(ctx, logger, s, i.GuildID)
	logger = appendChannelNameAttr(ctx, logger, s, i.ChannelID)

	return logger
}

func appendChannelNameAttr(ctx context.Context, l *slog.Logger, s *discordgo.Session, channelID string) *slog.Logger {
	if name, ok := channelNameCache[channelID]; ok && name != "" {
		return l.With(slog.String("channelName", name))
	}

	guild, err := s.Guild(channelID, discordgo.WithContext(ctx))
	if err != nil {
		l.DebugContext(ctx, "failed fetch guild", "error", err)
	} else {
		channelNameCache[channelID] = guild.Name
		return l.With(slog.String("channelName", guild.Name))
	}

	return l
}

func appendGuildNameAttr(ctx context.Context, l *slog.Logger, s *discordgo.Session, guidID string) *slog.Logger {
	if name, ok := guildNameCache[guidID]; ok && name != "" {
		return l.With(slog.String("guildName", name))
	}

	guild, err := s.Guild(guidID, discordgo.WithContext(ctx))
	if err != nil {
		l.DebugContext(ctx, "failed fetch guild", "error", err)
	} else {
		channelNameCache[guidID] = guild.Name
		return l.With(slog.String("guildName", guild.Name))
	}
	return l
}

func CreateLoggerFromMessage(ctx context.Context, s *discordgo.Session, m discordgo.Message) *slog.Logger {
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

	channel, err := s.Channel(m.ChannelID, discordgo.WithContext(ctx))
	if err != nil {
		logger.DebugContext(ctx, "failed fetch channel", "error", err)
	} else {
		logger = logger.With(slog.String("channelName", channel.Name))
	}

	guild, err := s.Guild(m.GuildID, discordgo.WithContext(ctx))
	if err != nil {
		logger.DebugContext(ctx, "failed fetch guild", "error", err)
	} else {
		logger = logger.With(slog.String("guildName", guild.Name))
	}

	return logger
}
