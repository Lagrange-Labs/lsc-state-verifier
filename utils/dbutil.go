package utils

import (
	"fmt"
	"time"

	"github.com/Lagrange-Labs/lagrange-node/logger"
	"github.com/Lagrange-Labs/state-verifier/config"
	"github.com/Lagrange-Labs/state-verifier/db"
)

// ProcessChainUsingDB processes the chain using the DB.
func ProcessChainUsingDB(db db.Database, chain config.ChainConfig) error {
	for {
        err := processBatchFromDB(db, chain.ChainID, chain.FromBatchNumber)
        if err != nil {
            logger.Errorf("Error processing batch %d for chain ID %d: %s", chain.FromBatchNumber, chain.ChainID, err)
        } else {
            // // Update the last processed batch number
            // if err := db.UpdateLastProcessedBatch(chain.ChainID, chain.FromBatchNumber); err != nil {
            //     logger.Errorf("Error updating last processed batch number for chain ID %d: %s", chain.ChainID, err)
            // }
        }
        chain.FromBatchNumber++
        time.Sleep(1 * time.Second)
        break
    }
    return nil
}

// processBatchFromDB processes the batch from the DB.
func processBatchFromDB(db db.Database, chainID, batchNumber int64) error {
	logger.Infof("Processing batch %d for chain %d", batchNumber, chainID)
	result, err := db.GetBatchStateProof(chainID, batchNumber)
	if err != nil {
		return fmt.Errorf("error fetching state proof from DB: %s", err)
	}
	if err := verifyStateProof(result); err != nil {
        if err.Error() == "committee root verification failed" {
            // Add failed batch
            // if err := db.AddFailedBatch(chainID, batchNumber); err != nil {
            //     logger.Errorf("Error adding failed batch %d for chain ID %d: %s", batchNumber, chainID, err)
            // }
			logger.Info("Error")
        }
        return err
    }
    return nil
}