package db

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"

	"story-monitor/types"
)

func InitDB(dbconf *types.DatabaseConfig) (*sqlx.DB, error) {
	port, _ := strconv.Atoi(dbconf.Port)
	pdqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		dbconf.Host, port, dbconf.Username, dbconf.Password, dbconf.Name)
	db, err := sqlx.Connect("postgres", pdqlInfo)
	if err != nil {
		logger.Errorf("Connected failed.err:%v\n", err)
		return nil, err
	}

	dbConnectionTimeout := time.NewTimer(15 * time.Second)
	go func() {
		<-dbConnectionTimeout.C
		logger.Fatalf("timeout while connecting to the database")
	}()
	err = db.Ping()
	if err != nil {
		logger.Errorf("ping db fail, err:%v", err)
	}

	dbConnectionTimeout.Stop()

	db.SetConnMaxIdleTime(time.Second * 30)
	db.SetConnMaxLifetime(time.Second * 60)
	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(200)

	logger.Info("Successfully connected!")
	return db, nil
}
