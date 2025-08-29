package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/wittano/yomoid/discord"
	"github.com/wittano/yomoid/logger"
)

var (
	update  = flag.Bool("u", false, "Update slash commands")
	token   = flag.String("t", "", "Discord bot token")
	guildID = flag.String("g", "", "Guild ID")
)

func main() {
	flag.Parse()

	if token == nil || *token == "" {
		log.Fatal("missing required discord token")
	}

	bot, err := discordgo.New("Bot " + *token)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.LogCloser(bot)

	if update != nil && *update {
		updateCommand(bot)
	}
}

func updateCommand(s *discordgo.Session) {
	pollCmd := discord.NewPollCommandDefinition()

	gID := ""
	if guildID != nil && *guildID != "" {
		gID = *guildID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := s.ApplicationCommandCreate("", gID, pollCmd, discordgo.WithContext(ctx)); err != nil {
		log.Fatal(err)
	}
}
