package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"story-monitor/config"
	"story-monitor/log"
	"story-monitor/monitor"
)

func init() {
	var cfg = pflag.StringP("config", "c", "config/.conf.yaml", "config file path.")
	pflag.Parse()
	err := config.InitConfig(*cfg)
	if err != nil {
		fmt.Println("read config file error:", err)
		return
	}
	log.InitLogger(log.Logger, viper.GetString("log.path"))
	log.InitLogger(log.DBLogger, viper.GetString("log.path"))
	log.InitLogger(log.HTTPLogger, viper.GetString("log.path"))
	log.InitLogger(log.MailLogger, viper.GetString("log.path"))
	log.InitLogger(log.ConfigLogger, viper.GetString("log.path"))
	log.InitJsonLogger(log.EventLogger, viper.GetString("log.eventlogpath"))

}

func main() {
	logger := log.Logger.WithField("module", "main")
	logger.Info("Successfully read config file!")
	m, err := monitor.NewMonitor()
	if err != nil {
		logger.Error("Failed to initialize monitoring client")
	}
	logger.Info("Starting Monitor")
	go m.Start()
	go m.ProcessData()
	m.WaitInterrupted()
}
