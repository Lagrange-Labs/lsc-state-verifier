package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Lagrange-Labs/hash-demo/crypto"
	"github.com/Lagrange-Labs/hash-demo/merkle"
	"github.com/joho/godotenv"
)

type BlockData struct {
	Addrs        []string `json:"address"`
	BLSPubKeys   []string `json:"pubkeys"`
	VotingPowers []uint64 `json:"votingPower"`
}

func main() {

	const (
		OPTIMISM = "11155420"
		ARBITRUM = "421614"
		MANTLE = "5003"
	)
	BLOCK_NUMBER := "14483253"

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("API_KEY")

	req, err := http.NewRequest("GET", fmt.Sprintf("https://querylayer.lagrange.dev/blocks/block-data?chain_id=%s&block_number=%s", ARBITRUM, BLOCK_NUMBER), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("x-api-key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var blockData BlockData
	if err = json.Unmarshal(body, &blockData); err != nil {
		log.Fatal(err)
	}

	blsScheme := crypto.NewBLSScheme(crypto.BN254)
	leaves := make([][]byte, len(blockData.Addrs))
	for i, addr := range blockData.Addrs {
		rawPubKey, err := blsScheme.GetRawKey(crypto.Hex2Bytes(blockData.BLSPubKeys[i]))
		if err != nil {
			panic(err)
		}
		fmt.Printf("\nRaw public key for %s: %x\n", addr, rawPubKey)
		leaves[i] = merkle.GetLeafHash(crypto.Hex2Bytes(addr), rawPubKey, blockData.VotingPowers[i])
		fmt.Printf("Leaf hash for %s: %x\n", addr, leaves[i])
	}

	rootHash := merkle.GetRootHash(leaves)
	fmt.Printf("\nCommittee Root: %x\n", rootHash)
}
