package heartbeat

import (
	"context"
	"net/http"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func NewHeartbeatWorker(
	lc fx.Lifecycle,
	config *configs.Config,
) error {
	if config.HeartbeatURL == "" {
		logrus.Info("heartbeat url is empty")
		return nil
	}
	log := logrus.WithField("heartbeatURL", config.HeartbeatURL)
	log.Debug("initializing heartbeat worker")
	cronJob := cron.New()
	cronFunc := func() {
		log.
			Debug("logging heartbeat")
		resp, err := http.Get(config.HeartbeatURL)
		if err != nil {
			log.WithError(err).Error("failed to ping heartbeat")
		} else {
			log.WithField("status", resp.Status).Info("pinged heartbeat")
		}
	}
	if _, err := cronJob.AddFunc("@every 1m", cronFunc); err != nil {
		return err
	}
	// Start this before lifecycle to ensure it is subscribed as soon as invoke is called
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			cronJob.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			cronJob.Stop()
			return nil
		},
	})
	return nil
}
