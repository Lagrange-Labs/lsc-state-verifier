package crypto

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/poseidon"
	"golang.org/x/crypto/sha3"
)

// Hex2Bytes converts a hex string to bytes.
func Hex2Bytes(hex string) []byte {
	return common.FromHex(hex)
}

// Bytes2Hex converts bytes to a hex string.
func Bytes2Hex(bytes []byte) string {
	return common.Bytes2Hex(bytes)
}

// PoseidonHash calculates the poseidon hash of elements.
func PoseidonHash(data ...[]byte) []byte {
	msg := []byte{}
	for _, d := range data {
		msg = append(msg, d...)
	}
	hash, err := poseidon.HashBytes(msg)
	if err != nil {
		panic(fmt.Errorf("poseidon hash failed: %v", err))
	}
	return hash.Bytes()
}

// Hash calculates  the keccak hash of elements.
func Hash(data ...[]byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	for _, d := range data {
		hash.Write(d[:]) //nolint:errcheck,gosec
	}
	return hash.Sum(nil)
}
