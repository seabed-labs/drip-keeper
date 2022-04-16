// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package dca_vault

import (
	"fmt"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
)

type Position struct {
	PositionAuthority        ag_solanago.PublicKey
	DepositedTokenAAmount    uint64
	WithdrawnTokenBAmount    uint64
	Vault                    ag_solanago.PublicKey
	DepositTimestamp         int64
	DcaPeriodIdBeforeDeposit uint64
	NumberOfSwaps            uint64
	PeriodicDripAmount       uint64
	IsClosed                 bool
	Bump                     uint8
}

var PositionDiscriminator = [8]byte{170, 188, 143, 228, 122, 64, 247, 208}

func (obj Position) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Write account discriminator:
	err = encoder.WriteBytes(PositionDiscriminator[:], false)
	if err != nil {
		return err
	}
	// Serialize `PositionAuthority` param:
	err = encoder.Encode(obj.PositionAuthority)
	if err != nil {
		return err
	}
	// Serialize `DepositedTokenAAmount` param:
	err = encoder.Encode(obj.DepositedTokenAAmount)
	if err != nil {
		return err
	}
	// Serialize `WithdrawnTokenBAmount` param:
	err = encoder.Encode(obj.WithdrawnTokenBAmount)
	if err != nil {
		return err
	}
	// Serialize `Vault` param:
	err = encoder.Encode(obj.Vault)
	if err != nil {
		return err
	}
	// Serialize `DepositTimestamp` param:
	err = encoder.Encode(obj.DepositTimestamp)
	if err != nil {
		return err
	}
	// Serialize `DcaPeriodIdBeforeDeposit` param:
	err = encoder.Encode(obj.DcaPeriodIdBeforeDeposit)
	if err != nil {
		return err
	}
	// Serialize `NumberOfSwaps` param:
	err = encoder.Encode(obj.NumberOfSwaps)
	if err != nil {
		return err
	}
	// Serialize `PeriodicDripAmount` param:
	err = encoder.Encode(obj.PeriodicDripAmount)
	if err != nil {
		return err
	}
	// Serialize `IsClosed` param:
	err = encoder.Encode(obj.IsClosed)
	if err != nil {
		return err
	}
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	return nil
}

func (obj *Position) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Read and check account discriminator:
	{
		discriminator, err := decoder.ReadTypeID()
		if err != nil {
			return err
		}
		if !discriminator.Equal(PositionDiscriminator[:]) {
			return fmt.Errorf(
				"wrong discriminator: wanted %s, got %s",
				"[170 188 143 228 122 64 247 208]",
				fmt.Sprint(discriminator[:]))
		}
	}
	// Deserialize `PositionAuthority`:
	err = decoder.Decode(&obj.PositionAuthority)
	if err != nil {
		return err
	}
	// Deserialize `DepositedTokenAAmount`:
	err = decoder.Decode(&obj.DepositedTokenAAmount)
	if err != nil {
		return err
	}
	// Deserialize `WithdrawnTokenBAmount`:
	err = decoder.Decode(&obj.WithdrawnTokenBAmount)
	if err != nil {
		return err
	}
	// Deserialize `Vault`:
	err = decoder.Decode(&obj.Vault)
	if err != nil {
		return err
	}
	// Deserialize `DepositTimestamp`:
	err = decoder.Decode(&obj.DepositTimestamp)
	if err != nil {
		return err
	}
	// Deserialize `DcaPeriodIdBeforeDeposit`:
	err = decoder.Decode(&obj.DcaPeriodIdBeforeDeposit)
	if err != nil {
		return err
	}
	// Deserialize `NumberOfSwaps`:
	err = decoder.Decode(&obj.NumberOfSwaps)
	if err != nil {
		return err
	}
	// Deserialize `PeriodicDripAmount`:
	err = decoder.Decode(&obj.PeriodicDripAmount)
	if err != nil {
		return err
	}
	// Deserialize `IsClosed`:
	err = decoder.Decode(&obj.IsClosed)
	if err != nil {
		return err
	}
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	return nil
}

type VaultPeriod struct {
	Vault    ag_solanago.PublicKey
	PeriodId uint64
	Twap     ag_binary.Uint128
	Dar      uint64
	Bump     uint8
}

var VaultPeriodDiscriminator = [8]byte{224, 196, 159, 18, 79, 227, 22, 122}

func (obj VaultPeriod) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Write account discriminator:
	err = encoder.WriteBytes(VaultPeriodDiscriminator[:], false)
	if err != nil {
		return err
	}
	// Serialize `Vault` param:
	err = encoder.Encode(obj.Vault)
	if err != nil {
		return err
	}
	// Serialize `PeriodId` param:
	err = encoder.Encode(obj.PeriodId)
	if err != nil {
		return err
	}
	// Serialize `Twap` param:
	err = encoder.Encode(obj.Twap)
	if err != nil {
		return err
	}
	// Serialize `Dar` param:
	err = encoder.Encode(obj.Dar)
	if err != nil {
		return err
	}
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	return nil
}

func (obj *VaultPeriod) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Read and check account discriminator:
	{
		discriminator, err := decoder.ReadTypeID()
		if err != nil {
			return err
		}
		if !discriminator.Equal(VaultPeriodDiscriminator[:]) {
			return fmt.Errorf(
				"wrong discriminator: wanted %s, got %s",
				"[224 196 159 18 79 227 22 122]",
				fmt.Sprint(discriminator[:]))
		}
	}
	// Deserialize `Vault`:
	err = decoder.Decode(&obj.Vault)
	if err != nil {
		return err
	}
	// Deserialize `PeriodId`:
	err = decoder.Decode(&obj.PeriodId)
	if err != nil {
		return err
	}
	// Deserialize `Twap`:
	err = decoder.Decode(&obj.Twap)
	if err != nil {
		return err
	}
	// Deserialize `Dar`:
	err = decoder.Decode(&obj.Dar)
	if err != nil {
		return err
	}
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	return nil
}

type VaultProtoConfig struct {
	Granularity          uint64
	TriggerDcaSpread     uint16
	BaseWithdrawalSpread uint16
}

var VaultProtoConfigDiscriminator = [8]byte{173, 22, 36, 165, 190, 3, 142, 199}

func (obj VaultProtoConfig) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Write account discriminator:
	err = encoder.WriteBytes(VaultProtoConfigDiscriminator[:], false)
	if err != nil {
		return err
	}
	// Serialize `Granularity` param:
	err = encoder.Encode(obj.Granularity)
	if err != nil {
		return err
	}
	// Serialize `TriggerDcaSpread` param:
	err = encoder.Encode(obj.TriggerDcaSpread)
	if err != nil {
		return err
	}
	// Serialize `BaseWithdrawalSpread` param:
	err = encoder.Encode(obj.BaseWithdrawalSpread)
	if err != nil {
		return err
	}
	return nil
}

func (obj *VaultProtoConfig) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Read and check account discriminator:
	{
		discriminator, err := decoder.ReadTypeID()
		if err != nil {
			return err
		}
		if !discriminator.Equal(VaultProtoConfigDiscriminator[:]) {
			return fmt.Errorf(
				"wrong discriminator: wanted %s, got %s",
				"[173 22 36 165 190 3 142 199]",
				fmt.Sprint(discriminator[:]))
		}
	}
	// Deserialize `Granularity`:
	err = decoder.Decode(&obj.Granularity)
	if err != nil {
		return err
	}
	// Deserialize `TriggerDcaSpread`:
	err = decoder.Decode(&obj.TriggerDcaSpread)
	if err != nil {
		return err
	}
	// Deserialize `BaseWithdrawalSpread`:
	err = decoder.Decode(&obj.BaseWithdrawalSpread)
	if err != nil {
		return err
	}
	return nil
}

type Vault struct {
	ProtoConfig            ag_solanago.PublicKey
	TokenAMint             ag_solanago.PublicKey
	TokenBMint             ag_solanago.PublicKey
	TokenAAccount          ag_solanago.PublicKey
	TokenBAccount          ag_solanago.PublicKey
	TreasuryTokenBAccount  ag_solanago.PublicKey
	LastDcaPeriod          uint64
	DripAmount             uint64
	DcaActivationTimestamp int64
	Bump                   uint8
}

var VaultDiscriminator = [8]byte{211, 8, 232, 43, 2, 152, 117, 119}

func (obj Vault) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Write account discriminator:
	err = encoder.WriteBytes(VaultDiscriminator[:], false)
	if err != nil {
		return err
	}
	// Serialize `ProtoConfig` param:
	err = encoder.Encode(obj.ProtoConfig)
	if err != nil {
		return err
	}
	// Serialize `TokenAMint` param:
	err = encoder.Encode(obj.TokenAMint)
	if err != nil {
		return err
	}
	// Serialize `TokenBMint` param:
	err = encoder.Encode(obj.TokenBMint)
	if err != nil {
		return err
	}
	// Serialize `TokenAAccount` param:
	err = encoder.Encode(obj.TokenAAccount)
	if err != nil {
		return err
	}
	// Serialize `TokenBAccount` param:
	err = encoder.Encode(obj.TokenBAccount)
	if err != nil {
		return err
	}
	// Serialize `TreasuryTokenBAccount` param:
	err = encoder.Encode(obj.TreasuryTokenBAccount)
	if err != nil {
		return err
	}
	// Serialize `LastDcaPeriod` param:
	err = encoder.Encode(obj.LastDcaPeriod)
	if err != nil {
		return err
	}
	// Serialize `DripAmount` param:
	err = encoder.Encode(obj.DripAmount)
	if err != nil {
		return err
	}
	// Serialize `DcaActivationTimestamp` param:
	err = encoder.Encode(obj.DcaActivationTimestamp)
	if err != nil {
		return err
	}
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	return nil
}

func (obj *Vault) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Read and check account discriminator:
	{
		discriminator, err := decoder.ReadTypeID()
		if err != nil {
			return err
		}
		if !discriminator.Equal(VaultDiscriminator[:]) {
			return fmt.Errorf(
				"wrong discriminator: wanted %s, got %s",
				"[211 8 232 43 2 152 117 119]",
				fmt.Sprint(discriminator[:]))
		}
	}
	// Deserialize `ProtoConfig`:
	err = decoder.Decode(&obj.ProtoConfig)
	if err != nil {
		return err
	}
	// Deserialize `TokenAMint`:
	err = decoder.Decode(&obj.TokenAMint)
	if err != nil {
		return err
	}
	// Deserialize `TokenBMint`:
	err = decoder.Decode(&obj.TokenBMint)
	if err != nil {
		return err
	}
	// Deserialize `TokenAAccount`:
	err = decoder.Decode(&obj.TokenAAccount)
	if err != nil {
		return err
	}
	// Deserialize `TokenBAccount`:
	err = decoder.Decode(&obj.TokenBAccount)
	if err != nil {
		return err
	}
	// Deserialize `TreasuryTokenBAccount`:
	err = decoder.Decode(&obj.TreasuryTokenBAccount)
	if err != nil {
		return err
	}
	// Deserialize `LastDcaPeriod`:
	err = decoder.Decode(&obj.LastDcaPeriod)
	if err != nil {
		return err
	}
	// Deserialize `DripAmount`:
	err = decoder.Decode(&obj.DripAmount)
	if err != nil {
		return err
	}
	// Deserialize `DcaActivationTimestamp`:
	err = decoder.Decode(&obj.DcaActivationTimestamp)
	if err != nil {
		return err
	}
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	return nil
}
