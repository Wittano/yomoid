package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"strings"
	"time"
)

type PollCommand map[string]DiscordSlashCommandHandler

const (
	pollDetailsCommandName = "details"
)

func (p PollCommand) HandleSlashCommand(ctx context.Context, l *slog.Logger, s *discordgo.Session, m *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	handler, ok := p[m.ApplicationCommandData().Options[0].Name]
	if !ok {
		return nil, fmt.Errorf("poll: unknown option %q", m.ApplicationCommandData().Options[0].Name)
	}

	return handler.HandleSlashCommand(ctx, l, s, m)
}

func NewPollCommand(db DatabaseQueries) PollCommand {
	return map[string]DiscordSlashCommandHandler{
		pollDetailsCommandName: PollDetailsCommand{Db: db},
	}
}

type PollDetailsCommand struct {
	Db DatabaseQueries
}

func (c PollDetailsCommand) HandleSlashCommand(ctx context.Context, l *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	id, title := findIdAndTitleInInteractionArgs(*i.Interaction)

	if id == 0 && title == "" {
		l.WarnContext(ctx, "missing id or title argument in poll details subcommand")

		return nil, DiscordMessageErr{
			Msg:         "Missing required poll id or title argument",
			CommandName: "poll-details",
		}
	}

	p, err := c.Db.FindPoll(ctx, i.GuildID, id, title)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return nil, DiscordMessageErr{
			error:       err,
			CommandName: "poll-details",
			Msg:         fmt.Sprintf("Poll with id %d or title %s not found", id, title),
		}
	} else if err != nil {
		return nil, err
	}

	l.InfoContext(ctx, "poll found", "pollID", p.ID, "pollQuestion", p.Question)

	u, err := s.User(p.AuthorID)
	if err != nil {
		l.WarnContext(ctx, "failed fetch user", "error", err)
	}

	return createPollDetails(ctx, u, p), nil
}

func findIdAndTitleInInteractionArgs(i discordgo.Interaction) (id int64, title string) {
	args := parseInteractionInput(i)

	if rawId, ok := args["id"]; ok {
		id = int64(rawId.(float64))
	}

	if rawTitle, ok := args["title"]; ok {
		title = rawTitle.(string)
	}

	return
}

func createPollDetails(ctx context.Context, user *discordgo.User, p Poll) *discordgo.InteractionResponse {
	var (
		author discordgo.MessageEmbedAuthor
		color  uint32
	)
	if user != nil {
		author.IconURL = user.AvatarURL("")
		author.Name = user.GlobalName

		var err error
		r, err := downloadImage(author.IconURL)
		imgCtx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		if err != nil {
			color = 0x0
		} else if color, err = imageMainColor(imgCtx, r); err != nil {
			color = 0x0
		}
	}

	for i, opt := range p.Options {
		p.Options[i] = fmt.Sprintf(" - %s", opt)
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Author:      &author,
					Color:       int(color),
					Title:       fmt.Sprintf("Poll **#%d**", p.ID),
					Description: fmt.Sprintf("**Question**: %s\n**Duration**: %s\n**Options**:\n%s", p.Question, time.Duration(int64(p.Duration)*int64(time.Hour)), strings.Join(p.Options, "\n")),
					Footer: &discordgo.MessageEmbedFooter{
						Text: fmt.Sprintf("Created at: %s", p.CreatedAt.Time.Format(time.RFC822)),
					},
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}
}

func CreatePollDetailsCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "poll",
		Description: "Manage poll",
		Type:        discordgo.ChatApplicationCommand,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "details",
				Description: "Show poll details",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "id",
						Description: "Poll ID",
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "title",
						Description: "Find first Poll by specific name",
					},
				},
			},
		},
	}
}
