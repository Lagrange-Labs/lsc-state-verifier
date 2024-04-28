# Lagrange State Verifier

## Overview

This repository contains a simple state verifier for the Lagrange State Committee. It includes 3 main steps.

1. Verifying the committee root: checks if the committee root is calculated correctly.
2. Verifying the voting power: checks if the voting power of the aggregated signature is enough to pass the threshold (2/3).
3. Verifying the aggregated signature: checks if the aggregated signature is valid. Currently, LSC only supports `BN254` curve.

## Demo Usage

1. `go mod download`
2. Add your `API_KEY` in `.env` file.
3. Choose `CHAIN_ID` and `BATCH_NUMBER` in `.env` file.
   - `CHAIN_ID` is the chain id of the chain you want to verify the state of. `Optimism` is `10` and `Base` is `8452`.
   - `BATCH_NUMBER` is the batch number of the chain you want to verify the state of. 1-indexed.
4. Run `go run main.go` to verify the state of the chain.
