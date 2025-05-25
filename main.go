package main

import (
	"context"
	"flag"
	"github.com/bwmarrin/discordgo"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
)

func closeAndLog(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Println(err)
	}
}

var (
	level   = flag.String("level", "", "Log level")
	verbose = flag.Bool("verbose", false, "Verbose mode")

	bot *discordgo.Session
)

func init() {
	flag.Parse()
}

func main() {
	slog.SetLogLoggerLevel(parseLogLevel())

	token, ok := os.LookupEnv("DISCORD_TOKEN")
	if !ok {
		slog.Error("Missing required environment variable: DISCORD_TOKEN")
		os.Exit(1)
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

func IsVerbose() bool {
	return verbose != nil && *verbose
}

func parseLogLevel() slog.Level {
	if verbose != nil && *verbose {
		return slog.LevelDebug
	}

	if level == nil {
		return slog.LevelInfo
	}

	switch strings.ToUpper(*level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func ready(_ *discordgo.Session, _ *discordgo.Ready) {
	slog.Info("Bot is ready. Press CTRL+C to exit.")
}
