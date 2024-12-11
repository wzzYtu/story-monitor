package monitor

import (
	"github.com/spf13/viper"
	"os"
	"sort"
	"story-monitor/config"
	"story-monitor/db"
	"story-monitor/http_service"
	"story-monitor/log"
	"story-monitor/notification"
	"story-monitor/types"
	"story-monitor/utils"
	"strings"
	"time"
)

var (
	preValinActive       = make(map[string]struct{}, 0)
	monitorHeight  int64 = 0
)

type Monitor struct {
	DbCli           *db.DbCli
	MailClient      *notification.Client
	termChan        chan os.Signal
	missedSignChan  chan []*types.ValSignMissed
	valIsActiveChan chan []string
}

func NewMonitor() (*Monitor, error) {
	// init DB client
	dbConf := &types.DatabaseConfig{
		Username: viper.GetString("postgres.user"),
		Password: viper.GetString("postgres.password"),
		Name:     viper.GetString("postgres.name"),
		Host:     viper.GetString("postgres.host"),
		Port:     viper.GetString("postgres.port"),
	}
	dbCli, err := db.InitDB(dbConf)
	if err != nil {
		logger.Error("connect database server error: ", err)
	}
	// init email client
	mailClient := notification.NewClient(
		viper.GetString("mail.host"),
		viper.GetInt("mail.port"),
		viper.GetString("mail.username"),
		viper.GetString("mail.Password"),
	)

	return &Monitor{
		DbCli:           &db.DbCli{Conn: dbCli},
		MailClient:      mailClient,
		termChan:        make(chan os.Signal),
		missedSignChan:  make(chan []*types.ValSignMissed),
		valIsActiveChan: make(chan []string),
	}, nil
}
func (m *Monitor) WaitInterrupted() {
	<-m.termChan
	logger.Info("monitor shutdown signal received")
}
func (m *Monitor) Start() {
	mailSender := viper.GetString("mail.sender")
	receiver1 := viper.GetString("mail.receiver")
	mailReceiver := strings.Join([]string{receiver1}, ",")

	epochTicker := time.NewTicker(time.Duration(viper.GetInt("alert.timeInterval")) * time.Second)
	for range epochTicker.C {
		// list validator indices from config file
		logger.Info("Getting validators from config file.")
		operatorAdds := config.GetoperatorAddrs()
		monitorObjs := utils.Operator2Hex(operatorAdds)
		latestBlock := http_service.GetLatestBlock()

		logger.Info("start getting validators performance")
		startBlock, err := m.DbCli.GetBlockHeightFromDb()
		if err != nil {
			logger.Error("Failed to query block height from database，err:", err)
		}
		valSign, valSignMissed, err := http_service.GetValPerformance(startBlock, latestBlock, monitorObjs)
		if err != nil {
			logger.Error("get proposal error: ", err)
			res := utils.Retry(func() bool {
				valSign, valSignMissed, err = http_service.GetValPerformance(startBlock, latestBlock, monitorObjs)
				if err != nil {
					return false
				} else {
					return true
				}
			}, []int{1, 3})
			if !res {
				m.MailClient.SendMail(mailSender, mailReceiver, "RPC Exception", "get cared data from RPC node error, please check.")
				continue
			}
		}
		logger.Info("Successfully get validators performance")

		logger.Info("start getting validators status")

		inactiveVal, err := http_service.GetValidatorStatus(latestBlock, monitorObjs)
		if err != nil {
			logger.Error("get proposal error: ", err)
			res := utils.Retry(func() bool {
				inactiveVal, err = http_service.GetValidatorStatus(latestBlock, monitorObjs)
				if err != nil {
					return false
				} else {
					return true
				}
			}, []int{1, 3})
			if !res {
				m.MailClient.SendMail(mailSender, mailReceiver, "RPC Exception", "get cared data from RPC node error, please check.")
				continue
			}
		}
		logger.Info("Successfully get validators status")

		m.processData(&types.CaredData{
			ValSigns:       valSign,
			ValSignMisseds: valSignMissed,
			InactiveVal:    inactiveVal,
		})

	}
}

func (m *Monitor) processData(caredData *types.CaredData) {
	if len(caredData.InactiveVal) > 0 {
		valIsActive := make([]string, 0)
		newpreValinActive := make(map[string]struct{}, 0)

		for _, valInfo := range caredData.InactiveVal {
			if _, ok := preValinActive[valInfo.OperatorAddr]; !ok {
				valIsActive = append(valIsActive, valInfo.OperatorAddr)
			}
			newpreValinActive[valInfo.OperatorAddr] = struct{}{}
		}

		preValinActive = newpreValinActive
		m.valIsActiveChan <- valIsActive
	}

	var end int64

	if caredData.ValSigns != nil && len(caredData.ValSigns) > 0 {
		logger.Info("Start saving validator signs")
		err := m.DbCli.BatchSaveValSign(caredData.ValSigns)
		if err != nil {
			logger.Error("save validator sign fail:", err)
		}
		logger.Info("Save the validator sign successfully")
	}

	if caredData.ValSignMisseds != nil && len(caredData.ValSignMisseds) > 0 {
		logger.Info("Start saving validator sign misseds")
		err := m.DbCli.BatchSaveValSignMissed(caredData.ValSignMisseds)
		if err != nil {
			logger.Error("save validator sign missed fail:", err)
		}
		logger.Info("Save the validator sign missed successfully")

		end, err = m.DbCli.GetBlockHeightFromDb()
		if err != nil {
			logger.Error("Failed to query block height from database，err:", err)
		}
		interval := viper.GetInt("alert.blockInterval")
		start := end - int64(interval) + int64(1)
		valSignMissed, err := m.DbCli.GetValSignMissedFromDb(start, end)
		if err != nil {
			logger.Error("Failed to query validator sign missed from database, err:", err)
		}
		valSignMissedMap := make(map[string][]int)
		for _, signMissed := range valSignMissed {
			valSignMissedMap[signMissed.OperatorAddr] = append(valSignMissedMap[signMissed.OperatorAddr], int(signMissed.BlockHeight))
		}

		proportion := viper.GetFloat64("alert.proportion")
		missedSign := make([]*types.ValSignMissed, 0)
		recorded := make(map[string]struct{}, 0)

		for operatorAddr, missedBlcoks := range valSignMissedMap {
			if float64(len(missedBlcoks))/float64(interval) > proportion {
				missedSign = append(missedSign, &types.ValSignMissed{
					OperatorAddr: operatorAddr,
					BlockHeight:  end,
				})
				recorded[operatorAddr] = struct{}{}
			}

			if len(missedBlcoks) > 5 {
				sort.Ints(missedBlcoks)
				for i := 0; i < len(missedBlcoks)-5; i++ {
					if _, ok := recorded[operatorAddr]; !ok && missedBlcoks[i+4]-missedBlcoks[i] == 4 {
						missedSign = append(missedSign, &types.ValSignMissed{
							OperatorAddr: operatorAddr,
							BlockHeight:  end,
						})
						recorded[operatorAddr] = struct{}{}
					}
				}
			}

		}
		m.missedSignChan <- missedSign
	}

	timeInterval := viper.GetInt("alert.timeInterval")
	endHeight := end / int64(timeInterval) * int64(timeInterval)
	if monitorHeight != endHeight {
		m.DbCli.BatchSaveValStats(endHeight-int64(timeInterval)+int64(1), endHeight)
		monitorHeight = endHeight
	}
}

func (m *Monitor) ProcessData() {
	mailSender := viper.GetString("mail.sender")
	receiver := viper.GetString("mail.receiver")
	mailReceiver := strings.Join([]string{receiver}, ",")
	for {
		select {
		case valisAtive := <-m.valIsActiveChan:
			if len(valisAtive) == 0 {
				break
			}
			va := notification.ParseValisActiveException(valisAtive)
			err := m.MailClient.SendMail(mailSender, mailReceiver, va.Name(), va.Message())
			if err != nil {
				eventLogger.Error("send  validator inActive email error: ", err)
			}
			eventLogger.Info("send validator inActive email successful")
		case <-time.After(time.Second):
		}
		select {
		case missedSign := <-m.missedSignChan:
			if len(missedSign) == 0 {
				break
			}
			va := notification.ParseSyncException(missedSign)
			err := m.MailClient.SendMail(mailSender, mailReceiver, va.Name(), va.Message())
			if err != nil {
				eventLogger.Error("send sign missed email error: ", err)
			}
			eventLogger.Info("send validator sign missed email successful")
		case <-time.After(time.Second):
		}
	}
}

var logger = log.Logger.WithField("module", "monitor")
var eventLogger = log.EventLogger.WithField("module", "event")
