package service

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"tampar-api/model"
	"time"

	"github.com/xuri/excelize/v2"
)

const (
	ORACLE_TYPE_OBJECT     = "TYPE"
	ORACLE_TYPE_TABLE      = "TABLE"
	ORACLE_TYPE_FUNCTION   = "FUNCTION"
	ORACLE_TYPE_PROCEDURE  = "PROCEDURE"
	ORACLE_TYPE_PACKAGE    = "PACKAGE"
	ORACLE_TYPE_VIEW       = "VIEW"
	ORACLE_TYPE_MV         = "MATERIALIZED VIEW"
	ORACLE_TYPE_SEQUENCE   = "SEQUENCE"
	ORACLE_TYPE_INDEX      = "INDEX"
	ORACLE_OBJECT_EXEPTION = "EXCEPTION"
	ORACLE_TYPE_TRIGGER  = "TRIGGER"
	ACTION_MOVE_FILE       = "MOVE"
	ACTION_READ_MOVE_FILE  = "READ_MOVE"

	MODE_DIR_PATH = "MODE_DIR_PATH"
	MODE_DIR_FILE = "MODE_DIR_FILE"

	MODE_GET_OBJECT_EXCEL        = "MODE_GET_OBJECT_EXCEL"
	MODE_GET_OBJECT_LAST_COMPILE = "MODE_GET_OBJECT_LAST_COMPILE"
	MODE_GET_OBJECT_COMBINE      = "MODE_GET_OBJECT_COMBINE"
	MODE_GET_OBJECT_ALL          = "MODE_GET_OBJECT_ALL"

	MODE_GET_OBJ_BY_EXCEL      = 99
	MODE_GET_OBJ_DEFINE_SCHEMA = 100
	GENERAL_PATH               = "C:\\SM_TOOL\\EPROC\\"

	OBJ_STS_MOD_LISTED = "MOD_LISTED"
	OBJ_STS_NEW= "NEW"
	OBJ_STS_MOD_NOT_LISTED= "NOT_LISTED"
	OBJ_STS_MISSING_TARGET= "MISSING_TARGET"
	OBJ_STS_EQUALS= "EQUALS"
	OBJ_STS_MISSING_SOURCE= "MISSING_SOURCE"
)

func NewOracleDatabase(config model.OracleDbConfig) (model.Database, error) {
	database, errData := sql.Open(config.DbName, config.DbName+"://"+config.DbUsername+":"+config.DbPassword+"@"+config.DbUrl+":"+config.DbPort+"/"+config.DbSid)
	if errData != nil {
		log.Println(errData)
	}
	return model.Database{Database: database, Schema: config.DbUsername, Enviroment: config.DbEnv}, errData
}

func GetOracleDBForCompare(listDbConfig []model.OracleDbConfig, data model.DataExcel)([]model.Database, error){
	var errData error
	oraDbList := make([]model.Database, 0)
	
	oraSourceDbList, errData := GetOraSource(listDbConfig, data.Schema, data.EnvSource)
	if errData != nil {
		return nil, errData
	}
	oraTargetDbList, errData := GetOraSource(listDbConfig, data.Schema, data.EnvTarget)
	if errData != nil {
		return nil, errData
	}
	oraDbList = append(oraDbList, oraSourceDbList...)
	oraDbList = append(oraDbList, oraTargetDbList...)
	return oraDbList, errData
}

func GetObjectFromExcel(f *excelize.File) (obj []model.OracleUserObject, exclude []model.OracleUserObject, errData error) {
	var (
		data model.OracleUserObject
	)
	//Read New Object DB
	rowsTable, _ := f.GetRows(ORACLE_TYPE_TABLE)
	rowsView, _ := f.GetRows(ORACLE_TYPE_VIEW)
	rowsMv, _ := f.GetRows(ORACLE_TYPE_MV)
	rowsSeq, _ := f.GetRows(ORACLE_TYPE_SEQUENCE)
	rowsIndex, _ := f.GetRows(ORACLE_TYPE_INDEX)
	rowsType, _ := f.GetRows(ORACLE_TYPE_OBJECT)
	rowsFunction, _ := f.GetRows(ORACLE_TYPE_FUNCTION)
	rowsProcedure, _ := f.GetRows(ORACLE_TYPE_PROCEDURE)
	rowsTrigger, _ := f.GetRows(ORACLE_TYPE_TRIGGER)
	
	data.IsListed = "Y"
	if len(rowsTable) > 1 {
		data.ObjectType = ORACLE_TYPE_TABLE
		for i, row := range rowsTable[1:] {
			if len(row) > 0 {
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.ObjectName = strings.TrimSpace(row[1])
				data.Remark = strings.TrimSpace(row[2])
				data.Pic = strings.TrimSpace(row[3])
				data.Cell = fmt.Sprintf("A%d", i+1)
				data.ObjectSeq = 1
				obj = append(obj, data)
			}
		}
	}

	if len(rowsView) > 1 {
		data.ObjectType = ORACLE_TYPE_VIEW
		for i, row := range rowsView[1:] {
			if len(row) > 0 {
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.ObjectName = strings.TrimSpace(row[1])
				data.Remark = strings.TrimSpace(row[2])
				data.Pic = strings.TrimSpace(row[3])
				data.Cell = fmt.Sprintf("A%d", i+1)
				data.ObjectSeq = 2
				obj = append(obj, data)
			}
		}
	}

	if len(rowsMv) > 1 {
		data.ObjectType = ORACLE_TYPE_MV
		for i, row := range rowsMv[1:] {
			if len(row) > 0 {
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.ObjectName = strings.TrimSpace(row[1])
				data.Remark = strings.TrimSpace(row[2])
				data.Pic = strings.TrimSpace(row[3])
				data.Cell = fmt.Sprintf("A%d", i+1)
				data.ObjectSeq = 3
				obj = append(obj, data)
			}
		}
	}

	if len(rowsSeq) > 1 {
		data.ObjectType = ORACLE_TYPE_SEQUENCE
		for i, row := range rowsSeq[1:] {
			if len(row) > 0 {
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.ObjectName = strings.TrimSpace(row[1])
				data.Remark = strings.TrimSpace(row[2])
				data.Pic = strings.TrimSpace(row[3])
				data.Cell = fmt.Sprintf("A%d", i+1)
				data.ObjectSeq = 4
				obj = append(obj, data)
			}
		}
	}

	if len(rowsIndex) > 1 {
		data.ObjectType = ORACLE_TYPE_INDEX
		for i, row := range rowsIndex[1:] {
			if len(row) > 0 {
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.ObjectName = strings.TrimSpace(row[1])
				data.Remark = strings.TrimSpace(row[2])
				data.Pic = strings.TrimSpace(row[3])
				data.Cell = fmt.Sprintf("A%d", i+1)
				data.ObjectSeq = 5
				obj = append(obj, data)
			}
		}
	}

	if len(rowsType) > 1 {
		data.ObjectType = ORACLE_TYPE_OBJECT
		for i, row := range rowsType[1:] {
			if len(row) > 0 {
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.ObjectName = strings.TrimSpace(row[1])
				data.Remark = strings.TrimSpace(row[2])
				data.Pic = strings.TrimSpace(row[3])
				data.Cell = fmt.Sprintf("A%d", i+1)
				data.ObjectSeq = 6
				obj = append(obj, data)
			}
		}
	}
	if len(rowsFunction) > 1 {
		data.ObjectType = ORACLE_TYPE_FUNCTION
		for i, row := range rowsFunction[1:] {
			if len(row) > 0 {
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.ObjectName = strings.TrimSpace(row[1])
				data.Remark = strings.TrimSpace(row[2])
				data.Pic = strings.TrimSpace(row[3])
				data.Cell = fmt.Sprintf("A%d", i+1)
				data.ObjectSeq = 7
				obj = append(obj, data)
			}
		}
	}
	if len(rowsProcedure) > 1 {
		data.ObjectType = ORACLE_TYPE_PROCEDURE
		for i, row := range rowsProcedure[1:] {
			if len(row) > 0 {
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.ObjectName = strings.TrimSpace(row[1])
				data.Remark = strings.TrimSpace(row[2])
				data.Pic = strings.TrimSpace(row[3])
				data.Cell = fmt.Sprintf("A%d", i+1)
				data.ObjectSeq = 8
				obj = append(obj, data)
			}
		}
	}
	
	if len(rowsTrigger) > 1 {
		data.ObjectType = ORACLE_TYPE_TRIGGER
		for i, row := range rowsTrigger[1:] {
			if len(row) > 0 {
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.ObjectName = strings.TrimSpace(row[1])
				data.Remark = strings.TrimSpace(row[2])
				data.Pic = strings.TrimSpace(row[3])
				data.Cell = fmt.Sprintf("A%d", i+1)
				data.ObjectSeq = 9
				obj = append(obj, data)
			}
		}
	}

	objExc := make([]model.OracleUserObject, 0)
	rowsException, _ := f.GetRows(ORACLE_OBJECT_EXEPTION)
	if len(rowsException) > 1 {
		for _, row := range rowsException[1:] {
			if len(row) > 0 {
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.ObjectName = strings.TrimSpace(row[1])
				data.ObjectType = strings.TrimSpace(row[2])
				data.Pic = strings.TrimSpace(row[3])
				objExc = append(objExc, data)
			}
		}
		exclude = append(exclude, objExc...)
	}

	//obj = DistinctSlice(obj)
	obj = FilterException(obj, objExc)

	errData = ValidateObjectExcel(obj)
	if errData != nil {
		return nil, nil, errData
	}

	return obj, exclude, errData
}

func ValidateObjectExcel(objDb []model.OracleUserObject)error{
	type ErrObj struct {
        ObjectOwner    string
        ObjectName     string
        ObjectType 	   string
    }
	listError := make([]ErrObj,0)
	for _, o := range objDb {
		if o.ObjectOwner == "" || o.ObjectName == "" || o.ObjectType == "" {
			listError = append(listError, ErrObj{ObjectOwner: o.ObjectOwner, ObjectName: o.ObjectName, ObjectType: o.ObjectType}) 
		}
	}
	if len(listError) > 0 {
		return fmt.Errorf("%+v\n", listError)
	}
	return nil
}

func GetSchemaByObject(userObjects []model.OracleUserObject) (owner []string) {
	m := map[string]bool{}
	for _, v := range userObjects {
		if !m[v.ObjectOwner] {
			m[v.ObjectOwner] = true
			owner = append(owner, v.ObjectOwner)
		}
	}
	return owner
}

func GetOraSource(listDbConfig []model.OracleDbConfig, listSchema []string, env string) (oraSourceDbList []model.Database, errData error) {
	// var (
	// 	oraSourceDbList []model.Database
	// 	errData error
	// )

	for i := range listSchema{
		found := false
		for _, db := range listDbConfig {
			if db.DbEnv == env && db.DbUsername == listSchema[i]{
				openDb, errData := NewOracleDatabase(db)
				if errData != nil {
					goto errorDb
				}
				oraSourceDbList = append(oraSourceDbList, openDb)
				found = true
				break
			}
		}
		if !found {
			errData = fmt.Errorf("DB Config for %s Environment %s Not Found", listSchema[i], env)
			goto errorDb
		}
	}

	if len(oraSourceDbList) == 0 {
		return nil, errors.New("Oracle Connection empty")
	}

	errorDb:
	if len(oraSourceDbList) > 0 {
		for _, db := range oraSourceDbList {
			db.Database.Close()
		}
	}
	return oraSourceDbList, errData
}

func GetListObjectDb(oraDbList []model.Database, listObjectExcel []model.OracleUserObject, data model.DataExcel) ([]model.OracleUserObject) {
	listObjectDbResult := make([]model.OracleUserObject, 0)
	
	for _, db := range oraDbList {
		listObjectDbs := make([]model.OracleUserObject, 0)
		if (data.Mode == "GENERATE" && data.UseExcel == "Y") || (data.Mode == "COMPARE" && data.UseExcel == "Y" && data.OutputMode == "EXCEL") {
			listObjectDbs = GetObjectBySchema(listObjectExcel, db.Schema)
		} else {
			listObjectDbs = GetObjects(db)
		}
		
		if len(listObjectDbs) > 0 {
			OrderObjDb(listObjectDbs)
			errData := GetDdl(db, &listObjectDbs, db.Schema)
			if errData != nil {
				log.Println(errData)
			}
			listObjectDbResult = append(listObjectDbResult, listObjectDbs...)
		}
		db.Database.Close()
	}

	//Set ddl to listObjectExcel
	if data.Mode == "COMPARE" && data.UseExcel == "Y"  {
		for i := range listObjectExcel {
			for j := range listObjectDbResult {
				if strings.EqualFold(listObjectDbResult[j].ObjectEnv, data.EnvSource) && strings.EqualFold(listObjectExcel[i].ObjectOwner, listObjectDbResult[j].ObjectOwner) && strings.EqualFold(listObjectExcel[i].ObjectType, listObjectDbResult[j].ObjectType) && 
				strings.EqualFold(listObjectExcel[i].ObjectName, listObjectDbResult[j].ObjectName) {
					listObjectExcel[i].Ddl = listObjectDbResult[j].Ddl
					listObjectExcel[i].Status = listObjectDbResult[j].Status//Valid-Invalid
					listObjectDbResult[j].IsListed = listObjectExcel[i].IsListed
					listObjectDbResult[j].Remark = listObjectExcel[i].Remark
					listObjectDbResult[j].Pic = listObjectExcel[i].Pic
					if listObjectExcel[i].ObjectStatus != "" {
						listObjectDbResult[j].ObjectStatus = listObjectDbResult[j].ObjectStatus+", "+listObjectExcel[i].ObjectStatus
					}
					listObjectExcel[i].ObjectStatus = listObjectDbResult[j].ObjectStatus
					break
				}
			}
		}
	}

	return DistinctObjectDB(listObjectDbResult)
}

func OrderObjDb(userObjects []model.OracleUserObject){
	sort.Slice(userObjects, func(i, j int) bool {
		if userObjects[i].ObjectEnv != userObjects[j].ObjectEnv {
			return userObjects[i].ObjectEnv < userObjects[j].ObjectEnv
		}

		if userObjects[i].ObjectOwner != userObjects[j].ObjectOwner {
			return userObjects[i].ObjectOwner < userObjects[j].ObjectOwner
		}

		if userObjects[i].ObjectSeq != userObjects[j].ObjectSeq {
			return userObjects[i].ObjectSeq < userObjects[j].ObjectSeq
		}
		return userObjects[i].ObjectName < userObjects[j].ObjectName
	})
}

func DistinctObjectDB(s []model.OracleUserObject) []model.OracleUserObject {
	m := map[string]bool{}
	var unique []model.OracleUserObject
	for _, v := range s {
		str := fmt.Sprintf("%s#%s#%s#%s", v.ObjectEnv, v.ObjectOwner, v.ObjectType, v.ObjectName)
		if !m[str] {
			m[str] = true
			v.ObjectSeq = ObjSeq(v.ObjectType)
			unique = append(unique, v)
		}
	}
	return unique
}

func CreateFileObjectDB(userObjects []model.OracleUserObject, env string, fileName string)([]byte, string, error) {
	userObjectsMap := make(map[string][]model.OracleUserObject)

	//timestamp := time.Now().Format("20060102150405")
	baseFolder := fileName
	zipFileName := baseFolder + ".zip"

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	OrderObjDb(userObjects)

	//grouping objects by owner
	for _, userObject := range userObjects {
		if _, ok := userObjectsMap[userObject.ObjectOwner]; !ok {
			userObjectsMap[userObject.ObjectOwner] = []model.OracleUserObject{}
		}
		userObjectsMap[userObject.ObjectOwner] = append(userObjectsMap[userObject.ObjectOwner], userObject)
	}

	for owner, objDBs := range userObjectsMap {
		allObjectSQL := new(strings.Builder)
		for _, userObject := range objDBs {
			if userObject.Ddl == "" {
				continue
			}
			userObject.Ddl = strings.ReplaceAll(userObject.Ddl, " EDITIONABLE ", " ")
			userObject.Ddl = strings.TrimSpace(userObject.Ddl)
			
			// Tampar-Object-DB-<time>/<ENV>/GENERATED/<OWNER>/<OBJECT TYPE>/<ObjectName>.sql
			filePath := filepath.Join(
				baseFolder,
				env,
				"GENERATED",
				userObject.ObjectOwner,
				userObject.ObjectType,
				userObject.ObjectName+".sql",
			)

			allObjectSQL.WriteString( "\n"+userObject.Ddl + "\n")
			allObjectSQL.WriteString("/")

			// Buat file dalam zip
			f, errData := zipWriter.Create(filePath)
			if errData != nil {
				return nil, "", errData
			}
			_, errData = f.Write([]byte(userObject.Ddl))
			if errData != nil {
				return nil, "", errData
			}
		}
		//Tampar-Object-DB-<time>/<ENV>/GENERATED/<OWNER>/All-OBJECT-<OWNER>-<ENV>.sql
		allObjectFilePath := filepath.Join(
			baseFolder,
			env,
			"GENERATED",
			owner,
			fmt.Sprintf("ALL-OBJECT-%s-%s.sql",owner, env),
		)

		f, errData := zipWriter.Create(allObjectFilePath)
		if errData != nil {
			return nil, "", errData
		}

		_, errData = f.Write([]byte(allObjectSQL.String()))
		if errData != nil {
			return nil, "", errData
		}
	}

	// Tutup zip writer
	errData := zipWriter.Close()
	if errData != nil {
		return nil, "", errData
	}

	return buf.Bytes(), zipFileName, nil
}

func GetDdl(conn model.Database, userObjects *[]model.OracleUserObject, schema string) error {

	_, errData := conn.Database.Exec("BEGIN DBMS_METADATA.SET_TRANSFORM_PARAM(DBMS_METADATA.SESSION_TRANSFORM,'STORAGE',FALSE); END;")
	if errData != nil {
		log.Println(errData)
		return errData
	}

	_, errData = conn.Database.Exec("BEGIN DBMS_METADATA.SET_TRANSFORM_PARAM(DBMS_METADATA.SESSION_TRANSFORM,'SEGMENT_ATTRIBUTES',FALSE); END;")
	if errData != nil {
		log.Println(errData)
		return errData
	}

	_, errData = conn.Database.Exec("BEGIN DBMS_METADATA.SET_TRANSFORM_PARAM(DBMS_METADATA.SESSION_TRANSFORM, 'EMIT_SCHEMA', FALSE); END;")
	if errData != nil {
		log.Println(errData)
		return errData
	}

	_, errData = conn.Database.Exec("BEGIN DBMS_METADATA.SET_TRANSFORM_PARAM(DBMS_METADATA.SESSION_TRANSFORM,'SQLTERMINATOR', TRUE); END;")
	if errData != nil {
		log.Println(errData)
		return errData
	}

	for i, userObject := range *userObjects {
		if strings.TrimSpace(userObject.ObjectOwner) != "" && !strings.EqualFold(strings.TrimSpace(userObject.ObjectOwner), schema) {
			(*userObjects)[i].ObjectStatus = "Object Owner Empty or Not Valid"
			continue
		}

		rows, errData := conn.Database.Query("SELECT DBMS_METADATA.GET_DDL('" + strings.TrimSpace(userObject.ObjectType) + "', '" + strings.TrimSpace(userObject.ObjectName) + "', '" + strings.TrimSpace(schema) + "') FROM DUAL")
		if errData != nil {
			(*userObjects)[i].ObjectStatus = errData.Error()
			continue
		}

		defer rows.Close()
		for rows.Next() {
			errData = rows.Scan(&(*userObjects)[i].Ddl)

			if errData != nil {
				(*userObjects)[i].ObjectStatus = errData.Error()
				continue
			}
		}
	}
	return errData
}

func GetObjectBySchema(listDb []model.OracleUserObject, schema string)([]model.OracleUserObject){
	listObjDb := make([]model.OracleUserObject, 0)
	for _, d := range listDb {
		if strings.EqualFold(d.ObjectOwner, schema) {
			listObjDb = append(listObjDb, d)
		}
	}
	return DistinctObjectDB(listObjDb)
}

func GetObjectByEnv(listDb []model.OracleUserObject, envSource string, envTarget string)([]model.OracleUserObject, []model.OracleUserObject){
	listObjDbSrc := make([]model.OracleUserObject, 0)
	listObjDbTrg := make([]model.OracleUserObject, 0)
	for _, d := range listDb {
		if strings.EqualFold(d.ObjectEnv, envSource) {
			listObjDbSrc = append(listObjDbSrc, d)
		} else if strings.EqualFold(d.ObjectEnv, envTarget) {
			listObjDbTrg = append(listObjDbTrg, d)
		}
	}
	return DistinctObjectDB(listObjDbSrc), DistinctObjectDB(listObjDbTrg)
}

// GetObjects retrieves user objects from the database and returns them as a slice of OracleUserObject.
func GetObjects(conn model.Database) []model.OracleUserObject {
	var (
		results []model.OracleUserObject
		result  model.OracleUserObject
	)
	rows, errData := conn.Database.Query(
			`SELECT DISTINCT OBJECT_NAME, OBJECT_TYPE, STATUS, SEQ FROM (
			SELECT OBJECT_NAME, OBJECT_TYPE, STATUS,
			CASE WHEN OBJECT_TYPE = 'TABLE' THEN 1
				WHEN OBJECT_TYPE = 'VIEW' THEN 2 
				WHEN OBJECT_TYPE = 'MATERIALIZED VIEW' THEN 3 
				WHEN OBJECT_TYPE = 'SEQUENCE' THEN 4
				WHEN OBJECT_TYPE = 'INDEX' THEN 5
				WHEN OBJECT_TYPE = 'TYPE' THEN 6 
				WHEN OBJECT_TYPE = 'FUNCTION' THEN 7 
				WHEN OBJECT_TYPE = 'PROCEDURE' THEN 8
				WHEN OBJECT_TYPE = 'TRIGGER' THEN 9
				ELSE 10
			END SEQ
			FROM USER_OBJECTS
			WHERE  OBJECT_TYPE IN ('TABLE', 'VIEW', 'MATERIALIZED VIEW', 'SEQUENCE', 'INDEX', 'TYPE', 'FUNCTION', 'PROCEDURE', 'TRIGGER')
			AND UPPER(OBJECT_NAME) NOT LIKE 'SYS.%' AND UPPER(OBJECT_NAME) NOT LIKE 'SYS_%' AND UPPER(OBJECT_NAME) NOT LIKE '%$%'
			)US 
			ORDER BY 
				US.SEQ ASC, 
				(SELECT CASE WHEN EXISTS(SELECT/*+FIRST_ROW(1)*/ 1 FROM ALL_DEPENDENCIES WHERE OWNER = '` + conn.Schema + `' AND NAME = US.OBJECT_NAME AND REFERENCED_TYPE IS NOT NULL AND ROWNUM = 1) THEN 1 ELSE 0 END FROM DUAL) ASC, 
				US.OBJECT_NAME ASC`)
	if errData != nil {
		log.Println(errData)
		return nil
	}

	defer rows.Close()
	for rows.Next() {
		errData = rows.Scan(&result.ObjectName, &result.ObjectType, &result.Status, &result.ObjectSeq)
		if errData != nil {
			result.ObjectStatus = errData.Error()
		}
		result.ObjectOwner = conn.Schema
		result.ObjectEnv = conn.Enviroment
		results = append(results, result)
	}
	return results
}

func CreateFileObjectDBCompare(userObjects []model.OracleUserObject, userObjectsExcel []model.OracleUserObject, data model.DataExcel)([]byte, string, error) {
	userObjectsMap := make(map[string][]model.OracleUserObject)

	timestamp := time.Now().Format("20060102150405")
	baseFolder := data.FileName
	zipFileName := baseFolder + ".zip"

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	OrderObjDb(userObjects)
	OrderObjDb(userObjectsExcel)

	summaryFilename := fmt.Sprintf("Summary-%s.xlsx", timestamp)
	summaryByte, errData := CreateSummaryCompare(userObjects, userObjectsExcel, data)
	if errData != nil {
		return nil, "", errData
	}
	// Tampar-Object-DB-<time>/Summary.xlsx
	filePathSummary := filepath.Join(
		baseFolder,
		summaryFilename,
	)

	// Buat file summary
	f, errData := zipWriter.Create(filePathSummary)
	if errData != nil {
		return nil, "", errData
	}
	_, errData = f.Write(summaryByte)
	if errData != nil {
		return nil, "", errData
	}

	//grouping objects by owner
	for _, userObject := range userObjects {
		if _, ok := userObjectsMap[userObject.ObjectOwner]; !ok {
			userObjectsMap[userObject.ObjectOwner] = []model.OracleUserObject{}
		}
		userObjectsMap[userObject.ObjectOwner] = append(userObjectsMap[userObject.ObjectOwner], userObject)
	}

	for owner, objDBs := range userObjectsMap {
		allObjectSQL := new(strings.Builder)
		for _, userObject := range objDBs {
			if userObject.Ddl == "" {
				continue
			}
			userObject.Ddl = strings.ReplaceAll(userObject.Ddl, " EDITIONABLE ", " ")
			userObject.Ddl = strings.TrimSpace(userObject.Ddl)
			
			// Tampar-Object-DB-<time>/<ENV>/GENERATED/<OWNER>/<OBJECT TYPE>/<ObjectName>.sql
			filePath := filepath.Join(
				baseFolder,
				data.EnvSource,
				"GENERATED",
				userObject.ObjectOwner,
				userObject.ObjectType,
				userObject.ObjectName+".sql",
			)

			allObjectSQL.WriteString("\n"+userObject.Ddl + "\n")
			allObjectSQL.WriteString("/")

			// Buat file dalam zip
			f, errData := zipWriter.Create(filePath)
			if errData != nil {
				return nil, "", errData
			}
			_, errData = f.Write([]byte(userObject.Ddl))
			if errData != nil {
				return nil, "", errData
			}
		}
		//Tampar-Object-DB-<time>/<ENV>/GENERATED/<OWNER>/All-OBJECT-<OWNER>-<ENV>.sql
		allObjectFilePath := filepath.Join(
			baseFolder,
			data.EnvSource,
			"GENERATED",
			owner,
			fmt.Sprintf("ALL-OBJECT-%s-%s.sql",owner, data.EnvSource),
		)

		f, errData := zipWriter.Create(allObjectFilePath)
		if errData != nil {
			return nil, "", errData
		}

		_, errData = f.Write([]byte(allObjectSQL.String()))
		if errData != nil {
			return nil, "", errData
		}
	}

	// Tutup zip writer
	errData = zipWriter.Close()
	if errData != nil {
		return nil, "", errData
	}

	return buf.Bytes(), zipFileName, nil
}

func CreateSummaryCompare(userObjects []model.OracleUserObject, userObjectsExcel []model.OracleUserObject, data model.DataExcel)([]byte, error){
	f := excelize.NewFile()
	defer f.Close()
	
	makeSummaryExcel(f, userObjects, userObjectsExcel, data)

	byteBuff, errData := f.WriteToBuffer()
	if errData != nil {
		return nil, errData
	}

	return byteBuff.Bytes(), nil
}

func makeSummaryExcel(f *excelize.File, userObjects []model.OracleUserObject, userObjectsExcel []model.OracleUserObject, data model.DataExcel){
	existsColor := "FFF333"
	//Create sheet
	sheetTable := "TABLE"
	sheetView := "VIEW"
	sheetMv := "MATERIALIZED VIEW"
	sheetSeq := "SEQUENCE"
	sheetIndex := "INDEX"
	sheetType := "TYPE"
	sheetFunction := "FUNCTION"
	sheetProcedure := "PROCEDURE"
	sheetTrigger := "TRIGGER"

	f.SetSheetName(f.GetSheetName(0), sheetTable)
	f.NewSheet(sheetView)
	f.NewSheet(sheetMv)
	f.NewSheet(sheetSeq)
	f.NewSheet(sheetIndex)
	f.NewSheet(sheetType)
	f.NewSheet(sheetFunction)
	f.NewSheet(sheetProcedure)
	f.NewSheet(sheetTrigger)

	sheets := []string{
		sheetTable,
		sheetView,
		sheetMv,
		sheetSeq,
		sheetIndex,
		sheetType,
		sheetFunction,
		sheetProcedure,
		sheetTrigger,
	}

	//Style Header
	style, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Font: &excelize.Font{
			Bold:   true,
			Family: "Calibri",
			Size:   11,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"E0EBF5"}, Pattern: 1},
	})

	styleCell, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Font: &excelize.Font{
			Family: "Calibri",
			Size:   11,
		},
		Alignment: &excelize.Alignment{
			WrapText:    true,
			ShrinkToFit: true,
			Horizontal:  "left",
			Vertical:    "top",
		},
	})

	styleCellNew, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"03fc4e"}, Pattern: 1},
	})
	styleCellModNL, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"fcca03"}, Pattern: 1},
	})
	styleCellMissT, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"fc3503"}, Pattern: 1},
	})
	styleCellEqual, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"0390fc"}, Pattern: 1},
	})
	styleCellMissS, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"8403fc"}, Pattern: 1},
	})

	styleCellInvalid, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ff0000"}, Pattern: 1},
	})

	for i := range sheets {
		f.SetColWidth(sheets[i], "A", "A", 20)
		f.SetColWidth(sheets[i], "B", "B", 45)
		f.SetColWidth(sheets[i], "C", "C", 15)
		f.SetColWidth(sheets[i], "D", "D", 15)
		f.SetColWidth(sheets[i], "E", "E", 20)
		f.SetColWidth(sheets[i], "F", "F", 10)
		f.SetCellValue(sheets[i], "A1", "OWNER")
		f.SetCellValue(sheets[i], "B1", "OBJECT NAME")
		f.SetCellValue(sheets[i], "C1", "REMARK")
		f.SetCellValue(sheets[i], "D1", "PIC")
		f.SetCellValue(sheets[i], "E1", "NOTES")
		f.SetCellValue(sheets[i], "F1", "STATUS")
		f.SetCellStyle(sheets[i], "A1", "F1", style)
	}

	if data.OutputMode == "EXCEL" {
		for _, sheetName := range sheets {
			row := 2
			found := false
			for _, d := range userObjectsExcel {
				if strings.EqualFold(d.ObjectType, sheetName) {
					f.SetCellValue(sheetName, fmt.Sprintf("A%v", row), d.ObjectOwner)
					f.SetCellValue(sheetName, fmt.Sprintf("B%v", row), d.ObjectName)
					f.SetCellValue(sheetName, fmt.Sprintf("C%v", row), d.Remark)
					f.SetCellValue(sheetName, fmt.Sprintf("D%v", row), d.Pic)
					f.SetCellValue(sheetName, fmt.Sprintf("E%v", row), d.ObjectStatus)
					f.SetCellValue(sheetName, fmt.Sprintf("F%v", row), d.Status)
					if d.ObjectStatus == "NEW_LISTED" {
						f.SetCellStyle(sheetName, fmt.Sprintf("A%v", row), fmt.Sprintf("F%d", row), styleCellNew)
					} else if d.ObjectStatus == "MOD_NOT_LISTED" {
						f.SetCellStyle(sheetName, fmt.Sprintf("A%v", row), fmt.Sprintf("F%d", row), styleCellModNL)
					} else if d.ObjectStatus == "MISSING_TARGET" {
						f.SetCellStyle(sheetName, fmt.Sprintf("A%v", row), fmt.Sprintf("F%d", row), styleCellMissT)
					} else if d.ObjectStatus == "EQUALS" {
						f.SetCellStyle(sheetName, fmt.Sprintf("A%v", row), fmt.Sprintf("F%d", row), styleCellEqual)
					} else if d.ObjectStatus == "MISSING_SOURCE" {
						f.SetCellStyle(sheetName, fmt.Sprintf("A%v", row), fmt.Sprintf("F%d", row), styleCellMissS)
					} 
					
					if d.Status == "INVALID" {
						f.SetCellStyle(sheetName, fmt.Sprintf("F%v", row), fmt.Sprintf("F%d", row), styleCellInvalid)
					}
					found = true
					row++
				}
			}
			if found {
				f.SetCellStyle(sheetName, "A2", fmt.Sprintf("F%d", row), styleCell)
				_ = f.SetSheetProps(sheetName, &excelize.SheetPropsOptions{
						TabColorRGB: &existsColor,
					})
			}
		}
	} else {
		for _, sheetName := range sheets {
			row := 2
			found := false
			for _, d := range userObjects {
				if strings.EqualFold(d.ObjectType, sheetName) && d.ObjectStatus != "" {
					f.SetCellValue(sheetName, fmt.Sprintf("A%v", row), d.ObjectOwner)
					f.SetCellValue(sheetName, fmt.Sprintf("B%v", row), d.ObjectName)
					f.SetCellValue(sheetName, fmt.Sprintf("C%v", row), d.ObjectStatus)
					f.SetCellValue(sheetName, fmt.Sprintf("D%v", row), d.Status)
					found = true
					row++
				}
			}
			if found {
				f.SetCellStyle(sheetName, "A2", fmt.Sprintf("F%d", row), styleCell)
				_ = f.SetSheetProps(sheetName, &excelize.SheetPropsOptions{
						TabColorRGB: &existsColor,
					})
			}
		}
	}

	_ = f.SetDocProps(&excelize.DocProperties{
		Creator:        "Tampar System",
		Description:    "Database Comparation Template",
		Identifier:     "xlsx",
		Keywords:       "Tampar, Database, Comparation, Template",
		LastModifiedBy: "Tampar System",
		Revision:       "0",
		Subject:        "Tampar Database Comparation Template",
		Title:          "Tampar",
		Version:        "1.0.0",
	})
}

func GetSchema() []string {
	return []string{
		"1. GENERAL DEV",
		"2. GENERAL PRODLIKE",
		"3. GENERAL QA",
		"4. GENERAL UAT",
		"5. MSTITEM DEV",
		"6. MSTITEM PRODLIKE",
		"7. MSTITEM QA",
		"8. MSTITEM UAT",
		"9. MSTVENDOR DEV",
		"10. MSTVENDOR PRODLIKE",
		"11. MSTVENDOR QA",
		"12. MSTVENDOR UAT",
		"13. PORTALUSER DEV",
		"14. PORTALUSER PRODLIKE",
		"15. PORTALUSER QA",
		"16. PORTALUSER UAT",
		"17. POUSER DEV",
		"18. POUSER PRODLIKE",
		"19. POUSER QA",
		"20. POUSER UAT",
		"21. PRUSER DEV",
		"22. PRUSER PRODLIKE",
		"23. PRUSER QA",
		"24. PRUSER UAT",
		"25. RFIUSER DEV",
		"26. RFIUSER PRODLIKE",
		"27. RFIUSER QA",
		"28. RFIUSER UAT",
	}
}

func NormalizeDdl(s string) string {
	// Hapus semua comment --
	re := regexp.MustCompile(`--.*`)
	s = re.ReplaceAllString(s, "")

	// Hapus semua blok comment /* ... */ kecuali hint /*+ ... */
	var result strings.Builder
	i := 0
	for i < len(s) {
		if i+2 < len(s) && s[i] == '/' && (s[i+1] == '*' || (s[i+1] == '\n' && s[i+2] == '*')) {
			// Jika hint /*+ ... */
			if i+2 < len(s) && s[i+2] == '+' {
				// Cari akhir blok */
				end := strings.Index(s[i:], "*/")
				if end == -1 {
					result.WriteString(s[i:])
					break
				}
				result.WriteString(s[i : i+end+2])
				i += end + 2
			} else {
				// Bukan hint, hapus blok comment
                end := strings.Index(s[i:], "*/")
                end1 := strings.Index(s[i:], "*\n/")
                if end == -1 && end1 == -1 {
                    i = len(s)
                } else {
                  if end >= 0 && (end <= end1 || end1 == -1) {
                     i += end + 2
                  } else {
                     i += end1 + 3
                  }
                }
			}
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	s = result.String()

	// Hapus whitespace di akhir string
	s = strings.TrimSpace(s)

	// Hapus semua karakter "
	s = strings.ReplaceAll(s, `"`, "")

	// Hapus / di awal dan akhir string
	s = strings.Trim(s, "/")

	// Hapus semua spasi, enter, tab, newline
	re = regexp.MustCompile(`[\s\t\r\n]+`)
	s = re.ReplaceAllString(s, "")
	return strings.ToLower(s)
}

func FilterException(s, exc []model.OracleUserObject) []model.OracleUserObject {
	var (
		found  bool
		filter []model.OracleUserObject
	)
	if len(exc) > 0 {
		for _, v := range s {
			found = false
			for _, e := range exc {
				if strings.EqualFold(v.ObjectOwner, e.ObjectOwner) && strings.EqualFold(v.ObjectType, e.ObjectType) && strings.EqualFold(v.ObjectName, e.ObjectName) {
					found = true
					break
				}
			}

			if !found {
				filter = append(filter, v)
			}
		}
	} else {
		return s
	}
	return DistinctObjectDB(filter)
}

func CompareObjectDb(listObjectDbAll []model.OracleUserObject, listObjectDbExcel []model.OracleUserObject, 
	listExclude []model.OracleUserObject, data model.DataExcel)([]model.OracleUserObject, []model.OracleUserObject){
	var (
		listObjFromAll,listObjEnvSrcAll, listObjEnvTrgAll,listObjEnvSrc, listObjEnvTrg []model.OracleUserObject
		found,similar   bool
	)

	listObjEnvSrc, listObjEnvTrg = GetObjectByEnv(listObjectDbExcel, data.EnvSource, data.EnvTarget)
	OrderObjDb(listObjEnvSrc)
	OrderObjDb(listObjEnvTrg)

	// Compare Object DB from Excel
	for i, src := range listObjEnvSrc {
		found = false
		similar = false
		if src.Ddl == "" {
			listObjEnvSrc[i].ObjectStatus = "MISSING_SOURCE" //Object tidak ada DDL atau disource namun ada di list
			continue
		} else {
			for _, trg := range listObjEnvTrg {
				if strings.EqualFold(src.ObjectOwner, trg.ObjectOwner) && strings.EqualFold(src.ObjectType, trg.ObjectType) && 
					strings.EqualFold(src.ObjectName, trg.ObjectName) {

					found = true
					if(NormalizeDdl(src.Ddl) == NormalizeDdl(trg.Ddl)) {
						similar = true
					}
					break
				}
			}
			
			if !found {
				listObjEnvSrc[i].ObjectStatus = "NEW_LISTED"//Object baru, ada di list
			} else if found && !similar {
				listObjEnvSrc[i].ObjectStatus = "MOD_LISTED"//Object berbeda, ada di list
			} else if found && similar{
				listObjEnvSrc[i].ObjectStatus = "EQUALS"//Ada dilist tapi Sama dengan Env Target
			}
		}
	}


	if data.OutputMode == "FULL" {
		listObjEnvSrcAll, listObjEnvTrgAll = GetObjectByEnv(listObjectDbAll, data.EnvSource, data.EnvTarget)
		OrderObjDb(listObjEnvSrcAll)
		OrderObjDb(listObjEnvTrgAll)
	
		//Compare Object DB from All
		for i, src := range listObjEnvSrcAll {
			found = false
			similar = false
			
			for _, trg := range listObjEnvTrgAll {
				if strings.EqualFold(src.ObjectOwner, trg.ObjectOwner) && strings.EqualFold(src.ObjectType, trg.ObjectType) && 
					strings.EqualFold(src.ObjectName, trg.ObjectName) {

					found = true
					if(NormalizeDdl(src.Ddl) == NormalizeDdl(trg.Ddl)) {
						similar = true
					}
					break
				}
			}
			
			if !found {
				listObjEnvSrcAll[i].ObjectStatus = "MISSING_TARGET"//Object baru, ada di list
			} else if found && !similar {
				listObjEnvSrcAll[i].ObjectStatus = "MOD_NOT_LISTED"//Object berbeda, ada di list
			}
		}

		for _, v := range listObjEnvSrcAll {
			found = false
			for _, e := range listExclude {
				if strings.EqualFold(v.ObjectOwner, e.ObjectOwner) && strings.EqualFold(v.ObjectType, e.ObjectType) && 
					strings.EqualFold(v.ObjectName, e.ObjectName) {		
					found = true
					break
				}
			}
			if !found && (v.ObjectStatus != "" || v.Status == "INVALID") {
				listObjFromAll = append(listObjFromAll, v)
			}
		}
	}

	//
	for _, v := range listObjEnvSrcAll {
		found = false
		for _, e := range listExclude {
			if strings.EqualFold(v.ObjectOwner, e.ObjectOwner) && strings.EqualFold(v.ObjectType, e.ObjectType) && 
				strings.EqualFold(v.ObjectName, e.ObjectName) {		
				found = true
				break
			}
		}
		if !found && (v.ObjectStatus != "" || v.Status == "INVALID") {
			listObjFromAll = append(listObjFromAll, v)
		}
	}

	return listObjFromAll, listObjEnvSrc

}

func ObjSeq(objType string)int{
	var seq int
	switch objType {
	case ORACLE_TYPE_TABLE:
		seq = 1
	case ORACLE_TYPE_VIEW:
		seq = 2
	case ORACLE_TYPE_MV:
		seq = 3
	case ORACLE_TYPE_SEQUENCE:
		seq = 4
	case ORACLE_TYPE_INDEX:
		seq = 5
	case ORACLE_TYPE_OBJECT:
		seq = 6
	case ORACLE_TYPE_FUNCTION:
		seq = 7
	case ORACLE_TYPE_PROCEDURE:
		seq = 8
	case ORACLE_TYPE_TRIGGER:
		seq = 9
	default:
		seq = 10
	}
	return seq
}
type RestBody struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
	Code    int    `json:"code"`
}

func SuccessBody(data any, message ...string) (int, RestBody) {
	result := RestBody{}
	result.Data = data
	if message != nil && len(message) == 1 {
		result.Message = message[0]
	} else {
		result.Message = "Success"
	}
	result.Code = http.StatusOK
	return result.Code, result
}

func ErrorBody(err error, code ...int) (int, RestBody) {

	result := RestBody{}
	result.Data = nil
	if len(code) == 1 {
		result.Code = code[0]
	} else {
		result.Code = http.StatusInternalServerError
	}
	result.Message = err.Error()

	return result.Code, result
}

func MakeTemplateExcel(f *excelize.File){
	//Create sheet
	sheetTable := "TABLE"
	sheetView := "VIEW"
	sheetMv := "MATERIALIZED VIEW"
	sheetSeq := "SEQUENCE"
	sheetIndex := "INDEX"
	sheetPackage := "PACKAGE"
	sheetType := "TYPE"
	sheetFunction := "FUNCTION"
	sheetProcedure := "PROCEDURE"
	sheetException := "EXCEPTION"
	f.SetSheetName(f.GetSheetName(0), sheetTable)
	f.NewSheet(sheetView)
	f.NewSheet(sheetMv)
	f.NewSheet(sheetSeq)
	f.NewSheet(sheetIndex)
	f.NewSheet(sheetPackage)
	f.NewSheet(sheetType)
	f.NewSheet(sheetFunction)
	f.NewSheet(sheetProcedure)
	f.NewSheet(sheetException)

	sheets := []string{
		sheetView,
		sheetMv,
		sheetSeq,
		sheetIndex,
		sheetPackage,
		sheetType,
		sheetFunction,
		sheetProcedure,
		sheetException,
	}

	//Style Header
	style, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Font: &excelize.Font{
			Bold:   true,
			Family: "Calibri",
			Size:   9,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"E0EBF5"}, Pattern: 1},
	})

	for i, _ := range sheets {
		
		f.SetColWidth(sheets[i], "A", "A", 40)
		f.SetColWidth(sheets[i], "B", "B", 50)
		f.SetColWidth(sheets[i], "C", "D", 40)
		f.SetCellValue(sheets[i], "A1", "OWNER")
		f.SetCellValue(sheets[i], "B1", "OBJECT NAME")
		if(sheets[i] == sheetException){
			f.SetCellValue(sheets[i], "C1", "OBJECT TYPE")
		} else {
			f.SetCellValue(sheets[i], "C1", "REMARK")
		}
		f.SetCellValue(sheets[i], "D1", "PIC")
		f.SetCellStyle(sheets[i], "A1", "D1", style)
	}
}