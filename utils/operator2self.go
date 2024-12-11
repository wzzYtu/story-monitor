package utils

import (
	"github.com/cosmos/btcutil/bech32"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosBech32 "github.com/cosmos/cosmos-sdk/types/bech32"
)

func Operator2SelfAdde(operator string) string {
	hrp, bz, err := cosmosBech32.DecodeAndConvert(operator)
	if err != nil {
		logger.Error("error decoding bech32 operator", err)
	}
	if hrp != "storyvaloper" {
		logger.Error("Please enter the validator address starting with storyvaloper")
	}

	operatorBytes := []byte(sdk.ValAddress(bz))
	operatorByte, err := bech32.ConvertBits(operatorBytes, 8, 5, true)
	if err != nil {
		logger.Error("Failed to convert string to bytes, Error:", err)
	}

	selfaddr, err := bech32.Encode("story", operatorByte)
	if err != nil {
		logger.Error("Failed to convert hex format to bech32, err:", err)
	}

	return selfaddr
}
