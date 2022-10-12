package solana

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gagliardetto/solana-go/rpc/jsonrpc"

	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/clients"

	"github.com/dcaf-labs/solana-go-clients/pkg/whirlpool"
	"github.com/gagliardetto/solana-go/programs/token"

	bin "github.com/gagliardetto/binary"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/dcaf-labs/solana-go-clients/pkg/drip"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
	"github.com/sirupsen/logrus"
)

type SolanaClient struct {
	wallet          *solana.Wallet
	feeWalletPubkey solana.PublicKey
	client          *rpc.Client
}

const ErrNotFound = "not found"

func New(
	config *configs.Config,
) (*SolanaClient, error) {
	url, callsPerSecond := GetURLWithRateLimit(config.Network)
	opts := &jsonrpc.RPCClientOpts{
		HTTPClient: clients.GetRateLimitedHTTPClient(callsPerSecond),
	}
	rpcClient := rpc.NewWithCustomRPCClient(jsonrpc.NewClientWithOpts(url, opts))
	resp, err := rpcClient.GetVersion(context.Background())
	if err != nil {
		logrus.WithError(err).Fatalf("failed to get client version info")
		return nil, err
	}
	logrus.
		WithFields(logrus.Fields{
			"version": resp.SolanaCore,
			"url":     url}).
		Info("created rpcClient")

	solanaClient := SolanaClient{client: rpcClient}
	var accountBytes []byte
	if err := json.Unmarshal([]byte(config.Wallet), &accountBytes); err != nil {
		return nil, err
	}
	priv := base58.Encode(accountBytes)
	solWallet, err := solana.WalletFromPrivateKeyBase58(priv)
	if err != nil {
		return nil, err
	}
	solanaClient.wallet = solWallet
	if config.FeeWallet != "" {
		solanaClient.feeWalletPubkey = solana.MustPublicKeyFromBase58(config.FeeWallet)
	} else {
		solanaClient.feeWalletPubkey = solanaClient.wallet.PublicKey()
	}
	logrus.
		WithField("Wallet", solanaClient.wallet.PublicKey()).
		WithField("FeeWalletPubkey", solanaClient.feeWalletPubkey.String()).
		Infof("loaded wallets")
	return &solanaClient, nil
}

func (w *SolanaClient) GetWallet() solana.PublicKey {
	return w.wallet.PublicKey()
}
func (w *SolanaClient) GetFeeWallet() solana.PublicKey {
	return w.feeWalletPubkey
}

func (w *SolanaClient) Send(
	ctx context.Context, instructions ...solana.Instruction,
) error {
	if len(instructions) == 0 {
		return nil
	}
	recent, err := w.client.GetRecentBlockhash(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		return err
	}
	logFields := logrus.Fields{"numInstructions": len(instructions), "block": recent.Value.Blockhash}

	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(w.wallet.PublicKey()),
	)
	if err != nil {
		return fmt.Errorf("failed to create transaction, err %s", err)
	}
	logrus.WithFields(logFields).Infof("built transaction")

	if _, err := tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if w.wallet.PublicKey().Equals(key) {
				return &w.wallet.PrivateKey
			}
			return nil
		},
	); err != nil {
		return fmt.Errorf("failed to sign transaction, err %s", err)
	}
	logrus.WithFields(logFields).Info("signed transaction")

	txHash, err := w.client.SendTransactionWithOpts(
		ctx, tx, rpc.TransactionOpts{
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentConfirmed,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send transaction, err %s", err)
	}
	logFields["txHash"] = txHash

	logrus.WithFields(logFields).Info("waiting for transaction to confirm")
	errC := make(chan error)
	go checkTxHash(w.client, txHash, errC)
	if err := <-errC; err != nil {
		return err
	}
	return nil
}

func (w *SolanaClient) GetVault(ctx context.Context, vaultAddress solana.PublicKey) (drip.Vault, error) {
	var vaultData drip.Vault
	if err := w.getAccount(ctx, vaultAddress, &vaultData); err != nil {
		return drip.Vault{}, err
	}
	return vaultData, nil
}

func (w *SolanaClient) GetVaultProtoConfig(ctx context.Context, vaultProtoConfigAddress solana.PublicKey) (drip.VaultProtoConfig, error) {
	var vaultProtoConfigData drip.VaultProtoConfig
	if err := w.getAccount(ctx, vaultProtoConfigAddress, &vaultProtoConfigData); err != nil {
		return drip.VaultProtoConfig{}, err
	}
	return vaultProtoConfigData, nil
}

func (w *SolanaClient) GetTokenAccount(ctx context.Context, tokenAccountAddress solana.PublicKey) (token.Account, error) {
	var tokenAccountData token.Account
	if err := w.getAccount(ctx, tokenAccountAddress, &tokenAccountData); err != nil {
		return token.Account{}, err
	}
	return tokenAccountData, nil
}

func (w *SolanaClient) GetOrcaWhirlpool(ctx context.Context, whirlpoolAddress solana.PublicKey) (whirlpool.Whirlpool, error) {
	var orcaWhirlpool whirlpool.Whirlpool
	if err := w.getAccount(ctx, whirlpoolAddress, &orcaWhirlpool); err != nil {
		return whirlpool.Whirlpool{}, err
	}
	return orcaWhirlpool, nil
}

func (w *SolanaClient) GetOrcaWhirlpoolTickArray(ctx context.Context, whirlpoolTickArrayAddress solana.PublicKey) (whirlpool.TickArray, error) {
	var orcaWhirlpoolTickArray whirlpool.TickArray
	if err := w.getAccount(ctx, whirlpoolTickArrayAddress, &orcaWhirlpoolTickArray); err != nil {
		return whirlpool.TickArray{}, err
	}
	return orcaWhirlpoolTickArray, nil
}

func (w *SolanaClient) GetMaybeUninitializedTokenAccount(
	ctx context.Context, owner solana.PublicKey, mint solana.PublicKey,
) (solana.PublicKey, solana.Instruction, error) {
	tokenAccount, _, err := solana.FindAssociatedTokenAddress(
		owner,
		mint,
	)
	if err != nil {
		logrus.
			WithError(err).
			WithField("feeWallet", owner.String()).
			WithField("mint", mint.String()).
			Errorf("failed to get ata")
		return solana.PublicKey{}, nil, err
	}
	var instruction solana.Instruction

	// Use GetAccountInfoWithOpts so we can pass in a commitment level
	if _, err := w.client.GetAccountInfoWithOpts(ctx, tokenAccount, &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: "confirmed",
		DataSlice:  nil,
	}); err != nil && err.Error() == ErrNotFound {
		instruction, err = w.CreateTokenAccount(ctx, owner, mint)
		if err != nil {
			return solana.PublicKey{}, nil, err
		}
	}
	return tokenAccount, instruction, nil
}

func (w *SolanaClient) GetMaybeUninitializedVaultPeriod(
	ctx context.Context,
	vault, vaultProtoConfig, tokenAMint, tokenBMint solana.PublicKey,
	vaultPeriodID int64,
) (solana.PublicKey, solana.Instruction, error) {
	log := logrus.WithField("vault", vault.String())
	vaultPeriod, _, err := solana.FindProgramAddress([][]byte{
		[]byte("vault_period"),
		vault[:],
		[]byte(strconv.FormatInt(vaultPeriodID, 10)),
	}, drip.ProgramID)
	if err != nil {
		log.
			WithError(err).
			WithField("programId", drip.ProgramID.String()).
			WithField("vaultPeriodID", vaultPeriodID).
			Errorf("failed to get vaultPeriodI PDA")
		return solana.PublicKey{}, nil, err
	}
	var instruction solana.Instruction
	// Use GetAccountInfoWithOpts so we can pass in a commitment level
	if _, err := w.client.GetAccountInfoWithOpts(ctx, vaultPeriod, &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: "confirmed",
		DataSlice:  nil,
	}); err != nil && err.Error() == ErrNotFound {
		// Failure is likely because the vault period is not initialized
		instruction, err = w.InitVaultPeriod(ctx, vault.String(), vaultProtoConfig.String(), vaultPeriod.String(), tokenAMint.String(), tokenBMint.String(), vaultPeriodID)
		if err != nil {
			log.
				WithError(err).
				WithField("vaultPeriodID", vaultPeriodID).
				Errorf("failed to create InitVaultPeriod instruction")
			return solana.PublicKey{}, nil, err
		}
	} else if err != nil {
		log.
			WithError(err).
			Infof("failed to GetMaybeUninitializedVaultPeriod")
		return solana.PublicKey{}, nil, err
	}
	return vaultPeriod, instruction, nil
}

func (w *SolanaClient) getAccount(ctx context.Context, address solana.PublicKey, v interface{}) error {
	resp, err := w.client.GetAccountInfoWithOpts(
		ctx,
		address,
		&rpc.GetAccountInfoOpts{
			Encoding:   solana.EncodingBase64,
			Commitment: "confirmed",
			DataSlice:  nil,
		})
	if err != nil {
		logrus.
			WithError(err).
			WithField("address", address).
			Errorf("couldn't get acount info")
		return err
	}
	if err := bin.NewBinDecoder(resp.Value.Data.GetBinary()).Decode(v); err != nil {
		logrus.
			WithError(err).
			WithField("address", address).
			Errorf("failed to decode")
		return err
	}
	return nil
}

func checkTxHash(
	client *rpc.Client, txHash solana.Signature, done chan error,
) {
	timeout := time.Second * 60
	ticker := time.NewTicker(timeout)
	for {
		select {
		case <-ticker.C:
			done <- fmt.Errorf("timeout waiting for tx to confirm")
			return
		default:
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		res, err := client.GetSignatureStatuses(
			ctx, false, txHash)
		cancel()
		if err == nil && res != nil && len(res.Value) == 1 && res.Value[0] != nil && res.Value[0].ConfirmationStatus == rpc.ConfirmationStatusConfirmed {
			done <- nil
			return
		}
		if err != nil {
			logrus.WithError(err).Warnf("error getting signature status, retrying")
		}
	}
}

func GetURLWithRateLimit(network configs.Network) (string, int) {
	switch network {
	case configs.MainnetNetwork:
		// return rpc.MainNetBeta_RPC, 3
		// mocha+1@dcaf.so
		return "https://palpable-warmhearted-hexagon.solana-mainnet.discover.quiknode.pro/5793cf44e6e16325347e62d571454890f16e0388", 1
	case configs.DevnetNetwork:
		// return rpc.DevNet_RPC, 3
		// mocha+2@dcaf.so
		return "https://wiser-icy-bush.solana-devnet.discover.quiknode.pro/7288cc56d980336f6fc0508eb1aa73e44fd2efcd", 1
	case configs.NilNetwork:
		fallthrough
	case configs.LocalNetwork:
		fallthrough
	default:
		return rpc.LocalNet_RPC, 2
	}
}
