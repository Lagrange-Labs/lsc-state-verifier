package stateproof

import (
	"bytes"
	"fmt"

	"github.com/Lagrange-Labs/lagrange-node/crypto"
	batchtypes "github.com/Lagrange-Labs/lagrange-node/sequencer/types/v2"
	"github.com/Lagrange-Labs/lagrange-node/utils"
)

// StateProof is the struct that represents the state proof.
type StateProof struct {
	OperatorAddresses   []string                    `json:"addresses"`
	BLSPublicKeys       []string                    `json:"public_keys"`
	VotingPowers        []uint64                    `json:"voting_powers"`
	AggregatedSignature string                      `json:"agg_signature"`
	CommitteeHeader     *batchtypes.CommitteeHeader `json:"committee_header"`
	BatchHeader         *batchtypes.BatchHeader     `json:"batch_header"`
	AggregationBits     []uint8                     `json:"aggregation_bits"`
}

// GetBatchHash returns the batch hash of the state proof.
func (sp *StateProof) GetBatchHash() []byte {
	h := append([]byte{}, sp.BatchHeader.Hash()...)
	h = append(h, utils.Hex2Bytes(sp.CommitteeHeader.CurrentCommittee)...)
	h = append(h, utils.Hex2Bytes(sp.CommitteeHeader.NextCommittee)...)
	h = append(h, utils.Uint64ToBytes(sp.CommitteeHeader.TotalVotingPower)...)
	return utils.Hash(h)
}

// VerifyCommitteeRoot verifies the committee root of the state proof.
func (sp *StateProof) VerifyCommitteeRoot() bool {
	leaves := make([][]byte, len(sp.OperatorAddresses))
	for i, operator := range sp.OperatorAddresses {
		leaves[i] = GetLeafHash(utils.Hex2Bytes(operator), utils.Hex2Bytes(sp.BLSPublicKeys[i]), sp.VotingPowers[i])
	}
	rootHash := GetRootHash(leaves)
	return bytes.Equal(rootHash, utils.Hex2Bytes(sp.CommitteeHeader.CurrentCommittee))
}

// VerifyVotingPower verifies if signature has enough voting power.
func (sp *StateProof) VerifyVotingPower() (bool, error) {
	totalVotingPower := uint64(0)
	signedVotingPower := uint64(0)
	for i, votingPower := range sp.VotingPowers {
		totalVotingPower += votingPower
		if sp.AggregationBits[i] == 1 {
			signedVotingPower += votingPower
		}
	}
	if totalVotingPower != sp.CommitteeHeader.TotalVotingPower {
		return false, fmt.Errorf("total voting power %d mismatch with the committee header %d", totalVotingPower, sp.CommitteeHeader.TotalVotingPower)
	}
	if signedVotingPower*3 <= totalVotingPower*2 {
		return false, fmt.Errorf("signed voting power %d is less than 2/3 of total voting power %d", signedVotingPower, totalVotingPower)
	}
	return true, nil
}

// VerifyAggregatedSignature verifies the aggregated signature of the state proof.
func (sp *StateProof) VerifyAggregatedSignature(blsCurve crypto.BLSCurve) (bool, error) {
	pubKeys := make([][]byte, 0)
	for i, pubKey := range sp.BLSPublicKeys {
		if sp.AggregationBits[i] == 1 {
			pubKeys = append(pubKeys, utils.Hex2Bytes(pubKey))
		}
	}
	blsScheme := crypto.NewBLSScheme(blsCurve)
	commitHash := sp.GetBatchHash()
	return blsScheme.VerifyAggregatedSignature(pubKeys, commitHash, utils.Hex2Bytes(sp.AggregatedSignature))
}
