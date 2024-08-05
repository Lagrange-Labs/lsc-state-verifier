# Lagrange State Verifier

## Overview

This repository contains a simple state verifier for the Lagrange State Committee. It includes 3 main steps.

1. Verifying the committee root: checks if the committee root is calculated correctly.
2. Verifying the voting power: checks if the voting power of the aggregated signature is enough to pass the threshold (2/3).
3. Verifying the aggregated signature: checks if the aggregated signature is valid. Currently, LSC only supports `BN254` curve.

## Demo Usage

1. `go mod download`
2. Update config.toml with relevant information.
   - `chain_id` is the chain id of the chain you want to verify the state of. Please refer to chain id in our [docs](https://docs.lagrange.dev/state-committees/operator-guide/supported-chains).
   - `from_batch_number` is the batch number of the chain from where you want to start verifying the state proof.
3. Run `go run main.go` to verify the state of the chain.
