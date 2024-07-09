package utils

import (
	"fmt"

	"github.com/Lagrange-Labs/lagrange-node/crypto"
	"github.com/Lagrange-Labs/lagrange-node/logger"
	"github.com/Lagrange-Labs/state-verifier/stateproof"
)

// verifyStateProof verifies the state proof.
func verifyStateProof(proofData *stateproof.StateProof) error {
	// Verify the committee root
	if !proofData.VerifyCommitteeRoot() {
		return fmt.Errorf("committee root verification failed for batch %d for chain %d", proofData.BatchHeader.BatchNumber, proofData.BatchHeader.ChainId)
	}
	// Verify the voting power
	if ok, err := proofData.VerifyVotingPower(); !ok {
		return fmt.Errorf("voting power verification failed for batch %d for chain %d: %s", proofData.BatchHeader.BatchNumber, proofData.BatchHeader.ChainId, err)
	}
	// Verify the aggregated signature
	if ok, err := proofData.VerifyAggregatedSignature(crypto.BN254); !ok {
		return fmt.Errorf("aggregated signature verification failed for batch %d for chain %d: %s", proofData.BatchHeader.BatchNumber, proofData.BatchHeader.ChainId, err)
	}

	logger.Infof("State proof verification successful for batch %d for chain %d", proofData.BatchHeader.BatchNumber, proofData.BatchHeader.ChainId)
	return nil
}
