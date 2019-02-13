package main

import (	
	"os"
	_ "github.com/go-sql-driver/mysql"
	models "turbi.com.br/financial/models"
)

// PayoutsF ...
var PayoutsF []models.Payouts

// BanlanceTransfer ...
var BanlanceTransfer []string

// PayoutValue ...
var PayoutValue []string

// PSPcheckDB ...
var PSPcheckDB []models.DBCheckPSP

// DBvaluesCheckDb ...
var DBvaluesCheckDb []models.DBValues

// AnalisysRFS ...
var AnalisysRFS []models.AnalysisLot

var ym string
var loteAtual string

func main() {

	var xlsx []models.XLSX
	var sumXLSX models.SumXLSX
	var dbValues []models.DBValues

	if len(os.Args[1:]) != 0 {
		if os.Args[1:][0] != "adyen" && os.Args[1:][0] != "populatedb" {
			ym = os.Args[1:][0]
		} else {
			ym = ""
		}
	} else {
		ym = ""
	}

	WgetCSV()
	MergeFiles()
	GetPayOut()
	xlsx, sumXLSX, dbValues = CreateDetails()
	CheckDB()
	AnalisysCF()
	GenerateXLSX(xlsx, sumXLSX, dbValues)

	//PopulateDB(xlsx)

	if len(os.Args[1:]) != 0 {
		if os.Args[1:][0] == "adyen" {
			ReconciliationAdyen(xlsx)
		}
	}

	if len(os.Args[1:]) != 0 {
		if os.Args[1:][0] == "populatedb" {
			PopulateDB(xlsx)
		}
	}
}

// checkErr ...
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
