package model

import "database/sql"

type DataExcel struct {
	Mode        string   `json:"mode"`
	UseExcel    string   `json:"useExcel"`
	ExcelFile   []byte   `json:"excelFile"`
	Schema 		[]string `json:"schema"`
	EnvSource   string   `json:"envSource"`
	EnvTarget   string   `json:"envTarget"`
	OutputMode  string   `json:"outputMode"`
}

type ORACLE_OBJECT_TYPE string

const (
	ORACLE_TYPE_OBJECT    ORACLE_OBJECT_TYPE = "TYPE"
	ORACLE_TYPE_TABLE     ORACLE_OBJECT_TYPE = "TABLE"
	ORACLE_TYPE_FUNCTION  ORACLE_OBJECT_TYPE = "FUNCTION"
	ORACLE_TYPE_PROCEDURE ORACLE_OBJECT_TYPE = "PROCEDURE"
	ORACLE_TYPE_PACKAGE   ORACLE_OBJECT_TYPE = "PACKAGE"
	ORACLE_TYPE_VIEW      ORACLE_OBJECT_TYPE = "VIEW"
	ORACLE_TYPE_SEQUENCE  ORACLE_OBJECT_TYPE = "SEQUENCE"
	ORACLE_TYPE_INDEX     ORACLE_OBJECT_TYPE = "INDEX"
	ORACLE_TYPE_LOB       ORACLE_OBJECT_TYPE = "LOB"
)

type OracleDbConfig struct {
	DbName     string
	DbUsername string
	DbPassword string
	DbUrl      string
	DbPort     string
	DbSid      string
	DbEnv      string
}

type OracleDbColumn struct {
	Name      string
	TableName string
	Type      string
	Length    int
	Sequence  int
	RowValue  string
}

type Database struct {
	Database   *sql.DB
	Schema     string
	Enviroment string
}

type OracleUserObject struct {
	ObjectId     int
	ObjectOwner  string
	ObjectName   string
	ObjectType   string
	ObjectStatus string
	ObjectSeq    int
	ObjectEnv	 string
	Ddl          string
	Cell         string
	IsListed 	string
}

type ExcelUserObject struct {
	ObjectOwner  string
	ObjectName   string
	ObjectType   string
	ObjectStatus string
	Remark       string
	UserModified string
	Project      string
}
