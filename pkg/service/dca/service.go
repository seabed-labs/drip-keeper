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

type DCA struct {
	Cron   *cron.Cron
	Wallet *wallet.Wallet
}

func NewDCA(
	lc fx.Lifecycle,
	wallet *wallet.Wallet,
) (*DCA, error) {
	dca := DCA{Wallet: wallet, Cron: cron.New()}
	cronPeriod := 1 * time.Minute
	// TODO(Mocha): Need to change if we use multiple instances
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

func (dca *DCA) stopCron(
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

func (dca *DCA) run() {
	logrus.Info("running dca")
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	vault := "TODO: VAULT PUBKEY"
	if err := dca.Wallet.TriggerDCA(ctx, vault); err != nil {
		logrus.
			WithFields(logrus.Fields{"vault": vault}).
			WithError(err).
			Errorf("failed to trigger DCA")
	}
	logrus.Infof("done DCA")
}
