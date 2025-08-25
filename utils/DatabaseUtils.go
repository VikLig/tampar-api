package utils

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Database struct {
	Database *sql.DB
}

func NewDatabase(config Config, logger Logger) Database {
	database, err := sql.Open(config.DbName, config.DbUsername+":"+config.DbPassword+"@/"+config.DbSid)
	//database, err := sql.Open(config.DbName, config.DbName+"://"+config.DbUsername+":"+config.DbPassword+"@"+config.DbUrl+":"+config.DbPort+"/"+config.DbSid)
	if err != nil {
		logger.Logger.Panic(err)
	}
	database.SetMaxOpenConns(100)                  
	database.SetMaxIdleConns(20)                   
	database.SetConnMaxLifetime(10 * time.Minute) 
	logger.Logger.Info("Database connected succesfully.")
	if err := database.Ping(); err != nil {
		log.Panic(err)
	}
	return Database{Database: database}
}
