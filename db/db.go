package db

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"

	"story-monitor/log"
	"story-monitor/types"
)

type DBCli interface {
	BatchSaveValSign(valSigns []*types.ValSign) error
	BatchSaveValSignMissed(valSignMissed []*types.ValSignMissed) error
	BatchSaveSignNum(startBlock, endBlock int64, operatorAddrs []string) error
	BatchSaveUptime(startBlock, endBlock int64, operatorAddrs []string) error
	BatchSaveMissedSignNum(startBlock, endBlock int64, operatorAddrs []string) error
	GetBlockHeightFromDb() (int64, error)
	GetValSignMissedFromDb(start, end int64) ([]*types.ValSignMissed, error)
}

type DbCli struct {
	Conn *sqlx.DB
}

func (c *DbCli) BatchSaveValSign(valSigns []*types.ValSign) error {
	batchSize := 500
	for b := 0; b < len(valSigns); b += batchSize {
		logger.Infof("Start saving %d batch of validator sign\n", b+1)
		start := b
		end := b + batchSize
		if len(valSigns) < end {
			end = len(valSigns)
		}
		numArgs := 4
		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*numArgs)

		for i, v := range valSigns[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)",
				i*numArgs+1, i*numArgs+2, i*numArgs+3, i*numArgs+4))
			valueArgs = append(valueArgs, v.OperatorAddr)
			valueArgs = append(valueArgs, v.BlockHeight)
			valueArgs = append(valueArgs, v.Status)
			valueArgs = append(valueArgs, v.BlockHeight%10)
		}

		sql := fmt.Sprintf(`
			INSERT INTO val_sign_p (
				operator_addr,
				block_height,
				status,
				child_table
			)
			VALUES %v
			ON  CONFLICT (operator_addr, block_height, child_table) DO UPDATE SET
				status = EXCLUDED.status,
		`, strings.Join(valueStrings, ","))
		_, err := c.Conn.Exec(sql, valueArgs...)
		if err != nil {
			logger.Errorf("saving validator sign batch %v fail. err:%v \n", b+1, err)
			return err
		}

		logger.Infof("saving validator sign %v completed\n", b+1)
	}
	return nil
}
func (c *DbCli) BatchSaveValSignMissed(valSignMissed []*types.ValSignMissed) error {
	batchSize := 500
	for b := 0; b < len(valSignMissed); b += batchSize {
		logger.Infof("Start saving %d batch of validator sign missed\n", b+1)
		start := b
		end := b + batchSize
		if len(valSignMissed) < end {
			end = len(valSignMissed)
		}
		numArgs := 2
		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*numArgs)

		for i, v := range valSignMissed[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)",
				i*numArgs+1, i*numArgs+2))
			valueArgs = append(valueArgs, v.OperatorAddr)
			valueArgs = append(valueArgs, v.BlockHeight)
		}

		sql := fmt.Sprintf(`
			INSERT INTO val_sign_missed (
				operator_addr,
				block_height
			)
			VALUES %v;
		`, strings.Join(valueStrings, ","))
		_, err := c.Conn.Exec(sql, valueArgs...)
		if err != nil {
			logger.Errorf("saving validator sign missed batch %v fail. err:%v \n", b+1, err)
			return err
		}

		logger.Infof("saving validator sign missed %v completed\n", b+1)
	}
	return nil
}

func (c *DbCli) BatchSaveSignNum(startBlock, endBlock int64, operatorAddrs []string) error {
	for _, operatorAddr := range operatorAddrs {
		logger.Infof("Begin SaveSignNum for %v validator succeeded\n", operatorAddr)
		sql := `
			INSERT INTO val_stats (operator_addr, start_block, end_block, sign_num)
			(
				SELECT operator_addr, $2, $3, COUNT(*) FROM val_sign_p
				WHERE operator_addr = $1 AND block_height >= $2 AND block_height <= $3
				GROUP BY operator_addr
			)
			ON CONFLICT(operator_addr, start_block, end_block) do update set sign_num = excluded.sign_num;
		`
		_, err := c.Conn.Exec(sql, operatorAddr, startBlock, endBlock)
		if err != nil {
			logger.Errorf("Failed to save SaveSignNum for %v validator, err: %v \n", operatorAddr, err)
			return err
		}

		logger.Infof("Save SaveSignNum for %v validator succeeded\n", operatorAddr)
	}

	return nil
}

func (c *DbCli) BatchSaveUptime(startBlock, endBlock int64, operatorAddrs []string) error {
	logger.Infof("begin save uptime")
	valSignNum := make([]*types.ValSignNum, 0)
	valSignNumMap := make(map[string]float64)
	sql := `SELECT operator_addr, sign_num FROM val_stats WHERE start_block = $1 AND end_block = $2`
	err := c.Conn.Select(&valSignNum, sql, startBlock, endBlock)
	if err != nil {
		logger.Errorf("Failed to query validator missed attestation num, err:%v\n", err)
		return err
	}
	for _, v := range valSignNum {
		valSignNumMap[v.OperatorAddr] = float64(v.SignNum)
	}

	batchSize := 10
	for b := 0; b < len(operatorAddrs); b += batchSize {
		logger.Infof("Start saving %d batch of Proposals\n", b+1)
		start := b
		end := b + batchSize
		if len(operatorAddrs) < end {
			end = len(operatorAddrs)
		}
		numArgs := 4
		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*numArgs)

		for i, v := range operatorAddrs[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)",
				i*numArgs+1, i*numArgs+2, i*numArgs+3, i*numArgs+4))
			valueArgs = append(valueArgs, v)
			valueArgs = append(valueArgs, startBlock)
			valueArgs = append(valueArgs, endBlock)
			valueArgs = append(valueArgs, valSignNumMap[v]/(float64(endBlock)-float64(startBlock)+1))
		}

		sql := fmt.Sprintf(`
			INSERT INTO val_stats (
				operator_addr,
				start_block,
				end_block,
				uptime
			)
			VALUES %v
			ON  CONFLICT (operator_addr, start_block, end_block) DO UPDATE SET
				uptime = EXCLUDED.uptime;
		`, strings.Join(valueStrings, ","))
		_, err := c.Conn.Exec(sql, valueArgs...)
		if err != nil {
			logger.Errorf("saving uptime %v fail. err:%v \n", b+1, err)
			return err
		}
		logger.Infof("saving uptime %v completed\n", b+1)
	}
	return nil
}

func (c *DbCli) BatchSaveMissedSignNum(startBlock, endBlock int64, operatorAddrs []string) error {
	for _, operatorAddr := range operatorAddrs {
		logger.Infof("Begin MissedSignNum for %v validator succeeded\n", operatorAddr)
		sql := `
			INSERT INTO val_stats (operator_addr, start_block, end_block, missed_sign_num)
			(
				SELECT operator_addr, $2, $3, COUNT(*) FROM val_sign_missed
				WHERE operator_addr = $1 AND block_height >= $2 AND block_height <= $3
				GROUP BY operator_addr
			)
			ON CONFLICT(operator_addr, start_block, end_block) do update set  moniker = excluded.moniker, missed_sign_num = excluded.missed_sign_num;
		`
		_, err := c.Conn.Exec(sql, operatorAddr, startBlock, endBlock)
		if err != nil {
			logger.Errorf("Failed to save MissedSignNum for %v validator, err: %v \n", operatorAddr, err)
			return err
		}

		logger.Infof("Save MissedSignNum for %v validator succeeded\n", operatorAddr)
	}

	return nil
}

func (c *DbCli) GetBlockHeightFromDb() (int64, error) {
	var minHeight int64
	dbHeight := make([]types.MaxBlockHeight, 0)
	sqld := `select (select max(block_height) from val_sign_p) max_block_height_sign,
       max(block_height) max_block_height_missed from val_sign_p;`
	err := c.Conn.Select(&dbHeight, sqld)
	if err != nil {
		logger.Errorf("Failed to query block height from db, err:%v\n", err)
		return 0, err
	}

	for _, height := range dbHeight {
		var blockHeight int64
		if height.MaxBlockHeightSign.Int64 >= height.MaxBlockHeightMissed.Int64 {
			blockHeight = height.MaxBlockHeightSign.Int64
		} else {
			blockHeight = height.MaxBlockHeightMissed.Int64
		}
		if minHeight == 0 {
			minHeight = blockHeight
		} else {
			if minHeight > blockHeight {
				minHeight = blockHeight
			}
		}
	}
	if minHeight == 0 {
		minHeight = int64(viper.GetInt("alert.startingBlockHeight"))
	}
	return minHeight, nil
}

func (c *DbCli) GetValSignMissedFromDb(start, end int64) ([]*types.ValSignMissed, error) {
	valSignMissed := make([]*types.ValSignMissed, 0)
	sqld := `SELECT block_height, operator_addr FROM val_sign_missed WHERE block_height >= $1 AND block_height <= $2;`
	err := c.Conn.Select(&valSignMissed, sqld, start, end)
	if err != nil {
		logger.Errorf("Failed to query validator sign missed, err:%v\n", err)
		return nil, err
	}
	logger.Info("query validator sign missed successful")
	return valSignMissed, nil
}

func (c *DbCli) BatchSaveValStats(start, end int64) error {
	if start < 0 && end > 0 {
		start = 0
	} else if start < 0 && end == 0 {
		return errors.New("The starting block height is negative and the ending block height is 0")
	}
	allVal := make([]string, 0)
	sqld := `SELECT operator_addr FROM val_info`
	err := c.Conn.Select(&allVal, sqld)
	if err != nil {
		logger.Error("Failed to query all validator, err:", err)
		return err
	}
	err = c.BatchSaveSignNum(start, end, allVal)
	if err != nil {
		logger.Errorf("When the block height is %d to %d, saving the number of validator signatures failed, err:%v \n", start, end, err)
	}

	err = c.BatchSaveMissedSignNum(start, end, allVal)
	if err != nil {
		logger.Errorf("When the block height is %d to %d, it fails to save the number of unsigned validators, err:%v \n", start, end, err)
	}

	err = c.BatchSaveUptime(start, end, allVal)
	if err != nil {
		logger.Errorf("Failed to save validator signature rate when block height is %d to %d, err:%v \n", start, end, err)
	}

	return nil
}

var logger = log.DBLogger.WithField("module", "db")
