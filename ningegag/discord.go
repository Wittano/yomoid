package ningegag

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/wittano/yomoid/logger"
	"regexp"
	"strings"
)

var (
	nineGagRegex    = regexp.MustCompile(`^(https://img-9gag-fun)([\w/.]*)_460sv([a-z0-9]{3}).([a-z0-9]{3,4})$`)
	fixNineGagRegex = regexp.MustCompile(`_460sv([a-z0-9]{3})`)
)

func MessageFixer(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l := logger.CreateLoggerFromMessage(ctx, s, *m.Message)

	words := strings.Split(m.Message.Content, " ")
	if len(words) == 0 && hasNineGagLink(words) {
		l.WarnContext(ctx, "missing links from 9gag's form to fixing")
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
			l.Error("failed send 9gag fixed links message", "error", err)
		}
	}
}

func hasNineGagLink(words []string) bool {
	for _, link := range words {
		if nineGagRegex.MatchString(link) {
			return true
		}
	}

	return false
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
