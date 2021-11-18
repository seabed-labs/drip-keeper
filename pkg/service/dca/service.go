package dca

import (
	"context"
	"fmt"
	"time"

	"github.com/Dcaf-Protocol/keeper-bot/configs"
	"github.com/Dcaf-Protocol/keeper-bot/pkg/wallet"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type DCA struct {
	cron   *cron.Cron
	wallet *wallet.Wallet
}

func NewDCA(
	lc fx.Lifecycle,
	secrets *configs.Secrets,
	wallet *wallet.Wallet,
) (*DCA, error) {
	dca := DCA{wallet: wallet, cron: cron.New()}
	cronPeriod := 1 * time.Minute
	// TODO(Mocha): Need to change if we use multiple instances
	if _, err := dca.cron.AddFunc(
		fmt.Sprintf("@every %s", cronPeriod), dca.run); err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			dca.cron.Start()
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
	cronStop := dca.cron.Stop().Done()
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
}
