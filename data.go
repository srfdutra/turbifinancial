package main

import (
	"time"
	"strings"
	"fmt"
	"turbi.com.br/financial/models"
	"database/sql"
)

//var dsn = "root:GreatTeam8*@tcp(127.0.0.1:3307)/dbsqlt890"
var dsn = "root:000000@tcp(127.0.0.1:3306)/dbsqlt890"
var db, err = sql.Open("mysql", dsn)

// AnalisysCF ...
func AnalisysCF() {
	results, err := db.Query(`select advancementCode,  sum(dbValue) as DB, sum(commissionAdyen), sum(grossCredit- commissionAdyen) as Adyen,
	sum(commissionBS),sum(grossCredit -grossDebit - commissionBS - commissionAdyen) as BS from conciliation group by advancementCode `)
	if err != nil {
		panic(err.Error())
	}

	var analisysLot models.AnalysisLot

	for results.Next() {

		err = results.Scan(&analisysLot.AdvancementCode, &analisysLot.DB, &analisysLot.CommissionAdyen, &analisysLot.Adyen, 
			&analisysLot.CommissionBS, &analisysLot.BSTurbi)
		if err != nil {
			panic(err.Error())
		}
		analisysLot.BSReal, analisysLot.DataAnt, analisysLot.DataPag = ReadBSOperations(fmt.Sprint(analisysLot.AdvancementCode))
		AnalisysRFS = append(AnalisysRFS, analisysLot)
	}
}

// PopulateDB ...
func PopulateDB(xlsxFull []models.XLSX) {
	count := 0

	for count < len(xlsxFull) {
		dt := strings.Split(xlsxFull[count].DateTransaction, " ")
		dtf := dt[0] + "T" + dt[1] + "Z"
		DateTransaction, err := time.Parse(time.RFC3339, dtf)

		dc := "C"
		if xlsxFull[count].GrossCredit == 0 {
			dc = "D"
		}

		commissionBS := ReadBS(xlsxFull[count].AdvancementBatch)

		stmt, err := db.Prepare(`INSERT INTO conciliation (
					batch,
					pspReference,
					pspModification,
					dateTransaction,
					advancement,
					advancementBatch,
					advancementCode,
					grossCredit,
					grossDebit,
					netCredit,
					netDebit,
					commissionAdyen,
					commissionBS,
					dbValue,
					flag,
					dc) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)

		checkErr(err)

		_, err = stmt.Exec(xlsxFull[count].Batch, xlsxFull[count].PSPReference, xlsxFull[count].PSPModification,
			DateTransaction.Format("2006-01-02 15:04:05"), xlsxFull[count].Advancement, xlsxFull[count].AdvancementBatch, xlsxFull[count].AdvancementCode,
			xlsxFull[count].GrossCredit, xlsxFull[count].GrossDebit, xlsxFull[count].NetCredit, xlsxFull[count].NetDebit,
			xlsxFull[count].Commission, commissionBS, xlsxFull[count].DBvalue, xlsxFull[count].Flag, dc)
		checkErr(err)

		fmt.Println(count)

		count++
	}
}

// ReconciliationAdyen ...
func ReconciliationAdyen(xlsxFull []models.XLSX) {
	count := 0

	for count < len(xlsxFull) {
		if xlsxFull[count].DBvalue == 0 {

			var numberRows int

			rows := db.QueryRow(`SELECT id FROM transaction where transactionid = '` + xlsxFull[count].PSPModification + "'")
			err = rows.Scan(&numberRows)

			if numberRows == 0 {
				results, err := db.Query(`SELECT id, transactionid, bookingid, state, token, verb FROM transaction 
			where state <> 'PENDING' and state <> 'DECLINED' and  gateway = 'adyen' and transactionid = ?`, xlsxFull[count].PSPReference)
				if err != nil {
					panic(err.Error())
				}

				var transaction models.Transaction

				for results.Next() {
					err = results.Scan(&transaction.ID, &transaction.TransactionID, &transaction.BookingID, &transaction.State, &transaction.Token,
						&transaction.Verb)
					if err != nil {
						panic(err.Error())
					}
				}

				if transaction.Verb == "AUTHORIZATION" {
					dt := strings.Split(xlsxFull[count].DateTransaction, " ")
					dtf := dt[0] + "T" + dt[1] + "Z"
					DateTransaction, err := time.Parse(time.RFC3339, dtf)

					stmt, err := db.Prepare("INSERT transaction SET transactionid=?,bookingid=?,token=?, parentid=?, value=? , operationDate=?, transactiondate=?, verb='REFUND',state='APPROVED', responseCode='CONSILATION', gateway = 'adyen' ")
					checkErr(err)

					_, err = stmt.Exec(xlsxFull[count].PSPModification, transaction.BookingID,
						transaction.Token, transaction.ID, xlsxFull[count].GrossDebit, DateTransaction.UnixNano()/1000000, DateTransaction.Format("2006-01-02 15:04:05"))
					checkErr(err)
				}
			}
		}
		count++
	}

}

// DirtyStruct ...
func DirtyStruct() (dbValues []models.DBValues) {
	results, err := db.Query("SELECT transactionid, state, verb, responsecode, value FROM transaction where state <> 'PENDING' and  gateway = 'adyen'")
	if err != nil {
		panic(err.Error())
	}

	var dbValue models.DBValues

	for results.Next() {
		err = results.Scan(&dbValue.PSP, &dbValue.State, &dbValue.Verb, &dbValue.ResponseCode, &dbValue.Value)
		if err != nil {
			panic(err.Error())
		}

		dbValues = append(dbValues, dbValue)

	}

	return dbValues
}


// CheckDB ...
func CheckDB() {
	results, err := db.Query("SELECT transactionid, state, verb, responsecode, value FROM transaction where state = 'PENDING' and  gateway = 'adyen'")
	if err != nil {
		panic(err.Error())
	}

	var dbValue models.DBValues

	for results.Next() {
		err = results.Scan(&dbValue.PSP, &dbValue.State, &dbValue.Verb, &dbValue.ResponseCode, &dbValue.Value)
		if err != nil {
			panic(err.Error())
		}

		DBvaluesCheckDb = append(DBvaluesCheckDb, dbValue)

	}
}

