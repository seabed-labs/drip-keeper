# keeper-bot
[![Test](https://github.com/Dcaf-Protocol/keeper-bot/actions/workflows/test.yaml/badge.svg)](https://github.com/Dcaf-Protocol/keeper-bot/actions/workflows/test.yaml)

## Getting Started
Setup the `.env` file:

1. Create `.env` in the root
2. Set `KEEPER_BOT_ACCOUNT` env var with the contents of the json from `solana-keygen`

Run the Bot: `go run cmd/main.go`
Generate IDL (assumes solana-program is a sibling of keeper-bot): `anchor-go --src=../../solana-programs/target/idl/dca_vault.json`