package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/Lagrange-Labs/lagrange-node/logger"
	"github.com/Lagrange-Labs/lsc-state-verifier/config"
	"github.com/Lagrange-Labs/lsc-state-verifier/db"
	"github.com/Lagrange-Labs/lsc-state-verifier/stateproof"
)

const (
	HistoricalBatchSleep = 1 * time.Second
	NewBatchSleep        = 5 * time.Minute
)

// ProcessChainUsingDB processes the chain using the DB.
func ProcessChainUsingDB(db db.Database, chain config.ChainConfig) error {
	lastProcessedBatch, err := db.GetLastProcessedBatch(chain.ChainID)
	if err != nil {
		return fmt.Errorf("error fetching last processed batch for chain ID %d: %s", chain.ChainID, err)
	}
	if lastProcessedBatch > chain.FromBatchNumber {
		chain.FromBatchNumber = lastProcessedBatch + 1
	}

	var prevBatchStateProof *stateproof.StateProof
	if chain.FromBatchNumber > 1 {
		prevBatchStateProof, err = db.GetBatchStateProof(chain.ChainID, chain.FromBatchNumber-1)
		if err != nil {
			logger.Warnf("Could not fetch previous batch %d for chain %d: %s", chain.FromBatchNumber-1, chain.ChainID, err)
			prevBatchStateProof = nil
		}
	}

	for {
		result, err := db.GetBatchStateProof(chain.ChainID, chain.FromBatchNumber)
		if err != nil {
			if strings.Contains(err.Error(), "no results found") {
				logger.Infof("Batch %d for chain ID %d not found, waiting for the batch to be available", chain.FromBatchNumber, chain.ChainID)
				time.Sleep(NewBatchSleep)
				continue
			}
			logger.Errorf("Error fetching state proof for batch %d for chain ID %d: %s", chain.FromBatchNumber, chain.ChainID, err)
			continue
		}

		err = verifyStateProof(result, prevBatchStateProof)
		if err != nil {
			if strings.Contains(err.Error(), "committee root verification failed") {
				if err := db.AddFailedBatch(chain.ChainID, chain.FromBatchNumber); err != nil {
					logger.Errorf("Error adding failed batch %d for chain ID %d: %s", chain.FromBatchNumber, chain.ChainID, err)
				}
			} else {
				logger.Errorf("Error processing batch %d for chain ID %d: %s", chain.FromBatchNumber, chain.ChainID, err)
			}
		}

		// Update the last processed batch number after successful processing
		if err := db.UpdateLastProcessedBatch(chain.ChainID, chain.FromBatchNumber); err != nil {
			logger.Errorf("Error updating last processed batch number for chain ID %d: %s", chain.ChainID, err)
		}

		// Update prevBatchStateProof to the current result
		prevBatchStateProof = result

		chain.FromBatchNumber++
		time.Sleep(HistoricalBatchSleep)
	}
}
