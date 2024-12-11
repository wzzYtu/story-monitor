package utils

import (
	"math/big"
	"strconv"
	"time"

	"story-monitor/log"
)

func Div18(dividend string) float64 {
	dividendFloat, ok := new(big.Float).SetString(dividend)
	if !ok {
		logger.Error("Failed to convert commissionRates from string to big.Float")
	}
	result := new(big.Float).Quo(dividendFloat, big.NewFloat(1000000000000000000))
	resStr := result.String()
	res, err := strconv.ParseFloat(resStr, 64)
	if err != nil {
		logger.Error("Failed to convert commissionRates from string to float64")
	}
	return res
}

var RetryFlag = make(chan bool)

func Retry(f func() bool, rules []int) bool {
	index := 0

	for {
		go time.AfterFunc(time.Duration(rules[index])*time.Second, func() {
			RetryFlag <- f()
		})
		if <-RetryFlag {
			return true
		}
		if index == len(rules)-1 {
			return false
		}
		index++
	}
}

var logger = log.Logger.WithField("module", "utils")
