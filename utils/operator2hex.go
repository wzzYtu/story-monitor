package utils

import (
	"encoding/hex"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"story-monitor/types"
)

func Operator2Hex(bech32Addrs []string) []*types.MonitorObj {
	monitorObjs := make([]*types.MonitorObj, len(bech32Addrs))
	for _, addr := range bech32Addrs {
		_, data, err := bech32.DecodeAndConvert(addr)
		if err != nil {
			logger.Error("Failed to decode Bech32 address:", err)
		}
		monitorObjs = append(monitorObjs, &types.MonitorObj{
			OperatorAddr:    addr,
			OperatorAddrHex: hex.EncodeToString(data),
		})
	}
	return monitorObjs
}
