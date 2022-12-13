# drip-keeper

[![Build and Test](https://github.com/dcaf-labs/drip-keeper/actions/workflows/build-and-test.yaml/badge.svg?branch=main)](https://github.com/dcaf-labs/drip-keeper/actions/workflows/build-and-test.yaml)

## Getting Started

Setup the `.env` file:

1. Create `.env` in the root
2. Set `KEEPER_BOT_WALLET` env var with the contents of the json from `solana-keygen`

Install dependencies (needed for orca whirlpools)
1. install node `v16.13.0`
2. `npm i`

Run the Bot: `go run main.go`