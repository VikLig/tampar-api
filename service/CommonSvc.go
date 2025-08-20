package service

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"tampar-api/mapper"
	"tampar-api/model"
	"tampar-api/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type CommonSvc struct {
	handler      utils.RequestHandler
	logger       utils.Logger
	config       utils.Config
	commonMapper mapper.CommonMapper
}

func NewCommonSvc(
	handler utils.RequestHandler,
	logger utils.Logger,
	config utils.Config,
	commonMapper mapper.CommonMapper,
) CommonSvc {
	return CommonSvc{
		handler:      handler,
		logger:       logger,
		config:       config,
		commonMapper: commonMapper,
	}
}

func (s CommonSvc) Process(c *gin.Context) {
	var (
		criteria model.DataExcel
		data model.DataExcel
		zipFileName string
		zipBytes []byte
		errData error
	)
	
	errData = c.ShouldBindJSON(&criteria)
	if errData != nil {
		return
	}
	data.UseExcel = criteria.UseExcel
	data.EnvSource = criteria.EnvSource
	data.EnvTarget = criteria.EnvTarget
	data.Mode = criteria.Mode
	data.OutputMode = criteria.OutputMode
	data.Schema = criteria.Schema

	if criteria.UseExcel == "Y" && criteria.ExcelFile == nil {
		c.JSON(ErrorBody(errors.New("Excel file is required")))
		return
	}

	if data.Mode == "COMPARE" {
		zipBytes, zipFileName, errData = CompareObject(data, criteria.ExcelFile)
		if errData != nil {
			c.String(500, "Failed to create zip: %v", errData)
			return
		} else if zipFileName == "" || zipBytes == nil {
			c.String(400, "Empty Object DB")
			return
		} 
	} else if data.Mode == "GENERATE" {
		zipBytes, zipFileName, errData = GenerateObject(data, criteria.ExcelFile)
		if errData != nil {
			c.String(500, "Failed to create zip: %v", errData)
			return
		} else if zipFileName == "" || zipBytes == nil {
			c.String(400, "Empty Object DB")
			return
		} 
	}

	// Set header agar langsung download
	// c.Header("Content-Description", "File Transfer")
	// c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(zipFileName)))
	// c.Header("Content-Type", "application/zip")
	// c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	// c.Data(200, "application/zip", zipBytes)
	// c.Writer.Header().Add("Content-Description", "File Transfer")
	// c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(zipFileName)))
	// c.Writer.Header().Add("Content-Type", "application/zip")
	// c.Writer.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(zipFileName)))
	c.Header("Content-Type", "application/zip")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Data(200, "application/zip", zipBytes)

	// reader := bytes.NewReader(zipBytes)
	// _, errData = io.Copy(c.Writer, reader)
	// if errData != nil {
	// 	c.JSON(ErrorBody(errData))
	// 	return
	// }

	//c.JSON(SuccessBody(nil, "Success"))
}

func GenerateObject(data model.DataExcel, excelFile []byte)([]byte, string, error) {
	var (
		byteReader   *bytes.Reader
		xlsx         *excelize.File
		errData      error
		listObjectDb []model.OracleUserObject
		oraSourceDbList []model.Database
	)

	if data.UseExcel == "Y" {
		byteReader = bytes.NewReader(excelFile)
		xlsx, errData = excelize.OpenReader(byteReader)
		if errData != nil {
			return nil, "", errData
		}
		listObjectDb, _, errData = GetObjectFromExcel(xlsx)
		if len(listObjectDb) == 0 {
			return nil, "", nil
		} else if errData != nil {
			return nil, "", errData
		}
		data.Schema = GetSchemaByObject(listObjectDb)
	}

	oraSourceDbList = GetOraSource(data.Schema, data.EnvSource)
	listObjectDb = GetListObjectDb(oraSourceDbList, listObjectDb, data)
	
	return CreateFileObjectDB(listObjectDb, data.EnvSource)
}

func CompareObject(data model.DataExcel, excelFile []byte) ([]byte, string, error) {
	var (
		byteReader 		*bytes.Reader
		xlsx       		*excelize.File
		errData    		error
		listObjectDb 	[]model.OracleUserObject
		listExclude 	[]model.OracleUserObject
	)

	if data.UseExcel == "Y" {
		byteReader = bytes.NewReader(excelFile)
		xlsx, errData = excelize.OpenReader(byteReader)
		if errData != nil {
			return nil, "", errData
		}
		listObjectDb, listExclude, errData = GetObjectFromExcel(xlsx)
		if len(listObjectDb) == 0 {
			return nil, "", nil
		} else if errData != nil {
			return nil, "", errData
		}
		data.Schema = GetSchemaByObject(listObjectDb)
	}

	listObjectDbAll := GetListObjectDb(GetOracleDB(data), listObjectDb, data)
	listResultAll, listResultExcel := CompareObjectDb(listObjectDbAll, listObjectDb, listExclude, data)
	return CreateFileObjectDBCompare(listResultAll, listResultExcel, data)
}

func (s CommonSvc) DownloadTemplate(c *gin.Context) {
	timestamp := time.Now().Format("20060102150405")
	f := excelize.NewFile()
	defer func() {
        if err := f.Close(); err != nil {
            fmt.Println(err)
        }
    }()
	
	s.makeTemplateExcel(f)

	//Save File
	filename := fmt.Sprintf("Template-Tampar-Object-DB-%s.xlsx", timestamp)
	byteBuff, errData := f.WriteToBuffer()
	if errData != nil {
		c.JSON(ErrorBody(errData))
		return
	}

	// c.Writer.Header().Add("Content-Description", "File Transfer")
	// c.Writer.Header().Add("Content-Disposition", "attachment; filename="+filename)
	// c.Writer.Header().Add("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	// c.Writer.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")

	// reader := bytes.NewReader(byteBuff.Bytes())
	// _, errData = io.Copy(c.Writer, reader)
	// if errData != nil {
	// 	c.JSON(ErrorBody(errData))
	// 	return
	// }

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(filename)))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", byteBuff.Bytes())

	//c.JSON(SuccessBody(nil, "Success"))
}

func (s CommonSvc) makeTemplateExcel(f *excelize.File){
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
	sheetException := "EXCEPTION"
	f.SetSheetName(f.GetSheetName(0), sheetTable)
	f.NewSheet(sheetView)
	f.NewSheet(sheetMv)
	f.NewSheet(sheetSeq)
	f.NewSheet(sheetIndex)
	f.NewSheet(sheetType)
	f.NewSheet(sheetFunction)
	f.NewSheet(sheetProcedure)
	f.NewSheet(sheetTrigger)
	f.NewSheet(sheetException)

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

	for i := range sheets {
		
		f.SetColWidth(sheets[i], "A", "A", 20)
		f.SetColWidth(sheets[i], "B", "B", 45)
		f.SetColWidth(sheets[i], "C", "D", 20)
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
