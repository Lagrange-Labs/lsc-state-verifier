package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Lagrange-Labs/lagrange-node/logger"
	"github.com/Lagrange-Labs/state-verifier/config"
	"github.com/Lagrange-Labs/state-verifier/stateproof"
)

// ProcessChainUsingAPI processes the chain using the API.
func ProcessChainUsingAPI(apiUrl, apiKey string, chain config.ChainConfig) error {
	for {
		err := processBatchFromAPI(apiUrl, apiKey, chain.ChainID, chain.FromBatchNumber)
		if err != nil {
			logger.Errorf("Error processing batch %d for chain ID %d: %s", chain.FromBatchNumber, chain.ChainID, err)
		}
		chain.FromBatchNumber++
		break
		time.Sleep(1 * time.Second)
	}
	return nil
}

// processBatchFromAPI processes the batch from the API.
func processBatchFromAPI(apiUrl, apiKey string, chainID, batchNumber int64) error {
	logger.Infof("Processing batch %d for chain %d", batchNumber, chainID)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/batches/state-proofs?chain_id=%d&batch_number=%d", apiUrl, chainID, batchNumber), nil)
	if err != nil {
		return fmt.Errorf("error creating request: %s", err)
	}
	req.Header.Set("x-api-key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %s", err)
	}

	var proofData stateproof.StateProof
	if err = json.Unmarshal(body, &proofData); err != nil {
		return fmt.Errorf("error unmarshalling state proof data: %s", err)
	}
	if len(proofData.AggregatedSignature) == 0 {
		return fmt.Errorf("no state proof data found for chain ID %d, batch number %d", chainID, batchNumber)
	}

	return verifyStateProof(&proofData)
}