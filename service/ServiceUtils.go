package service

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	ORACLE_TYPE_LOB        = "LOB"
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

func NewOracleDatabase(config model.OracleDbConfig) model.Database {
	database, errData := sql.Open(config.DbName, config.DbName+"://"+config.DbUsername+":"+config.DbPassword+"@"+config.DbUrl+":"+config.DbPort+"/"+config.DbSid)
	if errData != nil {
		log.Println(errData)
	}
	return model.Database{Database: database, Schema: config.DbUsername, Enviroment: config.DbEnv}
}

func GetTableColumnsInfo(conn *model.Database, tableName string) []model.OracleDbColumn {
	var (
		columnResults []model.OracleDbColumn
		columnResult  model.OracleDbColumn
	)
	rows, errData := conn.Database.Query("SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, COLUMN_ID FROM USER_TAB_COLUMNS WHERE TABLE_NAME = '" + tableName + "' AND DATA_TYPE_OWNER IS NULL")
	if errData != nil {
		log.Println(errData)
	}
	defer rows.Close()

	for rows.Next() {
		errData = rows.Scan(&columnResult.TableName, &columnResult.Name, &columnResult.Type, &columnResult.Sequence)
		if errData != nil {
			log.Println(errData)
		}
		columnResults = append(columnResults, columnResult)
	}
	return columnResults
}

func GetDdl(conn model.Database, userObjects *[]model.OracleUserObject, schema string) {

	_, errData := conn.Database.Exec("BEGIN DBMS_METADATA.SET_TRANSFORM_PARAM(DBMS_METADATA.SESSION_TRANSFORM,'STORAGE',FALSE); END;")
	if errData != nil {
		log.Println(errData)
	}

	_, errData = conn.Database.Exec("BEGIN DBMS_METADATA.SET_TRANSFORM_PARAM(DBMS_METADATA.SESSION_TRANSFORM,'SEGMENT_ATTRIBUTES',FALSE); END;")
	if errData != nil {
		log.Println(errData)
	}

	_, errData = conn.Database.Exec("BEGIN DBMS_METADATA.SET_TRANSFORM_PARAM(DBMS_METADATA.SESSION_TRANSFORM, 'EMIT_SCHEMA', FALSE); END;")
	if errData != nil {
		log.Println(errData)
	}

	_, errData = conn.Database.Exec("BEGIN DBMS_METADATA.SET_TRANSFORM_PARAM(DBMS_METADATA.SESSION_TRANSFORM,'SQLTERMINATOR', TRUE); END;")
	if errData != nil {
		log.Println(errData)
	}

	sort.Slice(*userObjects, func(i, j int) bool {
		return (*userObjects)[i].ObjectType < (*userObjects)[j].ObjectType
	})

	for i, userObject := range *userObjects {
		if strings.TrimSpace(userObject.ObjectOwner) != "" && !strings.EqualFold(strings.TrimSpace(userObject.ObjectOwner), schema) {
			continue
		}

		fmt.Println("Get DDL from " + userObject.ObjectOwner + " for " + userObject.ObjectName + " of type " + userObject.ObjectType)
		rows, errData := conn.Database.Query("SELECT DBMS_METADATA.GET_DDL('" + strings.TrimSpace(userObject.ObjectType) + "', '" + strings.TrimSpace(userObject.ObjectName) + "', '" + strings.TrimSpace(schema) + "') FROM DUAL")
		if errData != nil {
			log.Println(errData)
			continue
		}

		defer rows.Close()
		for rows.Next() {
			errData = rows.Scan(&(*userObjects)[i].Ddl)

			if errData != nil {
				log.Println(errData)
				continue
			}
		}
	}
}

// GetUserObjects retrieves user objects from the database and returns them as a slice of OracleUserObject.
func GetUserObjects(conn *model.Database) []model.OracleUserObject {
	var (
		results []model.OracleUserObject
		result  model.OracleUserObject
	)
	rows, errData := conn.Database.Query("SELECT OBJECT_NAME, OBJECT_ID, OBJECT_TYPE FROM USER_OBJECTS ORDER BY CASE WHEN OBJECT_TYPE = 'TYPE' THEN 1 WHEN OBJECT_TYPE = 'SEQUENCE' THEN 2 WHEN OBJECT_TYPE = 'PACKAGE' THEN 3 WHEN OBJECT_TYPE = 'TABLE' THEN 4 WHEN OBJECT_TYPE = 'VIEW' THEN 5 WHEN OBJECT_TYPE = 'INDEX' THEN 6 WHEN OBJECT_TYPE = 'FUNCTION' THEN 7 WHEN OBJECT_TYPE = 'PROCEDURE' THEN 8 ELSE 9 END ASC, (select case when count(1) > 0 then 1 else 0 end from all_dependencies where name = OBJECT_NAME and REFERENCED_TYPE = 'TYPE') ASC, OBJECT_NAME ASC")
	if errData != nil {
		log.Println(errData)
	}

	defer rows.Close()
	for rows.Next() {
		errData = rows.Scan(&result.ObjectName, &result.ObjectId, &result.ObjectType)
		if errData != nil {
			log.Println(errData)
			continue
		}
		results = append(results, result)
	}
	return results
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
			GetDdl(db, &listObjectDbs, db.Schema)
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
					listObjectDbResult[j].IsListed = listObjectExcel[i].IsListed
					break
				}
			}
		}
	}

	return DistinctObjectDB(listObjectDbResult)
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
			`SELECT DISTINCT OBJECT_NAME, OBJECT_TYPE, SEQ FROM (
			SELECT OBJECT_NAME, OBJECT_TYPE,
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
	}

	defer rows.Close()
	for rows.Next() {
		errData = rows.Scan(&result.ObjectName, &result.ObjectType)
		if errData != nil {
			log.Println(errData)
			continue
		}
		result.ObjectOwner = conn.Schema
		result.ObjectEnv = conn.Enviroment
		results = append(results, result)
	}
	return results
}

// Create txt file from user objects list
// Dir uses "/" or "\\" suffix
func CreateTxtFromUserObjects(userObjects []model.OracleUserObject, path string) {
	userObjectsMap := make(map[string][]model.OracleUserObject)
	for _, userObject := range userObjects {

		if _, ok := userObjectsMap[userObject.ObjectType]; !ok {
			userObjectsMap[userObject.ObjectType] = []model.OracleUserObject{}
		}

		userObjectsMap[userObject.ObjectType] = append(userObjectsMap[userObject.ObjectType], userObject)
	}

	// loop per type
	for k, v := range userObjectsMap {
		var stringByte []byte
		var kIndex string
		for _, userObject := range v {
			if userObject.Ddl == "" {
				continue
			}
			stringByte = append(stringByte, strings.TrimSpace(userObject.Ddl)...)
		}

		switch k {
		case "TYPE":
			kIndex = "1"
		case "SEQUENCE":
			kIndex = "2"
		case "PACKAGE":
			kIndex = "3"
		case "TABLE":
			kIndex = "4"
		case "VIEW":
			kIndex = "5"
		case "INDEX":
			kIndex = "6"
		case "FUNCTION":
			kIndex = "7"
		case "PROCEDURE":
			kIndex = "8"
		case "JOB":
			kIndex = "9"
		default:
			kIndex = "10"
		}

		if path == "" {
			filePath := filepath.Join(kIndex + ". " + k + ".txt")
			fmt.Println("Writing " + filePath)
			errData := os.WriteFile(filePath, stringByte, 0644)
			if errData != nil {
				log.Println(errData)
			}
		} else {
			filePath := filepath.Dir(path)
			filePath = filepath.Join(filePath, kIndex+". "+k+".txt")
			os.MkdirAll(filepath.Dir(path), os.ModePerm)
			fmt.Println("Writing " + filePath)
			errData := os.WriteFile(filePath, stringByte, 0644)
			if errData != nil {
				log.Println(errData)
			}
		}
	}
}

func OrderObjDb(userObjects []model.OracleUserObject){
	sort.Slice(userObjects, func(i, j int) bool {
		if userObjects[i].ObjectOwner != userObjects[j].ObjectOwner {
			return userObjects[i].ObjectOwner < userObjects[j].ObjectOwner
		}

		if userObjects[i].ObjectSeq != userObjects[j].ObjectSeq {
			//return ObjSeq(userObjects[i].ObjectType) < ObjSeq(userObjects[j].ObjectType)
			return userObjects[i].ObjectSeq < userObjects[j].ObjectSeq
		}
		return userObjects[i].ObjectName < userObjects[j].ObjectName
	})
}

func CreateFileObjectDB(userObjects []model.OracleUserObject, env string)([]byte, string, error) {
	userObjectsMap := make(map[string][]model.OracleUserObject)

	timestamp := time.Now().Format("20060102150405")
	baseFolder := fmt.Sprintf("Tampar-Object-DB-%s-%s",env, timestamp)
	zipFileName := baseFolder + ".zip"

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	sort.Slice(userObjects, func(i, j int) bool {
		if userObjects[i].ObjectOwner != userObjects[j].ObjectOwner {
			return userObjects[i].ObjectOwner < userObjects[j].ObjectOwner
		}

		if userObjects[i].ObjectSeq != userObjects[j].ObjectSeq {
			//return ObjSeq(userObjects[i].ObjectType) < ObjSeq(userObjects[j].ObjectType)
			return userObjects[i].ObjectSeq < userObjects[j].ObjectSeq
		}
		return userObjects[i].ObjectName < userObjects[j].ObjectName
	})

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

			allObjectSQL.WriteString("--== START " + userObject.ObjectType + "-" + userObject.ObjectName +" ==--\n")
			allObjectSQL.WriteString(userObject.Ddl + "\n")
			allObjectSQL.WriteString("--== END " + userObject.ObjectType + "-" + userObject.ObjectName +" ==--\n\n")

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

func CreateFileObjectDBCompare(userObjects []model.OracleUserObject, data model.DataExcel)([]byte, string, error) {
	userObjectsMap := make(map[string][]model.OracleUserObject)

	timestamp := time.Now().Format("20060102150405")
	baseFolder := fmt.Sprintf("Tampar-Object-DB-%s-%s-%s",data.EnvSource, data.EnvTarget, timestamp)
	zipFileName := baseFolder + ".zip"

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	sort.Slice(userObjects, func(i, j int) bool {
		if userObjects[i].ObjectOwner != userObjects[j].ObjectOwner {
			return userObjects[i].ObjectOwner < userObjects[j].ObjectOwner
		}

		if userObjects[i].ObjectSeq != userObjects[j].ObjectSeq {
			//return ObjSeq(userObjects[i].ObjectType) < ObjSeq(userObjects[j].ObjectType)
			return userObjects[i].ObjectSeq < userObjects[j].ObjectSeq
		}
		return userObjects[i].ObjectName < userObjects[j].ObjectName
	})

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

			allObjectSQL.WriteString("--== START " + userObject.ObjectType + "-" + userObject.ObjectName +" ==--\n")
			allObjectSQL.WriteString(userObject.Ddl + "\n")
			allObjectSQL.WriteString("--== END " + userObject.ObjectType + "-" + userObject.ObjectName +" ==--\n\n")

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
	errData := zipWriter.Close()
	if errData != nil {
		return nil, "", errData
	}

	return buf.Bytes(), zipFileName, nil
}

func GetOraSource(listSchema []string, env string) []model.Database {
	var (
		OraSourceDbList []model.Database
	)
	return OraSourceDbList
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
		if i+2 < len(s) && s[i] == '/' && s[i+1] == '*' {
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
				if end == -1 {
					i = len(s)
				} else {
					i += end + 2
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

func cleanup(dbaList []model.Database) {
	for _, dba := range dbaList {
		dba.Database.Close()
	}
}

// func main() {
// 	var OraSourceDbList []model.Database
// 	schemas := GetSchema()
// 	// ============= CTRL + C =============
// 	c := make(chan os.Signal)
// 	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
// 	go func() {
// 		<-c
// 		cleanup(OraSourceDbList)
// 		os.Exit(1)
// 	}()

// 	var modeGet string
// 	var numberGet int
// 	var env string
// 	var number int
// 	var dateFrom string

// 	listSchema := make([]string, 0)
// 	listObjExcel := make([]model.OracleUserObject, 0)
// 	listObjDdl := make([]model.OracleUserObject, 0)
// 	//----	IMPORTANT !!! ---------------
// pilihNomor:
// 	for _, s := range schemas {
// 		fmt.Println(s)
// 	}
// 	fmt.Println("---------------")
// 	fmt.Println("99. BY EXCEL")
// 	//fmt.Println("100. DEFINE SCHEMA")
// 	fmt.Println("-----------------------------------------------------------")
// 	fmt.Println("Chosee Schema (number): ")

// 	_, err := fmt.Scanf("%d\n", &number)
// 	if err != nil || ((number-1) > len(schemas) && number != MODE_GET_OBJ_BY_EXCEL && number != MODE_GET_OBJ_DEFINE_SCHEMA) || number < 1 {
// 		fmt.Println("Number Valid")
// 		goto pilihNomor
// 	} else if number == MODE_GET_OBJ_BY_EXCEL {
// 	pilihEnv:
// 		fmt.Println("-----------------------------------------------------------")
// 		fmt.Println("Type Enviroment (DEV/QA/UAT/PRODLIKE): ")
// 		_, err := fmt.Scanln(&env)
// 		if err != nil {
// 			fmt.Println("Option not Valid")
// 			goto pilihEnv
// 		} else if !strings.EqualFold(env, "DEV") && !strings.EqualFold(env, "QA") &&
// 			!strings.EqualFold(env, "UAT") && !strings.EqualFold(env, "PRODLIKE") {

// 			fmt.Println("Enviroment not Valid")
// 			goto pilihEnv
// 		}
// 		env = strings.ToUpper(env)
// 		goto loopDB
// 	}

// pilihMode:
// 	fmt.Println("----------------------------MODE-------------------------------")
// 	fmt.Println("1. From Excel")
// 	fmt.Println("2. Last Compile")
// 	fmt.Println("3. From Excel & Last Compile")
// 	fmt.Println("4. All since The Big Bang")
// 	fmt.Println("Chosee Mode: ")
// 	_, err = fmt.Scanf("%d\n", &numberGet)
// 	if err != nil || numberGet > 4 || numberGet < 1 {
// 		fmt.Println("Number Valid")
// 		goto pilihMode
// 	} else if numberGet == 1 {
// 		modeGet = MODE_GET_OBJECT_EXCEL
// 	} else if numberGet == 2 {
// 		modeGet = MODE_GET_OBJECT_LAST_COMPILE
// 	} else if numberGet == 3 {
// 		modeGet = MODE_GET_OBJECT_COMBINE
// 	} else {
// 		modeGet = MODE_GET_OBJECT_ALL
// 		dateFrom = "2024-01-01"
// 	}

// pilihDate:
// 	if numberGet == 2 || numberGet == 3 {
// 		fmt.Println("Input Date From to get Object with format YYYY-MM-DD (example: 2024-06-25): ")
// 		n, err := fmt.Scanln(&dateFrom)

// 		if err != nil || n > 1 {
// 			fmt.Println("Date not Valid")
// 			goto pilihDate
// 		} else {
// 			_, err := time.Parse("2006-01-02", dateFrom)
// 			if err != nil {
// 				fmt.Println("Date not Valid")
// 				goto pilihDate
// 			}
// 		}
// 	}

// loopDB:
// 	if number == MODE_GET_OBJ_BY_EXCEL {
// 		timeFormat := (time.Now()).Format("20060102150405")
// 		_, _, _ = GetObjectFromExcel(&listObjExcel, "", "", timeFormat)
// 		listSchema = GetOwnerEnv(listObjExcel)
// 		OraSourceDbList = GetOraSource(listSchema, env)
// 	} else {
// 		schema, env := GetDataSchema(number)
// 		OraSourceDbList = GetOraSource([]string{schema}, env)
// 	}
// 	for _, db := range OraSourceDbList {
// 		if number != MODE_GET_OBJ_BY_EXCEL {
// 			fmt.Println("|-----------------------------------------------------------|")
// 			fmt.Println("Get object of " + db.Schema + " enviroment " + db.Enviroment + " from date: " + dateFrom + " until Now")
// 			fmt.Println("|-----------------------------------------------------------|")

// 			//Get and write to excel
// 			userObjects, version := GetUserObjectWriteToExcel(&db, dateFrom, modeGet)

// 			// loop user objects and get DDL
// 			if len(userObjects) > 0 || version != "" {
// 				GetDdl(&db, &userObjects, db.Schema)
// 				path := GENERAL_PATH + db.Schema + "\\" + db.Enviroment + "\\Generated\\" + version + "\\"
// 				EnsurePath(path, MODE_DIR_PATH)
// 				CreateSeparateTxtFromUserObjects(userObjects, path, version, "0")
// 			} else {
// 				fmt.Println("No Object Updated")
// 			}
// 		} else if number == MODE_GET_OBJ_BY_EXCEL && db.Enviroment == strings.ToUpper(env) && len(listObjExcel) > 0 {
// 			listData := GetObjectDBbyUser(db.Schema, listObjExcel)
// 			if len(listData) > 0 {
// 				GetDdl(&db, &listData, db.Schema)
// 				listObjDdl = append(listObjDdl, listData...)
// 			}
// 		}

// 		db.Database.Close()
// 	}

// 	if number == MODE_GET_OBJ_BY_EXCEL {
// 		pathArchieve := GENERAL_PATH + "ALL OBJECT\\" + strings.ToUpper(env)
// 		EnsurePath(pathArchieve, MODE_DIR_PATH)
// 		var numFile int
// 		files, errData := os.ReadDir(pathArchieve)
// 		if errData != nil {
// 			numFile = 0
// 		} else {
// 			numFile = len(files)
// 		}
// 		version := fmt.Sprintf("V%d", numFile+1)
// 		path := GENERAL_PATH + "ALL OBJECT\\" + strings.ToUpper(env) + "\\" + version + "\\"
// 		EnsurePath(path, MODE_DIR_PATH)
// 		CreateSeparateTxtFromUserObjects(listObjDdl, path, version, "0")
// 	}
// }

func GetOwnerEnv(dta []model.OracleUserObject) []string {
	schemas := make([]string, 0)
	m := map[string]bool{}
	for _, v := range dta {
		if !m[v.ObjectOwner] {
			m[v.ObjectOwner] = true
			schemas = append(schemas, v.ObjectOwner)
		}
	}
	return schemas
}

func DefineLastUpdate(path string) string {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "2024-03-01"
	}
	modTime := fileInfo.ModTime()
	return modTime.Format("2006-01-02")
}

// EnsurePath ensures the directory or file exists based on the mode.
func EnsurePath(path, mode string) error {
	switch mode {
	case MODE_DIR_PATH:
		// Check if the directory exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Create the directory
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return err
			}
			log.Println("Directory created:", path)
		}
	case MODE_DIR_FILE:
		// Get the directory of the file
		dir := filepath.Dir(path)
		// Check if the directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			// Create the directory
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}
			log.Println("Directory created:", dir)
		}
		// Check if the file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Copy the template file to the destination
			templatePath := GENERAL_PATH + "Template\\Template Object DB.xlsx"
			if err := CopyFile(templatePath, path); err != nil {
				return err
			}
			log.Println("File copied to:", path)
		}
	default:
		return fmt.Errorf("invalid mode: %s", mode)
	}
	return nil
}

// CopyFile copies a file from src to dst. If dst does not exist, it will be created.
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}
	return destinationFile.Sync()
}

// func GetUserObjectWriteToExcel(db *model.Database, dateFrom string, modeGet string) ([]model.OracleUserObject, string) {
// 	var (
// 		distinctObjects []model.OracleUserObject = []model.OracleUserObject{}
// 		version         string
// 	)
// 	if modeGet == MODE_GET_OBJECT_LAST_COMPILE || modeGet == MODE_GET_OBJECT_COMBINE || modeGet == MODE_GET_OBJECT_ALL {
// 		userObjects := GetUserObjects(db)
// 		if len(userObjects) == 0 {
// 			fmt.Println("No update object from DB")
// 		}
// 		distinctObjects = DistinctSlice(userObjects)
// 	}

// 	if len(distinctObjects) > 0 {
// 		filePath := GENERAL_PATH + db.Schema + "\\" + db.Enviroment + "\\" + db.Schema + " Object DB.xlsx"
// 		EnsurePath(filePath, MODE_DIR_FILE)
// 		f, err := excelize.OpenFile(filePath)
// 		if err != nil {
// 			fmt.Println(err)
// 			return distinctObjects, ""
// 		}

// 		rowsIndex, _ := f.GetRows(ORACLE_TYPE_INDEX)
// 		lastRowIndex := len(rowsIndex)
// 		rowsSeq, _ := f.GetRows(ORACLE_TYPE_SEQUENCE)
// 		lastRowSeq := len(rowsSeq)
// 		rowsType, _ := f.GetRows(ORACLE_TYPE_OBJECT)
// 		lastRowType := len(rowsType)
// 		rowsFunction, _ := f.GetRows(ORACLE_TYPE_FUNCTION)
// 		lastRowFunction := len(rowsFunction)
// 		rowsProcedure, _ := f.GetRows(ORACLE_TYPE_PROCEDURE)
// 		lastRowProcedure := len(rowsProcedure)

// 		idxType := lastRowType
// 		idxFunction := lastRowFunction
// 		idxProcedure := lastRowProcedure
// 		idxIndex := lastRowIndex
// 		idxSeq := lastRowSeq
// 		existsColor := "FFF333"
// 		remarks := "MODIFIED"
// 		userModified := "SYSTEM"
// 		project := "EPROC-"
// 		//Write Object DB to file
// 		for _, row := range distinctObjects {
// 			if row.ObjectType == ORACLE_TYPE_OBJECT {
// 				if lastRowType > 0 && !isExists(rowsType[0:], row.ObjectName) {
// 					idxType = idxType + 1
// 					cell, err := excelize.CoordinatesToCellName(2, idxType)
// 					if err != nil {
// 						log.Fatalf("Failed to get cell name: %v", err)
// 					}
// 					//Set At once
// 					if idxType == 2 || idxType == 3 {
// 						if err := f.SetSheetProps(ORACLE_TYPE_OBJECT, &excelize.SheetPropsOptions{
// 							TabColorRGB: &existsColor,
// 						}); err != nil {
// 							fmt.Println(err)
// 						}
// 					}
// 					f.SetCellValue(ORACLE_TYPE_OBJECT, fmt.Sprintf("A%d", idxType), db.Schema)
// 					f.SetCellValue(ORACLE_TYPE_OBJECT, fmt.Sprintf("C%d", idxType), remarks)
// 					f.SetCellValue(ORACLE_TYPE_OBJECT, fmt.Sprintf("D%d", idxType), userModified)
// 					f.SetCellValue(ORACLE_TYPE_OBJECT, fmt.Sprintf("E%d", idxType), project+db.Schema)
// 					f.SetCellValue(ORACLE_TYPE_OBJECT, cell, row.ObjectName)
// 				}
// 				if lastRowType > 0 && len(rowsType) > 1 {
// 					if err := f.SetSheetProps(ORACLE_TYPE_OBJECT, &excelize.SheetPropsOptions{
// 						TabColorRGB: &existsColor,
// 					}); err != nil {
// 						fmt.Println(err)
// 					}
// 				}
// 			} else if row.ObjectType == ORACLE_TYPE_FUNCTION {
// 				if lastRowFunction > 0 && !isExists(rowsFunction[0:], row.ObjectName) {
// 					idxFunction = idxFunction + 1
// 					cell, err := excelize.CoordinatesToCellName(2, idxFunction)
// 					if err != nil {
// 						log.Fatalf("Failed to get cell name: %v", err)
// 					}
// 					//Set At once
// 					if idxFunction == 2 || idxFunction == 3 {
// 						if err := f.SetSheetProps(ORACLE_TYPE_FUNCTION, &excelize.SheetPropsOptions{
// 							TabColorRGB: &existsColor,
// 						}); err != nil {
// 							fmt.Println(err)
// 						}
// 					}
// 					f.SetCellValue(ORACLE_TYPE_FUNCTION, fmt.Sprintf("A%d", idxFunction), db.Schema)
// 					f.SetCellValue(ORACLE_TYPE_FUNCTION, fmt.Sprintf("C%d", idxFunction), remarks)
// 					f.SetCellValue(ORACLE_TYPE_FUNCTION, fmt.Sprintf("D%d", idxFunction), userModified)
// 					f.SetCellValue(ORACLE_TYPE_FUNCTION, fmt.Sprintf("E%d", idxFunction), project+db.Schema)
// 					f.SetCellValue(ORACLE_TYPE_FUNCTION, cell, row.ObjectName)
// 				}
// 				if lastRowFunction > 0 && len(rowsFunction) > 1 {
// 					if err := f.SetSheetProps(ORACLE_TYPE_FUNCTION, &excelize.SheetPropsOptions{
// 						TabColorRGB: &existsColor,
// 					}); err != nil {
// 						fmt.Println(err)
// 					}
// 				}
// 			} else if row.ObjectType == ORACLE_TYPE_PROCEDURE {
// 				if lastRowProcedure > 0 && !isExists(rowsProcedure[0:], row.ObjectName) {
// 					idxProcedure = idxProcedure + 1
// 					cell, err := excelize.CoordinatesToCellName(2, idxProcedure)
// 					if err != nil {
// 						log.Fatalf("Failed to get cell name: %v", err)
// 					}
// 					//Set At once
// 					if idxProcedure == 2 || idxProcedure == 3 {
// 						if err := f.SetSheetProps(ORACLE_TYPE_PROCEDURE, &excelize.SheetPropsOptions{
// 							TabColorRGB: &existsColor,
// 						}); err != nil {
// 							fmt.Println(err)
// 						}
// 					}
// 					f.SetCellValue(ORACLE_TYPE_PROCEDURE, fmt.Sprintf("A%d", idxProcedure), db.Schema)
// 					f.SetCellValue(ORACLE_TYPE_PROCEDURE, fmt.Sprintf("C%d", idxProcedure), remarks)
// 					f.SetCellValue(ORACLE_TYPE_PROCEDURE, fmt.Sprintf("D%d", idxProcedure), userModified)
// 					f.SetCellValue(ORACLE_TYPE_PROCEDURE, fmt.Sprintf("E%d", idxProcedure), project+db.Schema)
// 					f.SetCellValue(ORACLE_TYPE_PROCEDURE, cell, row.ObjectName)
// 				}
// 				if lastRowProcedure > 0 && len(rowsProcedure) > 1 {
// 					if err := f.SetSheetProps(ORACLE_TYPE_PROCEDURE, &excelize.SheetPropsOptions{
// 						TabColorRGB: &existsColor,
// 					}); err != nil {
// 						fmt.Println(err)
// 					}
// 				}
// 			} else if row.ObjectType == ORACLE_TYPE_INDEX {
// 				if lastRowIndex > 0 && !isExists(rowsIndex[0:], row.ObjectName) {
// 					idxIndex = idxIndex + 1
// 					cell, err := excelize.CoordinatesToCellName(2, idxIndex)
// 					if err != nil {
// 						log.Fatalf("Failed to get cell name: %v", err)
// 					}
// 					//Set At once
// 					if idxIndex == 2 || idxIndex == 3 {
// 						if err := f.SetSheetProps(ORACLE_TYPE_INDEX, &excelize.SheetPropsOptions{
// 							TabColorRGB: &existsColor,
// 						}); err != nil {
// 							fmt.Println(err)
// 						}
// 					}
// 					f.SetCellValue(ORACLE_TYPE_INDEX, fmt.Sprintf("A%d", idxIndex), db.Schema)
// 					f.SetCellValue(ORACLE_TYPE_INDEX, fmt.Sprintf("C%d", idxIndex), remarks)
// 					f.SetCellValue(ORACLE_TYPE_INDEX, fmt.Sprintf("D%d", idxIndex), userModified)
// 					f.SetCellValue(ORACLE_TYPE_INDEX, fmt.Sprintf("E%d", idxIndex), project+db.Schema)
// 					f.SetCellValue(ORACLE_TYPE_INDEX, cell, row.ObjectName)
// 				}
// 				if lastRowIndex > 0 && len(rowsIndex) > 1 {
// 					if err := f.SetSheetProps(ORACLE_TYPE_INDEX, &excelize.SheetPropsOptions{
// 						TabColorRGB: &existsColor,
// 					}); err != nil {
// 						fmt.Println(err)
// 					}
// 				}
// 			} else if row.ObjectType == ORACLE_TYPE_SEQUENCE {
// 				if lastRowSeq > 0 && !isExists(rowsSeq[0:], row.ObjectName) {
// 					idxSeq = idxSeq + 1
// 					cell, err := excelize.CoordinatesToCellName(2, idxSeq)
// 					if err != nil {
// 						log.Fatalf("Failed to get cell name: %v", err)
// 					}
// 					//Set At once
// 					if idxSeq == 2 || idxSeq == 3 {
// 						if err := f.SetSheetProps(ORACLE_TYPE_SEQUENCE, &excelize.SheetPropsOptions{
// 							TabColorRGB: &existsColor,
// 						}); err != nil {
// 							fmt.Println(err)
// 						}
// 					}
// 					f.SetCellValue(ORACLE_TYPE_SEQUENCE, fmt.Sprintf("A%d", idxSeq), db.Schema)
// 					f.SetCellValue(ORACLE_TYPE_SEQUENCE, fmt.Sprintf("C%d", idxSeq), remarks)
// 					f.SetCellValue(ORACLE_TYPE_SEQUENCE, fmt.Sprintf("D%d", idxSeq), userModified)
// 					f.SetCellValue(ORACLE_TYPE_SEQUENCE, fmt.Sprintf("E%d", idxSeq), project+db.Schema)
// 					f.SetCellValue(ORACLE_TYPE_SEQUENCE, cell, row.ObjectName)
// 				}
// 				if lastRowSeq > 0 && len(rowsSeq) > 1 {
// 					if err := f.SetSheetProps(ORACLE_TYPE_SEQUENCE, &excelize.SheetPropsOptions{
// 						TabColorRGB: &existsColor,
// 					}); err != nil {
// 						fmt.Println(err)
// 					}
// 				}
// 			}
// 		}
// 		if err := f.Save(); err != nil {
// 			fmt.Println(err)
// 		}
// 	}
// 	pathArchieve := GENERAL_PATH + db.Schema + "\\" + db.Enviroment + "\\Archived"
// 	EnsurePath(pathArchieve, MODE_DIR_PATH)
// 	var numFile int
// 	files, errData := os.ReadDir(pathArchieve)
// 	if errData != nil {
// 		numFile = 0
// 	} else {
// 		numFile = len(files)
// 	}
// 	version = fmt.Sprintf("V%d", numFile+1)
// 	//Get Object from Excel
// 	if modeGet == MODE_GET_OBJECT_EXCEL || modeGet == MODE_GET_OBJECT_COMBINE {
// 		_, version, _ = GetObjectFromExcel(&distinctObjects, db.Schema, db.Enviroment, version)
// 	}
// 	return distinctObjects, version
// }

/*
Get and Merge Object DB from Excel
Viktorianusl 23 June 2024
*/
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
				data.ObjectName = strings.TrimSpace(row[1])
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.Cell = fmt.Sprintf("A%d", i+1)
				obj = append(obj, data)
			}
		}
	}

	if len(rowsView) > 1 {
		data.ObjectType = ORACLE_TYPE_VIEW
		for i, row := range rowsView[1:] {
			if len(row) > 0 {
				data.ObjectName = strings.TrimSpace(row[1])
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.Cell = fmt.Sprintf("A%d", i+1)
				obj = append(obj, data)
			}
		}
	}

	if len(rowsMv) > 1 {
		data.ObjectType = ORACLE_TYPE_MV
		for i, row := range rowsMv[1:] {
			if len(row) > 0 {
				data.ObjectName = strings.TrimSpace(row[1])
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.Cell = fmt.Sprintf("A%d", i+1)
				obj = append(obj, data)
			}
		}
	}

	if len(rowsSeq) > 1 {
		data.ObjectType = ORACLE_TYPE_SEQUENCE
		for i, row := range rowsSeq[1:] {
			if len(row) > 0 {
				data.ObjectName = strings.TrimSpace(row[1])
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.Cell = fmt.Sprintf("A%d", i+1)
				obj = append(obj, data)
			}
		}
	}

	if len(rowsIndex) > 1 {
		data.ObjectType = ORACLE_TYPE_INDEX
		for i, row := range rowsIndex[1:] {
			if len(row) > 0 {
				data.ObjectName = strings.TrimSpace(row[1])
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.Cell = fmt.Sprintf("A%d", i+1)
				obj = append(obj, data)
			}
		}
	}

	if len(rowsTrigger) > 1 {
		data.ObjectType = ORACLE_TYPE_TRIGGER
		for i, row := range rowsTrigger[1:] {
			if len(row) > 0 {
				data.ObjectName = strings.TrimSpace(row[1])
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.Cell = fmt.Sprintf("A%d", i+1)
				obj = append(obj, data)
			}
		}
	}

	if len(rowsType) > 1 {
		data.ObjectType = ORACLE_TYPE_OBJECT
		for i, row := range rowsType[1:] {
			if len(row) > 0 {
				data.ObjectName = strings.TrimSpace(row[1])
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.Cell = fmt.Sprintf("A%d", i+1)
				obj = append(obj, data)
			}
		}
	}
	if len(rowsFunction) > 1 {
		data.ObjectType = ORACLE_TYPE_FUNCTION
		for i, row := range rowsFunction[1:] {
			if len(row) > 0 {
				data.ObjectName = strings.TrimSpace(row[1])
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.Cell = fmt.Sprintf("A%d", i+1)
				obj = append(obj, data)
			}
		}
	}
	if len(rowsProcedure) > 1 {
		data.ObjectType = ORACLE_TYPE_PROCEDURE
		for i, row := range rowsProcedure[1:] {
			if len(row) > 0 {
				data.ObjectName = strings.TrimSpace(row[1])
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.Cell = fmt.Sprintf("A%d", i+1)
				obj = append(obj, data)
			}
		}
	}
	//}

	objExc := make([]model.OracleUserObject, 0)
	rowsException, _ := f.GetRows(ORACLE_OBJECT_EXEPTION)
	if len(rowsException) > 1 {
		for _, row := range rowsException[1:] {
			if len(row) > 0 {
				data.ObjectOwner = strings.TrimSpace(row[0])
				data.ObjectName = strings.TrimSpace(row[1])
				data.ObjectType = strings.TrimSpace(row[2])
				objExc = append(objExc, data)
			}
		}
		exclude = append(exclude, objExc...)
	}

	//obj = DistinctSlice(obj)
	obj = FilterException(obj, objExc)

	return obj, exclude, errData
}

func GetObjectDBbyUser(user string, dta []model.OracleUserObject) (res []model.OracleUserObject) {
	for i := range dta {
		if strings.EqualFold(user, dta[i].ObjectOwner) {
			res = append(res, dta[i])
		}
	}
	return res
}

func MergeObjectDbToProdFile(n *excelize.File, schema string, env string) {
	var (
		newObjectIndex     []model.OracleUserObject
		newObjectSeq       []model.OracleUserObject
		newObjectType      []model.OracleUserObject
		newObjectFunction  []model.OracleUserObject
		newObjectProcedure []model.OracleUserObject
		newObjectException []model.OracleUserObject
	)
	filePath := GENERAL_PATH + schema + "\\" + env + "\\To Prod\\Object DB Eproc - " + schema + " - PROD.xlsx"
	EnsurePath(filePath, MODE_DIR_FILE)
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	rowsIndex, _ := f.GetRows(ORACLE_TYPE_INDEX)
	lastRowIndex := len(rowsIndex)
	rowsSeq, _ := f.GetRows(ORACLE_TYPE_SEQUENCE)
	lastRowSeq := len(rowsSeq)
	rowsType, _ := f.GetRows(ORACLE_TYPE_OBJECT)
	lastRowType := len(rowsType)
	rowsFunction, _ := f.GetRows(ORACLE_TYPE_FUNCTION)
	lastRowFunction := len(rowsFunction)
	rowsProcedure, _ := f.GetRows(ORACLE_TYPE_PROCEDURE)
	lastRowProcedure := len(rowsProcedure)
	rowsException, _ := f.GetRows(ORACLE_OBJECT_EXEPTION)
	lastRowException := len(rowsException)

	rowsIndexNew, _ := n.GetRows(ORACLE_TYPE_INDEX)
	rowsSeqNew, _ := n.GetRows(ORACLE_TYPE_SEQUENCE)
	rowsTypeNew, _ := n.GetRows(ORACLE_TYPE_OBJECT)
	rowsFunctionNew, _ := n.GetRows(ORACLE_TYPE_FUNCTION)
	rowsProcedureNew, _ := n.GetRows(ORACLE_TYPE_PROCEDURE)
	rowsExceptionNew, _ := n.GetRows(ORACLE_OBJECT_EXEPTION)

	if len(rowsIndexNew) > 1 {
		for _, row := range rowsIndexNew[1:] {
			if len(row) > 0 {
				if !isExists(rowsIndex[1:], row[1]) {
					newObjectIndex = append(newObjectIndex, model.OracleUserObject{ObjectName: row[1]})
				}
			}
		}
	}
	if len(rowsSeqNew) > 1 {
		for _, row := range rowsSeqNew[1:] {
			if len(row) > 0 {
				if !isExists(rowsSeq[1:], row[1]) {
					newObjectSeq = append(newObjectSeq, model.OracleUserObject{ObjectName: row[1]})
				}
			}
		}
	}
	if len(rowsTypeNew) > 1 {
		for _, row := range rowsTypeNew[1:] {
			if len(row) > 0 {
				if !isExists(rowsType[1:], row[1]) {
					newObjectType = append(newObjectType, model.OracleUserObject{ObjectName: row[1]})
				}
			}
		}
	}
	if len(rowsFunctionNew) > 1 {
		for _, row := range rowsFunctionNew[1:] {
			if len(row) > 0 {
				if !isExists(rowsFunction[1:], row[1]) {
					newObjectFunction = append(newObjectFunction, model.OracleUserObject{ObjectName: row[1]})
				}
			}
		}
	}
	if len(rowsProcedureNew) > 1 {
		for _, row := range rowsProcedureNew[1:] {
			if len(row) > 0 {
				if !isExists(rowsProcedure[1:], row[1]) {
					newObjectProcedure = append(newObjectProcedure, model.OracleUserObject{ObjectName: row[1]})
				}
			}
		}
	}
	if len(rowsExceptionNew) > 1 {
		for _, row := range rowsExceptionNew[1:] {
			if len(row) > 0 {
				if !isExists(rowsException[1:], row[0]) {
					newObjectException = append(newObjectException, model.OracleUserObject{ObjectName: row[0], ObjectType: row[1]})
				}
			}
		}
	}
	existsColor := "FFF333"
	remarks := "MODIFIED"
	userModified := "SYSTEM"
	project := "EPROC-"
	idxRow := 0
	//Update To Prod Object Db File
	for i, row := range newObjectIndex {
		idxRow = lastRowIndex + 1 + i
		cell, err := excelize.CoordinatesToCellName(2, idxRow)
		if err != nil {
			log.Fatalf("Failed to get cell name: %v", err)
		}
		//Set At once
		if idxRow == 2 {
			if err := f.SetSheetProps(ORACLE_TYPE_INDEX, &excelize.SheetPropsOptions{
				TabColorRGB: &existsColor,
			}); err != nil {
				fmt.Println(err)
			}
		}
		f.SetCellValue(ORACLE_TYPE_INDEX, fmt.Sprintf("A%d", idxRow), schema)
		f.SetCellValue(ORACLE_TYPE_INDEX, fmt.Sprintf("C%d", idxRow), remarks)
		f.SetCellValue(ORACLE_TYPE_INDEX, fmt.Sprintf("D%d", idxRow), userModified)
		f.SetCellValue(ORACLE_TYPE_INDEX, fmt.Sprintf("E%d", idxRow), project+schema)
		f.SetCellValue(ORACLE_TYPE_INDEX, cell, row.ObjectName)
	}
	idxRow = 0
	for i, row := range newObjectSeq {
		idxRow = lastRowSeq + 1 + i
		cell, err := excelize.CoordinatesToCellName(2, idxRow)
		if err != nil {
			log.Fatalf("Failed to get cell name: %v", err)
		}
		//Set At once
		if idxRow == 2 {
			if err := f.SetSheetProps(ORACLE_TYPE_SEQUENCE, &excelize.SheetPropsOptions{
				TabColorRGB: &existsColor,
			}); err != nil {
				fmt.Println(err)
			}
		}
		f.SetCellValue(ORACLE_TYPE_SEQUENCE, fmt.Sprintf("A%d", idxRow), schema)
		f.SetCellValue(ORACLE_TYPE_SEQUENCE, fmt.Sprintf("C%d", idxRow), remarks)
		f.SetCellValue(ORACLE_TYPE_SEQUENCE, fmt.Sprintf("D%d", idxRow), userModified)
		f.SetCellValue(ORACLE_TYPE_SEQUENCE, fmt.Sprintf("E%d", idxRow), project+schema)
		f.SetCellValue(ORACLE_TYPE_SEQUENCE, cell, row.ObjectName)
	}
	idxRow = 0
	for i, row := range newObjectType {
		idxRow = lastRowType + 1 + i
		cell, err := excelize.CoordinatesToCellName(2, idxRow)
		if err != nil {
			log.Fatalf("Failed to get cell name: %v", err)
		}
		//Set At once
		if idxRow == 2 {
			if err := f.SetSheetProps(ORACLE_TYPE_OBJECT, &excelize.SheetPropsOptions{
				TabColorRGB: &existsColor,
			}); err != nil {
				fmt.Println(err)
			}
		}
		f.SetCellValue(ORACLE_TYPE_OBJECT, fmt.Sprintf("A%d", idxRow), schema)
		f.SetCellValue(ORACLE_TYPE_OBJECT, fmt.Sprintf("C%d", idxRow), remarks)
		f.SetCellValue(ORACLE_TYPE_OBJECT, fmt.Sprintf("D%d", idxRow), userModified)
		f.SetCellValue(ORACLE_TYPE_OBJECT, fmt.Sprintf("E%d", idxRow), project+schema)
		f.SetCellValue(ORACLE_TYPE_OBJECT, cell, row.ObjectName)
	}
	idxRow = 0
	for i, row := range newObjectFunction {
		idxRow := lastRowFunction + 1 + i
		cell, err := excelize.CoordinatesToCellName(2, idxRow)
		if err != nil {
			log.Fatalf("Failed to get cell name: %v", err)
		}
		if idxRow == 2 {
			if err := f.SetSheetProps(ORACLE_TYPE_OBJECT, &excelize.SheetPropsOptions{
				TabColorRGB: &existsColor,
			}); err != nil {
				fmt.Println(err)
			}
		}
		f.SetCellValue(ORACLE_TYPE_FUNCTION, fmt.Sprintf("A%d", idxRow), schema)
		f.SetCellValue(ORACLE_TYPE_FUNCTION, fmt.Sprintf("C%d", idxRow), remarks)
		f.SetCellValue(ORACLE_TYPE_FUNCTION, fmt.Sprintf("D%d", idxRow), userModified)
		f.SetCellValue(ORACLE_TYPE_FUNCTION, fmt.Sprintf("E%d", idxRow), project+schema)
		f.SetCellValue(ORACLE_TYPE_FUNCTION, cell, row.ObjectName)
	}
	idxRow = 0
	for i, row := range newObjectProcedure {
		idxRow := lastRowProcedure + 1 + i
		cell, err := excelize.CoordinatesToCellName(2, idxRow)
		if err != nil {
			log.Fatalf("Failed to get cell name: %v", err)
		}
		if idxRow == 2 {
			if err := f.SetSheetProps(ORACLE_TYPE_OBJECT, &excelize.SheetPropsOptions{
				TabColorRGB: &existsColor,
			}); err != nil {
				fmt.Println(err)
			}
		}
		f.SetCellValue(ORACLE_TYPE_PROCEDURE, fmt.Sprintf("A%d", idxRow), schema)
		f.SetCellValue(ORACLE_TYPE_PROCEDURE, fmt.Sprintf("C%d", idxRow), remarks)
		f.SetCellValue(ORACLE_TYPE_PROCEDURE, fmt.Sprintf("D%d", idxRow), userModified)
		f.SetCellValue(ORACLE_TYPE_PROCEDURE, fmt.Sprintf("E%d", idxRow), project+schema)
		f.SetCellValue(ORACLE_TYPE_PROCEDURE, cell, row.ObjectName)
	}

	idxRow = 0
	for i, row := range newObjectException {
		idxRow := lastRowException + 1 + i
		cellObjName, err := excelize.CoordinatesToCellName(1, idxRow)
		cellObjType, err := excelize.CoordinatesToCellName(2, idxRow)
		if err != nil {
			log.Fatalf("Failed to get cell name: %v", err)
		}
		if idxRow == 2 {
			if err := f.SetSheetProps(ORACLE_OBJECT_EXEPTION, &excelize.SheetPropsOptions{
				TabColorRGB: &existsColor,
			}); err != nil {
				fmt.Println(err)
			}
		}
		f.SetCellValue(ORACLE_OBJECT_EXEPTION, cellObjName, row.ObjectName)
		f.SetCellValue(ORACLE_OBJECT_EXEPTION, cellObjType, row.ObjectType)
	}

	if err := f.Save(); err != nil {
		fmt.Println(err)
	}
}

func isExists(prodRows [][]string, row string) bool {
	for _, rowOld := range prodRows {
		if len(rowOld) > 0 {
			if strings.EqualFold(strings.TrimSpace(rowOld[1]), strings.TrimSpace(row)) {
				return true
			}
		}
	}
	return false
}

func DistinctSlice(s []model.OracleUserObject) []model.OracleUserObject {
	m := map[model.OracleUserObject]bool{}
	var unique []model.OracleUserObject
	for _, v := range s {
		if !m[v] {
			m[v] = true
			unique = append(unique, v)
		}
	}
	return unique
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

// func CompareFile() {
// 	b, err := os.ReadFile("file.txt")
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	strs := string(b)
// 	temp := strings.Split(strs, "\n")
// 	temp2 := SearchSameFile("file.txt")

// 	for i, v := range temp {
// 		for j, e := range temp2 {
// 			if strings.EqualFold(strings.TrimSpace(strings.ToLower(v)), strings.TrimSpace(strings.ToLower(e))) {
// 				if i == j { //line same
// 					break
// 				} else if i != j { //line different
// 					break
// 				}
// 			}
// 		}
// 	}
// }

// func SearchSameFile(fileName string) []string {
// 	b, err := os.ReadFile("file.txt")
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	strs := string(b)
// 	temp := strings.Split(strs, "\n")
// 	return temp
// }

func CompareObjectDb(listObjectDbAll []model.OracleUserObject, listObjectDbExcel []model.OracleUserObject, 
	listExclude []model.OracleUserObject, data model.DataExcel)([]model.OracleUserObject, []model.OracleUserObject){
	var (
		listObjFromAll []model.OracleUserObject
		found,similar   bool
	)

	listObjEnvSrc, listObjEnvTrg := GetObjectByEnv(listObjectDbExcel, data.EnvSource, data.EnvTarget)
	listObjEnvSrcAll, listObjEnvTrgAll := GetObjectByEnv(listObjectDbAll, data.EnvSource, data.EnvTarget)

	OrderObjDb(listObjEnvSrc)
	OrderObjDb(listObjEnvTrg)
	OrderObjDb(listObjEnvSrcAll)
	OrderObjDb(listObjEnvTrgAll)

	// Compare Object DB from Excel
	for i, src := range listObjEnvSrc {
		found = false
		similar = false
		if src.Ddl == "" {
			listObjEnvSrc[i].ObjectStatus = "MISSING_SOURCE" //Object tidak ada DDL/disource, ada di list
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
		if !found && v.ObjectStatus != "" {
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
	case ORACLE_TYPE_PACKAGE:
		seq = 6
	case ORACLE_TYPE_OBJECT:
		seq = 7
	case ORACLE_TYPE_FUNCTION:
		seq = 8
	case ORACLE_TYPE_PROCEDURE:
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