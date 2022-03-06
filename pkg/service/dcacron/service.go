package dca

import (
	"context"
	"fmt"
	"time"

	"github.com/Dcaf-Protocol/keeper-bot/pkg/wallet"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type DCACron struct {
	Cron   *cron.Cron
	Wallet *wallet.Wallet
}

func NewDCACron(
	lc fx.Lifecycle,
	wallet *wallet.Wallet,
) (*DCACron, error) {
	dca := DCACron{Wallet: wallet, Cron: cron.New()}
	cronPeriod := 1 * time.Minute
	if _, err := dca.Cron.AddFunc(
		fmt.Sprintf("@every %s", cronPeriod), dca.run); err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			dca.Cron.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return dca.stopCron(ctx)
		},
	})
	return &dca, nil
}

func (dca *DCACron) stopCron(
	ctx context.Context,
) error {
	// Stop cron and wait for it to finish or timeout
	timeout := time.Minute
	ticker := time.NewTicker(timeout)
	cronStop := dca.Cron.Stop().Done()
	ctxDone := ctx.Done()
	select {
	case <-ticker.C:
		return fmt.Errorf("failed to stop dca cron in %s", timeout.String())
	case <-cronStop:
	case <-ctxDone:
		return nil
	}
	return nil
}

func (dca *DCACron) run() {
	logrus.Info("running dca")
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	vault := "TODO: VAULT PUBKEY"
	if err := dca.Wallet.TriggerDCA(ctx, vault); err != nil {
		logrus.
			WithFields(logrus.Fields{"vault": vault}).
			WithError(err).
			Errorf("failed to trigger DCA")
		return
	}
	logrus.
		WithFields(logrus.Fields{"vault": vault}).
		Info("triggered DCA")
}
