package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/Lagrange-Labs/hash-demo/crypto"
	"github.com/Lagrange-Labs/hash-demo/merkle"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
)

type BlockData struct {
	Addrs        []string `json:"address"`
	BLSPubKeys   []string `json:"pubkeys"`
	VotingPowers []uint64 `json:"votingPower"`
	AggSignature string   `json:"agg_signature"`
	ChainHeader  struct {
		ChainID     uint32 `json:"chain_id"`
		BlockNumber uint64 `json:"block_number"`
		BlockHash   string `json:"block_hash"`
	} `json:"chain_header"`
	CurrentCommittee string  `json:"currentCommitteeeRoot"`
	NextCommittee    string  `json:"nextCommitteeRoot"`
	AggregationBits  []uint8 `json:"aggregationBits"`
}

func (b BlockData) Hash() []byte {
	var blockNumberBuf common.Hash
	blockHash := common.FromHex(b.ChainHeader.BlockHash)[:]
	blockNumber := big.NewInt(int64(b.ChainHeader.BlockNumber)).FillBytes(blockNumberBuf[:])
	chainID := make([]byte, 4)
	binary.BigEndian.PutUint32(chainID, b.ChainHeader.ChainID)
	chainHash := crypto.Hash(blockHash, blockNumber, chainID)

	committeeRoot := common.FromHex(b.CurrentCommittee)
	nextCommitteeRoot := common.FromHex(b.NextCommittee)
	committeeHash := crypto.PoseidonHash(chainHash, committeeRoot, nextCommitteeRoot)

	return committeeHash
}

func main() {

	const (
		OPTIMISM = "11155420"
		ARBITRUM = "421614"
		MANTLE   = "5003"
	)
	BLOCK_NUMBER := "17970459"

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

	fmt.Printf("Block data: %+v\n", blockData)

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

	if bytes.Equal(rootHash, common.Hex2Bytes(blockData.CurrentCommittee)) {
		commitHash := blockData.Hash()
		pubKeys := make([][]byte, 0)
		for i, pubKey := range blockData.BLSPubKeys {
			if blockData.AggregationBits[i] == 1 {
				pubKeys = append(pubKeys, crypto.Hex2Bytes(pubKey))
			}
		}
		fmt.Printf("\nCommittee hash: %x\n", commitHash)
		fmt.Printf("Aggregated signature: %s\n", blockData.AggSignature)
		fmt.Printf("Public keys: %x\n", pubKeys)
		verified, err := blsScheme.VerifyAggregatedSignature(pubKeys, commitHash, crypto.Hex2Bytes(blockData.AggSignature))
		if err != nil {
			panic(err)
		}
		fmt.Printf("Aggregated signature verified: %t\n", verified)
	} else {
		fmt.Printf("Root hash does not match: %x != %x\n", rootHash, common.Hex2Bytes(blockData.CurrentCommittee))
	}
}
