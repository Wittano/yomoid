package main

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"regexp"
	"strings"
)

var (
	nineGagRegex    = regexp.MustCompile(`^(https://img-9gag-fun)([\w/.]*)_460sv([a-z0-9]{3}).([a-z0-9]{3,4})$`)
	fixNineGagRegex = regexp.MustCompile(`_460sv([a-z0-9]{3})`)
)

func nineGagMessageFixer(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := createDiscordHandlerLogger(ctx, *m.Message)

	words := strings.Split(m.Message.Content, " ")
	if len(words) == 0 {
		logger.WarnContext(ctx, "missing links from 9gag's form to fixing")
		return
	}

	sendMessage := false
	for i, link := range words {
		var fixed bool
		words[i], fixed = fixNinegagLink(link)
		if fixed {
			sendMessage = true
		}
	}

	if sendMessage {
		_, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Content: "I fixed your 9gag's links: " + strings.Join(words, " "),
			Reference: &discordgo.MessageReference{
				MessageID: m.Message.ID,
				ChannelID: m.ChannelID,
				GuildID:   m.GuildID,
			},
		}, discordgo.WithContext(ctx))
		if err != nil {
			logger.Error("failed send 9gag fixed links message", "error", err)
		}
	}
}

func fixNinegagLink(link string) (string, bool) {
	if nineGagRegex.MatchString(link) {
		s := fixNineGagRegex.ReplaceAllString(link, "_460sv")
		if s != "" {
			return s, true
		}
	}

	return link, false
}
