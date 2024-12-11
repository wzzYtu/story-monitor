package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func Puk2addrHex(pubkey []byte) string {
	pubkeyBaseHash256 := sha256.Sum256(pubkey[2:])
	addrDate := pubkeyBaseHash256[:20]
	logger.Info("Convert public key to Base64 format successfully")
	return hex.EncodeToString(addrDate)
}
