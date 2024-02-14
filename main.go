package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Lagrange-Labs/hash-demo/crypto"
	"github.com/Lagrange-Labs/hash-demo/merkle"
)

type BlockData struct {
	Addrs        []string `json:"address"`
	BLSPubKeys   []string `json:"pubkeys"`
	VotingPowers []uint64 `json:"votingPower"`
}

func main() {
	// TODO: implement getting the data from the query-layer
	// instead of file reading
	file, err := os.Open("block_data.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var blockData BlockData
	if err = json.NewDecoder(file).Decode(&blockData); err != nil {
		panic(err)
	}

	blsScheme := crypto.NewBLSScheme(crypto.BN254)
	leaves := make([][]byte, len(blockData.Addrs))
	for i, addr := range blockData.Addrs {
		rawPubKey, err := blsScheme.GetRawKey(crypto.Hex2Bytes(blockData.BLSPubKeys[i]))
		if err != nil {
			panic(err)
		}
		fmt.Printf("Raw public key for %s: %x\n", addr, rawPubKey)
		leaves[i] = merkle.GetLeafHash(crypto.Hex2Bytes(addr), rawPubKey, blockData.VotingPowers[i])
		fmt.Printf("Leaf hash for %s: %x\n", addr, leaves[i])
	}

	rootHash := merkle.GetRootHash(leaves)
	fmt.Printf("Root hash: %x\n", rootHash) // 1768ba0473721525a355e3c8f16e3a081b124316923f501ae07b2cbfb0673d69
}
