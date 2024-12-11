package http_service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"story-monitor/config"
	"story-monitor/log"
	"story-monitor/types"
	"strconv"
	"strings"
)

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body: %v", err)
	}
	return body, nil
}

func GetLatestBlock() int64 {
	httpUrl := config.GetHttpRpc()

	url := httpUrl + "/status?"
	//url := "https://story-testnet-rpc.polkachu.com/status?"
	body, err := httpGet(url)
	if err != nil {
		logger.Error("getLatestBlock http get error: ", err)

	}
	status := new(types.ChainStatus)
	if err := json.Unmarshal(body, status); err != nil {
		logger.Error("Failed to serialize the acquired chain status, err:", err)

	}
	height, err := strconv.Atoi(status.Result.SyncInfo.LatestBlockHeight)
	if err != nil {
		logger.Error("getLatestBlock parse err:", err)
	}
	return int64(height)
}

func GetValPerformance(start, latestBlock int64, monitorObjs []*types.MonitorObj) ([]*types.ValSign, []*types.ValSignMissed, error) {
	valSigns := make([]*types.ValSign, 0)
	valSignMisseds := make([]*types.ValSignMissed, 0)
	valSignMaps := make(map[string]struct{})

	for i := start; i < latestBlock+1; i++ {
		httpUrl := config.GetHttpRpc()
		url := httpUrl + "/commit?height=" + strconv.Itoa(int(i))
		body, err := httpGet(url)
		if err != nil {
			logger.Error("GetValPerformance http get error: ", err)
			return []*types.ValSign{}, []*types.ValSignMissed{}, err
		}

		block := new(types.CommitBlock)
		if err := json.Unmarshal(body, block); err != nil {
			logger.Error("Failed to serialize the acquired block, err:", err)
			return []*types.ValSign{}, []*types.ValSignMissed{}, err
		}

		for _, valsignature := range block.Result.SignedHeader.Commit.Signatures {
			valSignMaps[strings.ToLower(valsignature.ValidatorAddress)] = struct{}{}
		}
		for _, monitorObj := range monitorObjs {
			if _, ok := valSignMaps[monitorObj.OperatorAddrHex]; ok {
				valSigns = append(valSigns, &types.ValSign{
					//Moniker:      monitorObj.Moniker,
					OperatorAddr: monitorObj.OperatorAddr,
					BlockHeight:  latestBlock,
					Status:       1,
					ChildTable:   latestBlock % 10,
				})
			} else {
				valSignMisseds = append(valSignMisseds, &types.ValSignMissed{
					//Moniker:      monitorObj.Moniker,
					OperatorAddr: monitorObj.OperatorAddr,
					BlockHeight:  latestBlock,
				})
			}
		}
	}

	return valSigns, valSignMisseds, nil
}

func GetValidatorStatus(latestBlock int64, monitorObjs []*types.MonitorObj) ([]*types.MonitorObj, error) {
	httpUrl := config.GetHttpRpc()
	inactiveVal := make([]*types.MonitorObj, 0)
	valInfo := make(map[string]*types.MonitorObj)
	for _, monitorObj := range monitorObjs {
		valInfo[monitorObj.OperatorAddrHex] = monitorObj
	}
	for {
		page := "1"
		url := httpUrl + "validators?height=" + strconv.Itoa(int(latestBlock)) + "&page=" + page + "&per_page=50"
		body, err := httpGet(url)
		if err != nil {
			logger.Error("GetValPerformance http get error: ", err)
			return []*types.MonitorObj{}, err
		}
		valStatus := new(types.ValStatus)
		if err := json.Unmarshal(body, valStatus); err != nil {
			logger.Error("Failed to serialize the acquired validator status, err:", err)
		}

		for _, validator := range valStatus.Result.Validators {
			votingPower, err := strconv.Atoi(validator.VotingPower)
			if err != nil {
				logger.Error("Failed to convert validator voting power, err:", err)
				continue
			}
			if votingPower <= 0 {
				inactiveVal = append(inactiveVal, valInfo[validator.Address])
			}
		}
		pageInt, _ := strconv.Atoi(page)
		total, _ := strconv.Atoi(valStatus.Result.Total)
		threshold := (total + pageInt - 1) / pageInt
		if pageInt <= threshold {
			page = strconv.Itoa(pageInt + 1)
		} else {
			break
		}
	}
	return inactiveVal, nil

}

var logger = log.HTTPLogger.WithField("module", "rpc")
