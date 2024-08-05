package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Lagrange-Labs/lagrange-node/logger"
	"github.com/Lagrange-Labs/state-verifier/config"
	"github.com/Lagrange-Labs/state-verifier/stateproof"
)

// ProcessChainUsingAPI processes the chain using the API.
func ProcessChainUsingAPI(apiUrl, apiKey string, chain config.ChainConfig) error {
	var prevBatchStateProof *stateproof.StateProof

	for {
		proofData, err := getBatchStateProofFromAPI(apiUrl, apiKey, chain.ChainID, chain.FromBatchNumber)
		if err != nil {
			if strings.Contains(err.Error(), "no results found") {
				logger.Infof("Batch %d for chain ID %d not found, waiting for the batch to be available", chain.FromBatchNumber, chain.ChainID)
				time.Sleep(NewBatchSleep)
				continue
			}
			logger.Errorf("Error processing batch %d for chain ID %d: %s", chain.FromBatchNumber, chain.ChainID, err)
		} else {
			err = verifyStateProof(proofData, prevBatchStateProof)
			if err != nil {
				logger.Errorf("Error verifying state proof for batch %d for chain ID %d: %s", chain.FromBatchNumber, chain.ChainID, err)
			}
		}
		prevBatchStateProof = proofData
		chain.FromBatchNumber++
		time.Sleep(HistoricalBatchSleep)
	}
}

// getBatchStateProofFromAPI fetches the batch state proof from the API.
func getBatchStateProofFromAPI(apiUrl, apiKey string, chainID, batchNumber int64) (*stateproof.StateProof, error) {
	logger.Infof("Processing batch %d for chain %d", batchNumber, chainID)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/batches/state-proofs?chain_id=%d&batch_number=%d", apiUrl, chainID, batchNumber), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	var proofData stateproof.StateProof
	if err = json.Unmarshal(body, &proofData); err != nil {
		return nil, fmt.Errorf("error unmarshalling state proof data: %s", err)
	}
	if len(proofData.AggregatedSignature) == 0 {
		return nil, fmt.Errorf("no state proof data found for chain ID %d, batch number %d", chainID, batchNumber)
	}

	return &proofData, nil
}