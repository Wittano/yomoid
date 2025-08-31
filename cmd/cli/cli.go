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
	update  = flag.Bool("update", false, "Update slash commands")
	token   = flag.String("token", "", "Discord bot token")
	appID   = flag.String("appID", "", "Discord bot application ID")
	guildID = flag.String("guildID", "", "Guild ID")
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

	readyCh := make(chan struct{})
	defer close(readyCh)
	bot.AddHandler(func(_ *discordgo.Session, msg *discordgo.Ready) {
		log.Println("Discord REST client is ready")

		readyCh <- struct{}{}
	})

	if err := bot.Open(); err != nil {
		log.Fatal(err)
	}

	<-readyCh

	if update != nil && *update {
		updateCommand(bot)
	}
}

func updateCommand(s *discordgo.Session) {
	if appID == nil || *appID == "" {
		log.Fatal("yomoid: missing required appID value")
	}

	pollCmd := discord.NewPollCommandDefinition()

	gID := ""
	if guildID != nil && *guildID != "" {
		gID = *guildID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := s.ApplicationCommandCreate(*appID, gID, pollCmd, discordgo.WithContext(ctx)); err != nil {
		log.Fatal(err)
	}

	log.Printf("yomoid: updated slash command for app '%s' on guild '%s'", *appID, gID)
}
