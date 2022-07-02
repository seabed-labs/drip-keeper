# drip-keeper

[![Build and Test](https://github.com/dcaf-labs/drip-keeper/actions/workflows/build-and-test.yaml/badge.svg?branch=main)](https://github.com/dcaf-labs/drip-keeper/actions/workflows/build-and-test.yaml)
[![Deploy](https://github.com/dcaf-labs/drip-keeper/actions/workflows/deploy-devnet.yaml/badge.svg?branch=main)](https://github.com/dcaf-labs/drip-keeper/actions/workflows/deploy-devnet.yaml)
[![Maintainability](https://api.codeclimate.com/v1/badges/5b6787b16c4570e6b052/maintainability)](https://codeclimate.com/repos/61a44f1543298e01a1003151/maintainability)
[![Code Coverage](https://api.codeclimate.com/v1/badges/5b6787b16c4570e6b052/test_coverage)](https://codeclimate.com/repos/61a44f1543298e01a1003151/test_coverage)

## Getting Started

Setup the `.env` file:

1. Create `.env` in the root
2. Set `KEEPER_BOT_WALLET` env var with the contents of the json from `solana-keygen`

Run the Bot: `go run main.go`

Generate IDL (assumes solana-program is a sibling of drip-kepper): `anchor-go --src=../drip-program/target/idl/drip.json`

Generate Drip Client (assumes drip-backend is a sibling of drip-keeper): `openapi-generator generate -i ../drip-backend/docs/swagger.yaml -g go -o pkg/client/drip --additional-properties=generateInterfaces=true --additional-properties=isGoSubmodule=true --additional-properties=packageName=drip`

TODO(Mocha): Figure out how to generate drip-client as a pkg or instead create a new repo for this

## Devnet

To run the bot in devnet run:
`ENV=DEVNET go run main.go`.

This will use the `devnet.yaml` config (the output of the solana-programs `yarn setup:dev`).
