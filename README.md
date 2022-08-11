# drip-keeper

[![Build and Test](https://github.com/dcaf-labs/drip-keeper/actions/workflows/build-and-test.yaml/badge.svg?branch=main)](https://github.com/dcaf-labs/drip-keeper/actions/workflows/build-and-test.yaml)
[![Deploy](https://github.com/dcaf-labs/drip-keeper/actions/workflows/deploy-devnet.yaml/badge.svg?branch=main)](https://github.com/dcaf-labs/drip-keeper/actions/workflows/deploy-devnet.yaml)
[![Maintainability](https://api.codeclimate.com/v1/badges/5b6787b16c4570e6b052/maintainability)](https://codeclimate.com/repos/61a44f1543298e01a1003151/maintainability)
[![Code Coverage](https://api.codeclimate.com/v1/badges/5b6787b16c4570e6b052/test_coverage)](https://codeclimate.com/repos/61a44f1543298e01a1003151/test_coverage)

## Getting Started

Setup the `.env` file:

1. Create `.env` in the root
2. Set `KEEPER_BOT_WALLET` env var with the contents of the json from `solana-keygen`

Install dependencies (needed for orca whirlpools)
1. install node `v16.13.0`
2. `npm i`

Run the Bot: `go run main.go`


## Devnet

To run the bot against devnet:
`ENV=DEVNET go run main.go`.

## Heroku
Enable the nodejs buildpack (needed because the bot spins off a node subprocess to use the orca SDK).
```bash
heroku buildpacks:add --index 1 heroku/nodejs -a keeper-bot-devnet
# verify with
heroku buildpacks -a keeper-bot-devnet
# The output should look like the following (go should be last)
# === keeper-bot-devnet Buildpack URLs
# 1. heroku/nodejs
# 2. heroku/go
```