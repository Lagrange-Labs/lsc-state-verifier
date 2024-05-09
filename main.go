package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/Lagrange-Labs/lagrange-node/crypto"
	"github.com/Lagrange-Labs/lagrange-node/logger"
	"github.com/Lagrange-Labs/state-verifier/stateproof"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		logger.Fatalf("Error loading .env file: %s", err)
	}
	apiKey := os.Getenv("API_KEY")
	chainID, err := strconv.ParseUint(os.Getenv("CHAIN_ID"), 10, 32)
	if err != nil {
		logger.Fatalf("Error parsing CHAIN_ID: %s", err)
	}
	batchNumber, err := strconv.ParseUint(os.Getenv("BATCH_NUMBER"), 10, 64)
	if err != nil {
		logger.Fatalf("Error parsing BATCH_NUMBER: %s", err)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.lagrange.dev/batches/state-proofs?chain_id=%d&batch_number=%d", chainID, batchNumber), nil)
	if err != nil {
		logger.Fatalf("Error creating request: %s", err)
	}
	req.Header.Set("x-api-key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Fatalf("Error sending request: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalf("Error reading response body: %s", err)
	}

	var proofData stateproof.StateProof
	if err = json.Unmarshal(body, &proofData); err != nil {
		logger.Fatalf("Error unmarshalling state proof data: %s", err)
	}
	if len(proofData.AggregatedSignature) == 0 {
		logger.Fatalf("No state proof data found, please check the `API_KEY`, `CHAIN_ID`, and `BATCH_NUMBER`")
	}

	// Verify the committee root
	if !proofData.VerifyCommitteeRoot() {
		logger.Fatalf("Committee root verification failed")
	}
	// Verify the voting power
	if ok, err := proofData.VerifyVotingPower(); !ok {
		logger.Fatalf("Voting power verification failed: %s", err)
	}
	// Verify the aggregated signature
	if ok, err := proofData.VerifyAggregatedSignature(crypto.BN254); !ok {
		logger.Fatalf("Aggregated signature verification failed: %s", err)
	}

	logger.Infof("State proof verification successful")
}
