package merkle

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	leafPrefix  = uint8(1)
	innerPrefix = uint8(2)
)

func GetLeafHash(addr, pubKey []byte, votingPower int) []byte {
	res := make([]byte, 0, 1+32+32+20+12)

	res = append(res, leafPrefix)
	res = append(res, pubKey...)
	res = append(res, addr...)
	res = append(res, common.LeftPadBytes(big.NewInt(int64(votingPower)).Bytes(), 12)...)

	fmt.Printf("Leaf pack: %x\n", res)

	return crypto.Keccak256(res)
}

func GetInnerHash(left, right []byte) []byte {
	res := make([]byte, 0, 1+32+32)

	res = append(res, innerPrefix)
	res = append(res, left...)
	res = append(res, right...)

	return crypto.Keccak256(res)
}

func GetRootHash(leaves [][]byte) []byte {
	count := len(leaves)
	if count == 0 {
		return []byte{}
	}

	leavesCount := 1
	for leavesCount < count {
		leavesCount *= 2
	}

	for i := count; i < leavesCount; i++ {
		leaves = append(leaves, make([]byte, 32))
	}

	for leavesCount > 1 {
		for i := 0; i < leavesCount/2; i++ {
			leaves[i] = GetInnerHash(leaves[i*2], leaves[i*2+1])
		}
		leavesCount /= 2
	}

	return leaves[0]
}
