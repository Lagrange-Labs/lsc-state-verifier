package stateproof_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/Lagrange-Labs/lagrange-node/crypto"
	"github.com/Lagrange-Labs/lsc-state-verifier/stateproof"
)

func TestStateProofVerifyAggregatedSignature(t *testing.T) {
	tests := []struct {
		testFile string
	}{
		{testFile: "testdata/state_proof_32840.json"},
		{testFile: "testdata/state_proof_52648.json"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.testFile, func(t *testing.T) {
			bs, err := os.ReadFile(tt.testFile)
			if err != nil {
				t.Fatalf("failed to read test file: %v", err)
			}

			var stateProof stateproof.StateProof
			if err := json.Unmarshal(bs, &stateProof); err != nil {
				t.Fatalf("failed to unmarshal state proof: %v", err)
			}

			ok, err := stateProof.VerifyAggregatedSignature(crypto.BN254)
			if err != nil {
				t.Fatalf("failed to verify aggregated signature: %v", err)
			}

			if !ok {
				t.Fatalf("aggregated signature verification failed")
			}
		})
	}
}
