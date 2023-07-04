package alert

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/test-go/testify/assert"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
)

func TestAlertService(t *testing.T) {
	// TODO: Make these actual unit tests
	//if testing.Short() {
	//	return
	//}

	configs.LoadEnv()

	cfg := &configs.Config{
		DiscordWebhookID:          os.Getenv("DISCORD_WEBHOOK_ID"),
		DiscordWebhookAccessToken: os.Getenv("DISCORD_ACCESS_TOKEN"),
		SlackWebhookURL:           os.Getenv("SLACK_WEBHOOK_URL"),
	}
	t.Run("should send info alert to discord and slack", func(t *testing.T) {
		svc, err := NewService(cfg)
		assert.NoError(t, err)
		assert.NoError(t, svc.SendInfo(context.Background(), "test keeper alert info"))
	})

	t.Run("should send error alert to discord and slack", func(t *testing.T) {
		svc, err := NewService(cfg)
		assert.NoError(t, err)
		assert.NoError(t, svc.SendError(context.Background(), fmt.Errorf("test keeper alert err")))
	})

	t.Run("should not send discord alerts", func(t *testing.T) {
		cfgWithoutDiscord := &configs.Config{
			SlackWebhookURL: cfg.SlackWebhookURL,
		}
		svc, err := NewService(cfgWithoutDiscord)
		assert.NoError(t, err)
		assert.NoError(t, svc.SendInfo(context.Background(), "test keeper alert info"))
		assert.NoError(t, svc.SendError(context.Background(), fmt.Errorf("test keeper alert err")))
	})

	t.Run("should not send slack alerts", func(t *testing.T) {
		cfgWithoutDiscord := &configs.Config{
			DiscordWebhookID:          cfg.DiscordWebhookID,
			DiscordWebhookAccessToken: cfg.DiscordWebhookAccessToken,
		}
		svc, err := NewService(cfgWithoutDiscord)
		assert.NoError(t, err)
		assert.NoError(t, svc.SendInfo(context.Background(), "test keeper alert info"))
		assert.NoError(t, svc.SendError(context.Background(), fmt.Errorf("test keeper alert err")))
	})
}
