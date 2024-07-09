package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/Lagrange-Labs/lagrange-node/logger"
	"github.com/Lagrange-Labs/state-verifier/config"
	"github.com/Lagrange-Labs/state-verifier/db"
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

	for {
		err := processBatchFromDB(db, chain.ChainID, chain.FromBatchNumber)
		if err != nil {
			if strings.Contains(err.Error(), "committee root verification failed") {
				if err := db.AddFailedBatch(chain.ChainID, chain.FromBatchNumber); err != nil {
					logger.Errorf("Error adding failed batch %d for chain ID %d: %s", chain.FromBatchNumber, chain.ChainID, err)
				}
				// Proceed to the next batch number
				chain.FromBatchNumber++
				time.Sleep(HistoricalBatchSleep)
				continue
			}
			if strings.Contains(err.Error(), "error fetching state proof from DB: no results found") {
				logger.Infof("Batch %d for chain ID %d not found, waiting for next batch", chain.FromBatchNumber, chain.ChainID)
				time.Sleep(NewBatchSleep)
				continue
			}
			// Log other errors and proceed to the next batch number
			logger.Errorf("Error processing batch %d for chain ID %d: %s", chain.FromBatchNumber, chain.ChainID, err)
		} else {
			// Update the last processed batch number
			if err := db.UpdateLastProcessedBatch(chain.ChainID, chain.FromBatchNumber); err != nil {
				logger.Errorf("Error updating last processed batch number for chain ID %d: %s", chain.ChainID, err)
			}
			chain.FromBatchNumber++
			time.Sleep(HistoricalBatchSleep)
		}
	}
}

// processBatchFromDB processes the batch from the DB.
func processBatchFromDB(db db.Database, chainID, batchNumber int64) error {
	logger.Infof("Processing batch %d for chain %d", batchNumber, chainID)
	result, err := db.GetBatchStateProof(chainID, batchNumber)
	if err != nil {
		return fmt.Errorf("error fetching state proof from DB: %s", err)
	}
	if err := verifyStateProof(result); err != nil {
		return err
	}
	return nil
}
