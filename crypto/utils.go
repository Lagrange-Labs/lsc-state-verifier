package crypto

import (
	"github.com/ethereum/go-ethereum/common"
)

// Hex2Bytes converts a hex string to bytes.
func Hex2Bytes(hex string) []byte {
	return common.FromHex(hex)
}

// Bytes2Hex converts bytes to a hex string.
func Bytes2Hex(bytes []byte) string {
	return common.Bytes2Hex(bytes)
}
