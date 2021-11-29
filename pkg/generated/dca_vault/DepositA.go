// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package dca_vault

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// DepositA is the `depositA` instruction.
type DepositA struct {
	Amount              *uint64
	TotalDurationMillis *uint64

	// [0] = [SIGNER] depositor
	//
	// [1] = [WRITE] vault
	//
	// [2] = [WRITE] depositorTokenAAccount
	//
	// [3] = [WRITE] depositorTokenBAccount
	//
	// [4] = [] tokenAMint
	//
	// [5] = [] tokenBMint
	//
	// [6] = [] tokenProgram
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewDepositAInstructionBuilder creates a new `DepositA` instruction builder.
func NewDepositAInstructionBuilder() *DepositA {
	nd := &DepositA{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 7),
	}
	return nd
}

// SetAmount sets the "amount" parameter.
func (inst *DepositA) SetAmount(amount uint64) *DepositA {
	inst.Amount = &amount
	return inst
}

// SetTotalDurationMillis sets the "totalDurationMillis" parameter.
func (inst *DepositA) SetTotalDurationMillis(totalDurationMillis uint64) *DepositA {
	inst.TotalDurationMillis = &totalDurationMillis
	return inst
}

// SetDepositorAccount sets the "depositor" account.
func (inst *DepositA) SetDepositorAccount(depositor ag_solanago.PublicKey) *DepositA {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(depositor).SIGNER()
	return inst
}

// GetDepositorAccount gets the "depositor" account.
func (inst *DepositA) GetDepositorAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

// SetVaultAccount sets the "vault" account.
func (inst *DepositA) SetVaultAccount(vault ag_solanago.PublicKey) *DepositA {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(vault).WRITE()
	return inst
}

// GetVaultAccount gets the "vault" account.
func (inst *DepositA) GetVaultAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[1]
}

// SetDepositorTokenAAccountAccount sets the "depositorTokenAAccount" account.
func (inst *DepositA) SetDepositorTokenAAccountAccount(depositorTokenAAccount ag_solanago.PublicKey) *DepositA {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(depositorTokenAAccount).WRITE()
	return inst
}

// GetDepositorTokenAAccountAccount gets the "depositorTokenAAccount" account.
func (inst *DepositA) GetDepositorTokenAAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[2]
}

// SetDepositorTokenBAccountAccount sets the "depositorTokenBAccount" account.
func (inst *DepositA) SetDepositorTokenBAccountAccount(depositorTokenBAccount ag_solanago.PublicKey) *DepositA {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(depositorTokenBAccount).WRITE()
	return inst
}

// GetDepositorTokenBAccountAccount gets the "depositorTokenBAccount" account.
func (inst *DepositA) GetDepositorTokenBAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[3]
}

// SetTokenAMintAccount sets the "tokenAMint" account.
func (inst *DepositA) SetTokenAMintAccount(tokenAMint ag_solanago.PublicKey) *DepositA {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(tokenAMint)
	return inst
}

// GetTokenAMintAccount gets the "tokenAMint" account.
func (inst *DepositA) GetTokenAMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[4]
}

// SetTokenBMintAccount sets the "tokenBMint" account.
func (inst *DepositA) SetTokenBMintAccount(tokenBMint ag_solanago.PublicKey) *DepositA {
	inst.AccountMetaSlice[5] = ag_solanago.Meta(tokenBMint)
	return inst
}

// GetTokenBMintAccount gets the "tokenBMint" account.
func (inst *DepositA) GetTokenBMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[5]
}

// SetTokenProgramAccount sets the "tokenProgram" account.
func (inst *DepositA) SetTokenProgramAccount(tokenProgram ag_solanago.PublicKey) *DepositA {
	inst.AccountMetaSlice[6] = ag_solanago.Meta(tokenProgram)
	return inst
}

// GetTokenProgramAccount gets the "tokenProgram" account.
func (inst *DepositA) GetTokenProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[6]
}

func (inst DepositA) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_DepositA,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst DepositA) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *DepositA) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Amount == nil {
			return errors.New("Amount parameter is not set")
		}
		if inst.TotalDurationMillis == nil {
			return errors.New("TotalDurationMillis parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Depositor is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.Vault is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.DepositorTokenAAccount is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.DepositorTokenBAccount is not set")
		}
		if inst.AccountMetaSlice[4] == nil {
			return errors.New("accounts.TokenAMint is not set")
		}
		if inst.AccountMetaSlice[5] == nil {
			return errors.New("accounts.TokenBMint is not set")
		}
		if inst.AccountMetaSlice[6] == nil {
			return errors.New("accounts.TokenProgram is not set")
		}
	}
	return nil
}

func (inst *DepositA) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("DepositA")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=2]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("             Amount", *inst.Amount))
						paramsBranch.Child(ag_format.Param("TotalDurationMillis", *inst.TotalDurationMillis))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=7]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("      depositor", inst.AccountMetaSlice[0]))
						accountsBranch.Child(ag_format.Meta("          vault", inst.AccountMetaSlice[1]))
						accountsBranch.Child(ag_format.Meta("depositorTokenA", inst.AccountMetaSlice[2]))
						accountsBranch.Child(ag_format.Meta("depositorTokenB", inst.AccountMetaSlice[3]))
						accountsBranch.Child(ag_format.Meta("     tokenAMint", inst.AccountMetaSlice[4]))
						accountsBranch.Child(ag_format.Meta("     tokenBMint", inst.AccountMetaSlice[5]))
						accountsBranch.Child(ag_format.Meta("   tokenProgram", inst.AccountMetaSlice[6]))
					})
				})
		})
}

func (obj DepositA) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Amount` param:
	err = encoder.Encode(obj.Amount)
	if err != nil {
		return err
	}
	// Serialize `TotalDurationMillis` param:
	err = encoder.Encode(obj.TotalDurationMillis)
	if err != nil {
		return err
	}
	return nil
}
func (obj *DepositA) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Amount`:
	err = decoder.Decode(&obj.Amount)
	if err != nil {
		return err
	}
	// Deserialize `TotalDurationMillis`:
	err = decoder.Decode(&obj.TotalDurationMillis)
	if err != nil {
		return err
	}
	return nil
}

// NewDepositAInstruction declares a new DepositA instruction with the provided parameters and accounts.
func NewDepositAInstruction(
	// Parameters:
	amount uint64,
	totalDurationMillis uint64,
	// Accounts:
	depositor ag_solanago.PublicKey,
	vault ag_solanago.PublicKey,
	depositorTokenAAccount ag_solanago.PublicKey,
	depositorTokenBAccount ag_solanago.PublicKey,
	tokenAMint ag_solanago.PublicKey,
	tokenBMint ag_solanago.PublicKey,
	tokenProgram ag_solanago.PublicKey) *DepositA {
	return NewDepositAInstructionBuilder().
		SetAmount(amount).
		SetTotalDurationMillis(totalDurationMillis).
		SetDepositorAccount(depositor).
		SetVaultAccount(vault).
		SetDepositorTokenAAccountAccount(depositorTokenAAccount).
		SetDepositorTokenBAccountAccount(depositorTokenBAccount).
		SetTokenAMintAccount(tokenAMint).
		SetTokenBMintAccount(tokenBMint).
		SetTokenProgramAccount(tokenProgram)
}