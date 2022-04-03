# keeper-bot

[![Build and Test](https://github.com/Dcaf-Protocol/keeper-bot/actions/workflows/build-and-test.yaml/badge.svg)](https://github.com/Dcaf-Protocol/keeper-bot/actions/workflows/build-and-test.yaml)
[![Maintainability](https://api.codeclimate.com/v1/badges/5b6787b16c4570e6b052/maintainability)](https://codeclimate.com/repos/61a44f1543298e01a1003151/maintainability)
[![Code Coverage](https://api.codeclimate.com/v1/badges/5b6787b16c4570e6b052/test_coverage)](https://codeclimate.com/repos/61a44f1543298e01a1003151/test_coverage)

## Getting Started

Setup the `.env` file:

1. Create `.env` in the root
2. Set `KEEPER_BOT_WALLET` env var with the contents of the json from `solana-keygen`

Run the Bot: `go run main.go`

Generate IDL (assumes solana-program is a sibling of keeper-bot): `anchor-go --src=../solana-programs/target/idl/dca_vault.json`

## Devnet

To run the bot in devnet run:
`ENV=DEVNET go run main.go`.

This will use the `devnet.yaml` config (the output of the solana-programs `yarn setup:dev`).
