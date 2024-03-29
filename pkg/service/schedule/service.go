package schedule

//
//import (
//	"context"
//	"fmt"
//	"runtime/debug"
//	"time"
//
//	"github.com/Dcaf-Protocol/drip-keeper/configs"
//	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/alert"
//	solclient "github.com/Dcaf-Protocol/drip-keeper/pkg/service/clients/solana"
//	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/eventbus"
//	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/keeper"
//	"github.com/asaskevich/EventBus"
//	"github.com/gagliardetto/solana-go"
//	cmap "github.com/orcaman/concurrent-map"
//	"github.com/robfig/cron/v3"
//	"github.com/sirupsen/logrus"
//	"go.uber.org/fx"
//)
//
//type Scheduler struct {
//	At    time.Time
//	Every time.Duration
//}
//
//func (s *Scheduler) Next(t time.Time) time.Time {
//	if t.After(s.At) {
//		return t.Add(s.Every)
//	}
//	return s.At
//}
//
//type DripSchedulerService struct {
//	alertService alert.Service
//	keeper       *keeper.KeeperService
//	solanaClient *solclient.SolanaClient
//
//	dripConfigs cmap.ConcurrentMap
//}
//
//type DripConfig struct {
//	Cron   *cron.Cron
//	Config configs.DripConfig
//}
//
//func NewSchedulerService(
//	lc fx.Lifecycle,
//	config *configs.Config,
//	eventBus EventBus.Bus,
//	alertService alert.Service,
//	keeper *keeper.KeeperService,
//	solanaClient *solclient.SolanaClient,
//) (*DripSchedulerService, error) {
//	logrus.Info("initializing dca cron service")
//	dcaCronService := DripSchedulerService{
//		alertService: alertService,
//		keeper:       keeper,
//		solanaClient: solanaClient,
//		dripConfigs:  cmap.New(),
//	}
//	// Start this before lifecycle to ensure it is subscribed as soon as invoke is called
//	if err := eventBus.Subscribe(string(eventbus.VaultConfigTopic), dcaCronService.registerDripConfig); err != nil {
//		return nil, err
//	}
//	scheduleStatusCron := cron.New()
//	if _, err := scheduleStatusCron.AddFunc("@every 1m", dcaCronService.scheduleStatus); err != nil {
//		return nil, err
//	}
//	scheduleStatusCron.Start()
//	lc.Append(fx.Hook{
//		OnStop: func(ctx context.Context) error {
//			if err := dcaCronService.stopCron(ctx, scheduleStatusCron); err != nil {
//				logrus.WithError(err).Error("failed to stop scheduleStatus cron")
//			}
//			for !dcaCronService.dripConfigs.IsEmpty() {
//				if config.ShouldDiscoverNewConfigs {
//					// Don't return err if this fails
//					// we need to stop the cronJobs
//					if err := eventBus.Unsubscribe(string(eventbus.VaultConfigTopic), dcaCronService.registerDripConfig); err != nil {
//						logrus.WithError(err).WithField("bus", eventbus.VaultConfigTopic).Error("failed to unsubscribe to event bus")
//					}
//				}
//				for _, key := range dcaCronService.dripConfigs.Keys() {
//					v, ok := dcaCronService.dripConfigs.Pop(key)
//					if !ok {
//						continue
//					}
//					dcaCron := v.(*DripConfig)
//					if err := dcaCronService.stopCron(ctx, dcaCron.Cron); err != nil {
//						logrus.WithError(err).WithField("vault", dcaCron.Config.VaultPub).Error("failed to stop dca cron job")
//					}
//				}
//			}
//			return nil
//		},
//	})
//	return nil, nil
//}
//
//func (dripScheduler *DripSchedulerService) scheduleStatus() {
//	logrus.Info("checking schedule status")
//	dripScheduler.dripConfigs.IterCb(func(key string, v interface{}) {
//		dripConfig := v.(*DripConfig)
//		if dripConfig == nil {
//			return
//		}
//		for _, entry := range dripConfig.Cron.Entries() {
//			logrus.
//				WithField("nextTryUnix", entry.Next.Unix()).
//				WithField("nextTry", entry.Next).
//				WithField("vault", dripConfig.Config.VaultPub).
//				Infof("next scheduled drip attempt")
//		}
//
//	})
//}
//
//func (dripScheduler *DripSchedulerService) registerDripConfig(newConfig configs.DripConfig) (*DripConfig, error) {
//	log := logrus.WithField("vault", newConfig.VaultPub)
//	log.Debug("received new config")
//
//	if v, ok := dripScheduler.dripConfigs.Get(newConfig.VaultPub); ok {
//		dripConfig := v.(*DripConfig)
//		// If there is a new whirlpool config, and it's different from what we have, set it
//		// If there is a new splTokenSwap config, and it's different from what we have, set it
//		if (newConfig.OrcaWhirlpoolConfig.WhirlpoolPub != "" && dripConfig.Config.OrcaWhirlpoolConfig.WhirlpoolPub != newConfig.OrcaWhirlpoolConfig.WhirlpoolPub) ||
//			(newConfig.SPLTokenSwapConfig.Swap != "" && dripConfig.Config.SPLTokenSwapConfig.Swap != newConfig.SPLTokenSwapConfig.Swap) {
//			logrus.
//				WithField("vault", newConfig.VaultPub).
//				WithField("oldSwap", dripConfig.Config.SPLTokenSwapConfig.Swap).
//				WithField("newSwap", newConfig.SPLTokenSwapConfig.Swap).
//				WithField("oldSwap", dripConfig.Config.SPLTokenSwapConfig.Swap).
//				WithField("newSwap", newConfig.SPLTokenSwapConfig.Swap).
//				Info("vault already registered, overriding swap")
//			dripConfig.Config = newConfig
//			dripScheduler.dripConfigs.Set(newConfig.VaultPub, dripConfig)
//			return dripConfig, nil
//		}
//		log.Debug("vault already registered, skipping cron creation")
//		return nil, nil
//	}
//	log.Debug("creating cron")
//	return dripScheduler.scheduleDrip(newConfig, true, "registerDripConfig")
//}
//
//func (dripScheduler *DripSchedulerService) stopCron(
//	ctx context.Context, cron *cron.Cron,
//) error {
//	// Stop cron and wait for it to finish or timeout
//	timeout := time.Minute
//	ticker := time.NewTicker(timeout)
//	cronStop := cron.Stop().Done()
//	ctxDone := ctx.Done()
//	select {
//	case <-ticker.C:
//		return fmt.Errorf("failed to stop dca cron in %s", timeout.String())
//	case <-cronStop:
//	case <-ctxDone:
//		return nil
//	}
//	return nil
//}
//
//func (dripScheduler *DripSchedulerService) runWithRetry(vault string, try, maxTry int, timeout int64) {
//	log := logrus.
//		WithField("vault", vault).
//		WithField("try", try).
//		WithField("maxTry", maxTry).
//		WithField("timeout", timeout)
//	v, ok := dripScheduler.dripConfigs.Get(vault)
//	if !ok {
//		log.Error("failed to get dcaCron from dripConfigs")
//		return
//	}
//	dripConfig := v.(*DripConfig)
//	config := dripConfig.Config
//
//	defer func() {
//		if r := recover(); r != nil {
//			_ = dripScheduler.alertService.SendError(context.Background(), fmt.Errorf("panic in runWithRetry"))
//			log.
//				WithField("r", r).
//				WithField("stackTrace", string(debug.Stack())).
//				WithField("config", config).
//				Errorf("panic in doDrip")
//		}
//	}()
//	if err := dripScheduler.keeper.Run(config); err != nil {
//		log := log.WithError(err)
//		if try >= maxTry {
//			if err.Error() != keeper.ErrDripAmount0 && err.Error() != keeper.ErrDripAlreadyTriggered {
//				log.Error("failed to drip")
//				if alertErr := dripScheduler.alertService.SendError(context.Background(), fmt.Errorf("err in runWithRetry, try %d, maxTry %d, vault %s, err %w", try, maxTry, config.VaultPub, err)); alertErr != nil {
//					log.WithField("alertErr", alertErr.Error()).Errorf("failed to send error alert")
//				}
//			}
//			// first stop the current cron to avoid mem leaks
//			if stopErr := dripScheduler.stopCron(context.Background(), dripConfig.Cron); stopErr != nil {
//				log.WithError(stopErr).Error("failed to stop cron job while trying to reschedule")
//			}
//
//			// create new drip handler
//			if err.Error() == keeper.ErrDripAlreadyTriggered {
//				if _, scheduleErr := dripScheduler.scheduleDrip(config, false, err.Error()); scheduleErr != nil {
//					log.WithField("scheduleErr", scheduleErr.Error()).WithField("snapToBeginning", false).Errorf("failed to reschedule drip")
//				}
//			} else {
//				if _, scheduleErr := dripScheduler.scheduleDrip(config, true, err.Error()); scheduleErr != nil {
//					log.WithField("scheduleErr", scheduleErr.Error()).WithField("snapToBeginning", true).Errorf("failed to reschedule drip")
//				}
//			}
//			return
//		}
//		//log.Warn("waiting before retrying drip")
//		time.Sleep(time.Duration(timeout) * time.Second)
//		dripScheduler.runWithRetry(config.VaultPub, try+1, maxTry, timeout)
//	}
//}
//
//func (dripScheduler *DripSchedulerService) scheduleDrip(config configs.DripConfig, snapToBeginning bool, reason string) (*DripConfig, error) {
//	log := logrus.WithField("vault", config.VaultPub).WithField("reason", reason).WithField("snapToBeginning", snapToBeginning)
//	schedule, granularity, err := dripScheduler.getScheduler(config.VaultProtoConfig, config.VaultPub, snapToBeginning)
//	if err != nil {
//		log.WithError(err).Error("failed to getScheduler")
//		return nil, err
//	}
//	log.WithField("nextSchedule", schedule.At.String())
//	cronJob := cron.New()
//	doDrip := dripScheduler.getDripFunc(config.VaultPub)
//	if _, err := cronJob.AddFunc(fmt.Sprintf("@every %ds", granularity), doDrip); err != nil {
//		log.WithError(err).Error("failed to addFunc to cronJob while trying to reschedule")
//		return nil, err
//	}
//	cronJob.Schedule(&schedule, cronJob)
//	newDripConfig := DripConfig{
//		Config: config,
//		Cron:   cronJob,
//	}
//	dripScheduler.dripConfigs.Set(config.VaultPub, &newDripConfig)
//	// Run the first trigger dca right now in case we created this cron past the lastDCAActivation timestamp
//	go doDrip()
//	newDripConfig.Cron.Start()
//	log.WithField("vault", config.VaultPub).Info("scheduled doDrip")
//	return &newDripConfig, nil
//}
//
//func (dripScheduler *DripSchedulerService) getDripFunc(vault string) func() {
//	return func() {
//		dripScheduler.runWithRetry(vault, 0, 3, 30)
//	}
//}
//
//func (dripScheduler *DripSchedulerService) getScheduler(address string, vault string, snapToBeginning bool) (Scheduler, uint64, error) {
//	protoConfigPubkey, err := solana.PublicKeyFromBase58(address)
//	if err != nil {
//		return Scheduler{}, 0, err
//	}
//	vaultProtoConfigData, err := dripScheduler.solanaClient.GetVaultProtoConfig(context.Background(), protoConfigPubkey)
//	if err != nil {
//		return Scheduler{}, 0, err
//	}
//	vaultPubkey, err := solana.PublicKeyFromBase58(vault)
//	if err != nil {
//		return Scheduler{}, 0, err
//	}
//	vaultData, err := dripScheduler.solanaClient.GetVault(context.Background(), vaultPubkey)
//	if err != nil {
//		return Scheduler{}, 0, err
//	}
//	schedule := Scheduler{
//		At:    time.Unix(vaultData.DripActivationTimestamp, 0),
//		Every: time.Second * time.Duration(vaultProtoConfigData.Granularity),
//	}
//	if snapToBeginning {
//		schedule = Scheduler{
//			At: time.
//				Now().
//				Add(-1 * time.Duration(time.Now().Unix()%int64(vaultProtoConfigData.Granularity))),
//			Every: time.Second * time.Duration(vaultProtoConfigData.Granularity),
//		}
//	}
//
//	return schedule, vaultProtoConfigData.Granularity, nil
//}
