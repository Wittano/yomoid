package main

import (
	"github.com/bwmarrin/discordgo"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
)

func closeAndLog(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Println(err)
	}
}

var bot *discordgo.Session

func main() {
	token, ok := os.LookupEnv("DISCORD_TOKEN")
	if !ok {
		log.Fatal("Missing required environment variable: DISCORD_TOKEN")
	}

	var err error
	bot, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}
	defer closeAndLog(bot)

	bot.AddHandler(nineGagMessageFixer)
	bot.AddHandler(ready)

	bot.Identify.Intents = discordgo.IntentMessageContent | discordgo.IntentGuildMessages

	if err = bot.Open(); err != nil {
		log.Fatal(err)
	}

	closeCh := make(chan os.Signal, 1)
	signal.Notify(closeCh, os.Interrupt)
	<-closeCh
}

func ready(_ *discordgo.Session, _ *discordgo.Ready) {
	slog.Info("Bot is ready. Press CTRL+C to exit.")
}
