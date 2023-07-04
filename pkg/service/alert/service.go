package alert

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgo/webhook"
	"github.com/disgoorg/snowflake/v2"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

type Color int

const (
	SuccessColor = Color(1752220)
	ErrorColor   = Color(15158332)
	WarnColor    = Color(15844367)
	InfoColor    = Color(0)
)

type Service interface {
	SendError(ctx context.Context, err error) error
	SendInfo(ctx context.Context, message string) error
}

func NewService(
	config *configs.Config,
) (Service, error) {
	logrus.WithField("discordWebhookID", config.DiscordWebhookID).Info("initiating alert service")
	service := serviceImpl{}
	if config.DiscordWebhookID != "" && config.DiscordWebhookAccessToken != "" {
		webhookID, err := strconv.ParseInt(config.DiscordWebhookID, 10, 64)
		if err != nil {
			return nil, err
		}
		client := webhook.New(snowflake.ID(webhookID), config.DiscordWebhookAccessToken,
			webhook.WithLogger(logrus.New()),
			webhook.WithDefaultAllowedMentions(discord.AllowedMentions{
				RepliedUser: false,
			}),
		)
		service.discordClient = &client
		if err := service.SendInfo(context.Background(), "initialized keeper alert service"); err != nil {
			return nil, err
		}
	}
	if config.SlackWebhookURL != "" {
		service.slackWebhookURL = pointer.ToString(config.SlackWebhookURL)
	}
	return service, nil
}

type serviceImpl struct {
	discordClient   *webhook.Client
	slackWebhookURL *string
}

func (a serviceImpl) SendError(ctx context.Context, err error) (retErr error) {
	if err == nil {
		return
	}
	if sendErr := a.sendDiscord(ctx, discord.Embed{
		Title:       "Error",
		Description: err.Error(),
		Color:       int(ErrorColor),
	}); sendErr != nil {
		retErr = multierror.Append(retErr, sendErr)
	}
	block := slack.NewContextBlock("summary",
		slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf(":rotating_light:\n%s", err.Error()), false, false),
	)
	if sendErr := a.sendSlack(ctx, block); sendErr != nil {
		retErr = multierror.Append(retErr, sendErr)
	}
	return retErr
}

func (a serviceImpl) SendInfo(ctx context.Context, message string) (retErr error) {
	if sendErr := a.sendDiscord(ctx, discord.Embed{
		Title:       "Info",
		Description: message,
		Color:       int(InfoColor),
	}); sendErr != nil {
		retErr = multierror.Append(retErr, sendErr)
	}
	block := slack.NewContextBlock("summary",
		slack.NewTextBlockObject(slack.MarkdownType, message, false, false),
	)
	if sendErr := a.sendSlack(ctx, block); sendErr != nil {
		retErr = multierror.Append(retErr, sendErr)
	}
	return retErr
}

func (a serviceImpl) sendDiscord(ctx context.Context, embeds ...discord.Embed) (err error) {
	if a.discordClient != nil {
		_, err = (*a.discordClient).CreateMessage(
			discord.NewWebhookMessageCreateBuilder().
				SetAvatarURL("https://pbs.twimg.com/profile_images/1512938686702403603/DDObiFjj_400x400.jpg").
				SetEmbeds(embeds...).
				Build(),
			// delay each request by 2 seconds
			rest.WithDelay(2*time.Second),
			rest.WithCtx(ctx),
		)
	} else {
		logrus.Info("alert service disabled, skipping info alert")
	}
	return err
}

func (a serviceImpl) sendSlack(ctx context.Context, blocks ...slack.Block) (err error) {
	if a.slackWebhookURL != nil {
		return slack.PostWebhookContext(ctx, *a.slackWebhookURL, &slack.WebhookMessage{
			Blocks: &slack.Blocks{BlockSet: blocks},
		})
	} else {
		logrus.Info("alert service disabled, skipping info alert")
	}
	return err
}
