package mapper

import (
	"errors"
	"fmt"
	"tampar-api/model"
	"tampar-api/utils"
)

type CommonMapper struct {
	handler     utils.RequestHandler
	logger      utils.Logger
	database    utils.Database
}

func NewCommonMapper(
	handler utils.RequestHandler,
	logger utils.Logger,
	database utils.Database,
) CommonMapper {
	return CommonMapper{
		handler:     handler,
		logger:      logger,
		database:    database,
	}
}

func (m CommonMapper) GetSchema() (results []string, errData error) {
	var (
		result	string
	)
	defer func() {
		if r := recover(); r != nil {
			errData = errors.New(fmt.Sprint(r))
		}
	}()
	rows, errData := m.database.Database.Query(`SELECT DISTINCT DB_USERNAME 
												FROM TAMPAR_CONFIG_DB 
												WHERE STATUS = 'Y' ORDER BY DB_USERNAME;`)
	if errData != nil {
		return results, errData
	}

	defer rows.Close()
	for rows.Next() {
		errData = rows.Scan(
			&result,
		)
		if errData != nil {
			return results, errData
		}
		results = append(results, result)
	}
	return results, errData
}

func (m CommonMapper) GetDBConfig(data model.DataExcel) (results []model.OracleDbConfig, errData error) {
	var (
		result	model.OracleDbConfig
		schemas string
	)
	defer func() {
		if r := recover(); r != nil {
			errData = errors.New(fmt.Sprint(r))
		}
	}()

	for i := range data.Schema {
		if i > 0 {
			schemas = schemas +","+data.Schema[i]
		} else {
			schemas = data.Schema[i]
		}
	}
	rows, errData := m.database.Database.Query(`SELECT DISTINCT DB_NAME, DB_USERNAME, DB_PASSWORD, DB_URL, DB_PORT, DB_SID, DB_ENV 
												FROM TAMPAR_CONFIG_DB WHERE STATUS = 'Y' 
												AND DB_ENV IN (:1,:2) AND DB_USERNAME IN (:3)`, data.EnvSource, data.EnvTarget, data.Schema)
	if errData != nil {
		return results, errData
	}

	defer rows.Close()
	for rows.Next() {
		errData = rows.Scan(
			&result.DbName,
			&result.DbUsername,
			&result.DbPassword,
			&result.DbUrl,
			&result.DbPort,
			&result.DbSid,
			&result.DbEnv,
		)
		if errData != nil {
			return results, errData
		}
		results = append(results, result)
	}

	return results, errData
}