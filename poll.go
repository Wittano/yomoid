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
	pollListCommandName    = "list"
	pollRemoveCommandName  = "remove"
	pollPostCommandName    = "create"
)

func (p PollCommand) HandleSlashCommand(ctx context.Context, l *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	handler, ok := p[i.ApplicationCommandData().Options[0].Name]
	if !ok {
		return nil, fmt.Errorf("poll: unknown option %q", i.ApplicationCommandData().Options[0].Name)
	}

	return handler.HandleSlashCommand(ctx, l, s, i)
}

func NewPollCommand(db DatabaseQueries, handler *PollMessageCreateHandler) PollCommand {
	if handler == nil {
		panic("poll: missing poll message create handler")
	}

	return map[string]DiscordSlashCommandHandler{
		pollDetailsCommandName: PollDetailsCommand{Db: db},
		pollListCommandName:    PollListCommand{Db: db},
		pollRemoveCommandName:  PollRemoveCommand{Db: db},
		pollPostCommandName:    PollPostCommand{Db: db, PollMessageHandler: handler},
	}
}

type PollPostCommand struct {
	Db                 DatabaseQueries
	PollMessageHandler *PollMessageCreateHandler
}

func (p PollPostCommand) HandleSlashCommand(ctx context.Context, l *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	textChannel := i.ApplicationCommandData().Options[0].Options[1].ChannelValue(s)
	if textChannel == nil {
		return nil, fmt.Errorf("poll: missing channel argument")
	}

	pollID := int64(i.ApplicationCommandData().Options[0].Options[0].FloatValue())

	l.InfoContext(ctx, "valid poll post request received", "requestPollID", pollID, "requestChannelID", textChannel.ID)

	poll, err := p.Db.FindPoll(ctx, i.GuildID, pollID, "")
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return nil, DiscordMessageErr{
			error:       err,
			CommandName: "post",
			Msg:         "Invalid poll ID",
		}
	} else if err != nil {
		return nil, err
	}

	l.InfoContext(ctx, "poll found", "pollID", poll.ID, "pollQuestion", poll.Question)

	discordPoll, err := createDiscordPoll(poll)
	if err != nil {
		return nil, err
	}

	if _, err := s.ChannelMessageSendComplex(textChannel.ID, &discordgo.MessageSend{
		Poll: &discordPoll,
	}); err != nil {
		return nil, err
	}

	l.InfoContext(ctx, fmt.Sprintf("poll posted on channel #%s(%s)", textChannel.Name, textChannel.ID), "pollID", pollID)

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Poll #%d was created here", pollID),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}, nil
}

func createDiscordPoll(p Poll) (dp discordgo.Poll, err error) {
	answers := make([]discordgo.PollAnswer, len(p.Options))
	for i, opt := range p.Options {
		var (
			text, emoji string
		)
		textWithEmoji := strings.Split(opt, "  ")
		if len(textWithEmoji) == 1 {
			text = textWithEmoji[1]
		} else if len(textWithEmoji) == 2 {
			text = textWithEmoji[1]
			emoji = textWithEmoji[0]
		} else {
			err = errors.New("invalid poll option. Option cannot be empty")
			return
		}

		answers[i].Media = &discordgo.PollMedia{
			Text: text,
			Emoji: &discordgo.ComponentEmoji{
				Name: emoji,
			},
		}
	}

	return discordgo.Poll{
		Question: discordgo.PollMedia{
			Text: p.Question,
		},
		Answers:          answers,
		AllowMultiselect: p.IsMulti,
		Duration:         int(p.Duration),
	}, nil
}

type PollListCommand struct {
	Db DatabaseQueries
}

func (p PollListCommand) HandleSlashCommand(ctx context.Context, l *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	var title string

	if rawTitle, ok := parseInteractionInput(*i.Interaction)["title"]; !ok && (rawTitle != nil) {
		return nil, fmt.Errorf("poll: missing title argument")
	} else {
		title = rawTitle.(string)
	}

	polls, err := p.Db.FindAllPoll(ctx, i.GuildID, title, 0)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return nil, err
	} else if len(polls) == 0 {
		return CreateSimpleDiscordResponse("No found any poll with title: " + title), nil
	}

	return createPollDetails(ctx, nil, polls...), nil
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

func createPollDetails(ctx context.Context, user *discordgo.User, p ...Poll) *discordgo.InteractionResponse {
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

	pollCount := min(len(p), 10)
	embeds := make([]*discordgo.MessageEmbed, pollCount)
	for i, poll := range p[:pollCount] {
		for j, opt := range poll.Options {
			poll.Options[j] = fmt.Sprintf(" - %s", opt)
		}

		embeds[i] = &discordgo.MessageEmbed{
			Author:      &author,
			Color:       int(color),
			Title:       fmt.Sprintf("Poll **#%d**", poll.ID),
			Description: fmt.Sprintf("**Question**: %s\n**Duration**: %s\n**Options**:\n%s", poll.Question, time.Duration(int64(poll.Duration)*int64(time.Hour)), strings.Join(poll.Options, "\n")),
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Created at: %s", poll.CreatedAt.Time.Format(time.RFC822)),
			},
		}
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	}
}

type PollRemoveCommand struct {
	Db DatabaseQueries
}

func (p PollRemoveCommand) HandleSlashCommand(ctx context.Context, l *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	var id int64

	rawId, ok := parseInteractionInput(*i.Interaction)["id"]
	if !ok {
		return nil, fmt.Errorf("poll: missing id argument")
	} else {
		id = int64(rawId.(float64))
	}

	if err := p.Db.DeletePoll(ctx, id); err != nil {
		return nil, DiscordMessageErr{
			error:       err,
			CommandName: "remove",
			Msg:         "Invalid poll ID",
		}
	}

	l.InfoContext(ctx, "poll removed", "pollID", id)

	return CreateSimpleDiscordResponse("Poll removed"), nil
}

func CreatePollDetailsCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "poll",
		Description: "Manage poll",
		Type:        discordgo.ChatApplicationCommand,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        pollDetailsCommandName,
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
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        pollListCommandName,
				Description: "Show details of specific poll by question",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "title",
						Required:    true,
						Description: "Find polls by specific name",
					},
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "page",
						Description: "Page number",
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        pollRemoveCommandName,
				Description: "Remove poll by id",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "id",
						Required:    true,
						Description: "Poll's ID",
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        pollPostCommandName,
				Description: "Post poll from template",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "id",
						Required:    true,
						Description: "Poll's ID",
					},
					{
						Type: discordgo.ApplicationCommandOptionChannel,
						Name: "channel",
						ChannelTypes: []discordgo.ChannelType{
							discordgo.ChannelTypeGuildText,
						},
						Required:    true,
						Description: "Text channel where post will be posted",
					},
				},
			},
		},
	}
}
