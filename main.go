package main

import (
	"context"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	if err = InitDb(ctx); err != nil {
		slog.ErrorContext(ctx, "failed init database", "error", err)
		os.Exit(1)
	}

	bot, err = discordgo.New("Bot " + token)
	if err != nil {
		slog.ErrorContext(ctx, "failed create discord session", "error", err)
		os.Exit(1)
	}
	defer closeAndLog(bot)

	bot.AddHandler(nineGagMessageFixer)
	bot.AddHandler(createPollFromMessage)
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
