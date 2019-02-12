package main

import (
	"bufio"
	"database/sql"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
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

//var dsn = "root:GreatTeam8*@tcp(127.0.0.1:3307)/dbsqlt890"
var dsn = "root:000000@tcp(127.0.0.1:3306)/dbsqlt890"
var db, err = sql.Open("mysql", dsn)

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

// ReadBS ...
func ReadBS(advancementBatch string) (commissionBS float64) {
	xlsx, err := excelize.OpenFile("LotesAntecipados.xlsx")
	checkErr(err)
	commissionBS = 0
	rows := xlsx.GetRows("LotesAntecipadosExterno")
	x := 0
	for x < len(rows) {
		if fmt.Sprint(rows[x][0]) == advancementBatch && loteAtual != advancementBatch {
			commissionBS, err = strconv.ParseFloat(rows[x][8], 64)
			loteAtual = advancementBatch
			return commissionBS
		}
		x++
	}
	return commissionBS
}

// ReadBSOperations ...
func ReadBSOperations(batchCod string) (valueBS float64) {
	xlsx, err := excelize.OpenFile("OperacoesPedidoExterno.xlsx")
	checkErr(err)
	valueBS = 0
	rows := xlsx.GetRows("OperacoesPedidoExterno")
	x := 0
	for x < len(rows) {
		if fmt.Sprint(rows[x][9]) == batchCod {
			valueF := strings.Replace(rows[x][5], "R$", "", 10)
			valueF = strings.Replace(valueF, ".", "", 10)
			valueF = strings.Replace(valueF, ",", ".", 10)
			valueF = strings.Replace(valueF, " ", "", 10)
			valueBS, err = strconv.ParseFloat(valueF, 64)

			return valueBS
		}
		x++
	}
	return valueBS
}

// AnalisysCF ...
func AnalisysCF() {
	results, err := db.Query(`select advancementCode,  sum(dbValue) as DB, sum(grossCredit- commissionAdyen) as Adyen,
	sum(grossCredit -grossDebit - commissionBS - commissionAdyen) as BS from conciliation group by advancementCode `)
	if err != nil {
		panic(err.Error())
	}

	var analisysLot models.AnalysisLot

	for results.Next() {

		err = results.Scan(&analisysLot.AdvancementCode, &analisysLot.DB, &analisysLot.Adyen, &analisysLot.BSTurbi)
		if err != nil {
			panic(err.Error())
		}
		analisysLot.BSReal = ReadBSOperations(fmt.Sprint(analisysLot.AdvancementCode))
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

// CreateDetails ...
func CreateDetails() (xlsx []models.XLSX, sumXLSX models.SumXLSX, dbValues []models.DBValues) {

	dbValues = DirtyStruct()

	filename := "merged.csv"

	var soma float64

	f, _ := os.Open(filename)
	g := 1
	r := csv.NewReader(bufio.NewReader(f))
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		_, err = strconv.ParseInt(record[8], 10, 64)
		if err != nil {
		} else {
			runes := []rune(record[5])
			yearMonthSubstring := ""
			if ym != "" {
				yearMonthSubstring = string(runes[0:7])
			}
			if yearMonthSubstring == ym {
				var lineXlsx models.XLSX
				lineXlsx.PSPModification = record[8]
				lineXlsx.PSPReference = record[2]
				lineXlsx.Flag = record[4]
				lineXlsx.DateTransaction = record[5]

				// _, err = strconv.ParseInt(record[22], 10, 64)
				// if err == nil {
				// 	lineXlsx.Batch = record[22]
				// } else {
				lineXlsx.Batch = record[24]
				// }

				lineXlsx.GrossDebit, err = strconv.ParseFloat(record[10], 64)
				lineXlsx.GrossCredit, err = strconv.ParseFloat(record[11], 64)
				lineXlsx.NetDebit, err = strconv.ParseFloat(record[14], 64)
				lineXlsx.NetCredit, err = strconv.ParseFloat(record[15], 64)
				lineXlsx.Commission, err = strconv.ParseFloat(record[16], 64)
				lineXlsx.Advancement, err = strconv.ParseFloat(record[20], 64)
				lineXlsx.AdvancementBatch = record[22]
				lineXlsx.AdvancementCode, err = strconv.ParseInt(record[21], 10, 64)
				// TESTE DE DICIONARIO
				// var details = map[DBValues]string{}
				// bdValuesDict := make(map[DBValues]string) //Shorthand and make
				// var t = map[string]DBValues{}
				// bdValuesDict["PSP"]

				x := 1
				for x < len(dbValues) {
					if dbValues[x].PSP == lineXlsx.PSPModification {
						lineXlsx.DBvalue = dbValues[x].Value
						break
					}

					x++
				}

				sumXLSX.Commission = sumXLSX.Commission + lineXlsx.Commission

				sumXLSX.Gross = sumXLSX.Gross + (lineXlsx.GrossCredit - lineXlsx.GrossDebit)

				sumXLSX.Net = sumXLSX.Net + (lineXlsx.NetCredit - lineXlsx.NetDebit - lineXlsx.Advancement)

				if lineXlsx.GrossCredit != 0 {
					sumXLSX.GrossDB = sumXLSX.GrossDB + lineXlsx.DBvalue
				} else {
					sumXLSX.GrossDB = sumXLSX.GrossDB - lineXlsx.DBvalue
				}

				xlsx = append(xlsx, lineXlsx)

				if lineXlsx.DBvalue == 0 {
					soma = soma + lineXlsx.GrossDebit
				}
			}
		}
		g++

	}

	return xlsx, sumXLSX, dbValues
}

// GenerateXLSX ...
func GenerateXLSX(xlsxFull []models.XLSX, sumXLSX models.SumXLSX, dbValues []models.DBValues) {
	xlsx := excelize.NewFile()
	var balance []string
	var batchValue float64
	var batchValue2 float64
	var dbcheckPSP models.DBCheckPSP

	xlsx.NewSheet("Details")
	xlsx.DeleteSheet("Sheet1")

	styleHeader, _ := xlsx.NewStyle(`{"fill":{"type":"pattern","color":["#81BEF7"],"pattern":1}}`)
	

	xlsx.SetCellValue("Details", "A1", "Batch")
	xlsx.SetCellValue("Details", "B1", "PSP Reference")
	xlsx.SetCellValue("Details", "C1", "PSP Modification")
	xlsx.SetCellValue("Details", "D1", "Flag")
	xlsx.SetCellValue("Details", "E1", "Date Operation")
	xlsx.SetCellValue("Details", "F1", "Gross Credit")
	xlsx.SetCellValue("Details", "G1", "Gross Debit")
	xlsx.SetCellValue("Details", "H1", "Net Credit")
	xlsx.SetCellValue("Details", "I1", "Net Debit")
	xlsx.SetCellValue("Details", "J1", "Commission")
	xlsx.SetCellValue("Details", "K1", "DB Value")
	xlsx.SetCellValue("Details", "L1", "Adyen Operation")
	xlsx.SetCellValue("Details", "M1", "Advancement")
	xlsx.SetCellValue("Details", "N1", "Advancement Batch")
	xlsx.SetCellValue("Details", "O1", "Advancement Code")
	xlsx.SetCellValue("Details", "P1", "Sum Advancement")

	xlsx.SetPanes("Details", `{"freeze":true,"split":false,"x_split":0,"y_split":1,"top_left_cell":"B1","active_pane":"topRight"}`)
	xlsx.SetCellStyle("Details", "A1", "P1", styleHeader)

	// sort details
	sort.Slice(xlsxFull[:], func(i, j int) bool {
		return xlsxFull[i].AdvancementBatch < xlsxFull[j].AdvancementBatch
	})

	col := "0"
	count := 0
	line := 2
	batchB := xlsxFull[count].Batch
	batchAdvancementB := xlsxFull[count].AdvancementBatch
	for count < len(xlsxFull) {
		col = strconv.Itoa(line)
		xlsx.SetCellValue("Details", "A"+col, xlsxFull[count].Batch)
		xlsx.SetCellValue("Details", "B"+col, xlsxFull[count].PSPReference)
		xlsx.SetCellValue("Details", "C"+col, xlsxFull[count].PSPModification)
		xlsx.SetCellValue("Details", "D"+col, xlsxFull[count].Flag)
		xlsx.SetCellValue("Details", "E"+col, xlsxFull[count].DateTransaction)
		xlsx.SetCellValue("Details", "F"+col, xlsxFull[count].GrossCredit)
		xlsx.SetCellValue("Details", "G"+col, xlsxFull[count].GrossDebit)
		xlsx.SetCellValue("Details", "H"+col, xlsxFull[count].NetCredit)
		xlsx.SetCellValue("Details", "I"+col, xlsxFull[count].NetDebit)
		xlsx.SetCellValue("Details", "J"+col, xlsxFull[count].Commission)
		xlsx.SetCellValue("Details", "K"+col, xlsxFull[count].DBvalue)
		xlsx.SetCellValue("Details", "L"+col, 0)
		xlsx.SetCellValue("Details", "M"+col, xlsxFull[count].Advancement)
		xlsx.SetCellValue("Details", "N"+col, xlsxFull[count].AdvancementBatch)
		xlsx.SetCellValue("Details", "O"+col, xlsxFull[count].AdvancementCode)

		if xlsxFull[count].DBvalue == 0 {
			xlsx.SetCellValue("Details", "L"+col, xlsxFull[count].GrossDebit)

			dbcheckPSP.PSPReference = xlsxFull[count].PSPReference
			dbcheckPSP.PSPModification = xlsxFull[count].PSPModification
			dbcheckPSP.Type = "Adyen"
			if xlsxFull[count].GrossCredit == 0 {
				dbcheckPSP.Value = xlsxFull[count].GrossDebit
			} else {
				dbcheckPSP.Value = xlsxFull[count].GrossCredit
			}

			PSPcheckDB = append(PSPcheckDB, dbcheckPSP)
		}

		batch := xlsxFull[count].Batch

		i := 0
		if batchB != batch {
			for i < len(BanlanceTransfer) {
				findS := "batch " + xlsxFull[count].Batch
				if strings.Contains(BanlanceTransfer[i], findS) {
					balance = append(balance, BanlanceTransfer[i])
					balance = append(balance, PayoutValue[i])
					balance = append(balance, "")
					batchValue = 0

				}
				i++
			}
			batchB = xlsxFull[count].Batch
		}
		if xlsxFull[count].AdvancementCode != 0 {
			batchbatchAdvancement := xlsxFull[count].AdvancementBatch
			if batchAdvancementB != batchbatchAdvancement {
				l, _ := strconv.Atoi(col)
				l = l - 1
				if batchValue2 != 0 {
					xlsx.SetCellValue("Details", "P"+strconv.Itoa(l), batchValue2)
				}
				batchValue2 = 0
				batchAdvancementB = xlsxFull[count].AdvancementBatch
			}
		}

		batchValue = batchValue + (xlsxFull[count].NetCredit - xlsxFull[count].NetDebit - xlsxFull[count].Advancement)
		batchValue2 = batchValue2 + xlsxFull[count].Advancement //(xlsxFull[count].GrossCredit - xlsxFull[count].GrossDebit - xlsxFull[count].Commission)

		line++
		count++
	}

	xlsx.MergeCell("Details", "F"+strconv.Itoa(line), "G"+strconv.Itoa(line))
	xlsx.MergeCell("Details", "H"+strconv.Itoa(line), "I"+strconv.Itoa(line))

	xlsx.SetCellFormula("Details", "L"+strconv.Itoa(line), "SUM(L1:L"+strconv.Itoa(line-1)+")")
	xlsx.SetCellFormula("Details", "M"+strconv.Itoa(line), "SUM(M1:M"+strconv.Itoa(line-1)+")")
	xlsx.SetCellFormula("Details", "L"+strconv.Itoa(line+1), "SUM(K"+strconv.Itoa(line)+"-L"+strconv.Itoa(line)+")")
	xlsx.SetCellFormula("Details", "P"+strconv.Itoa(line), "SUM(P1:P"+strconv.Itoa(line-1)+")")

	styleNumber, _ := xlsx.NewStyle(`{"number_format": 26, "decimal_places": 2}`)

	xlsx.SetCellValue("Details", "F"+strconv.Itoa(line), sumXLSX.Gross)
	xlsx.SetCellValue("Details", "H"+strconv.Itoa(line), sumXLSX.Net)
	xlsx.SetCellValue("Details", "J"+strconv.Itoa(line), sumXLSX.Commission)
	xlsx.SetCellValue("Details", "K"+strconv.Itoa(line), sumXLSX.GrossDB)

	xlsx.SetCellStyle("Details", "F"+strconv.Itoa(line), "K"+strconv.Itoa(line), styleNumber)
	xlsx.SetCellStyle("Details", "H"+strconv.Itoa(line), "H"+strconv.Itoa(line), styleNumber)
	xlsx.SetCellStyle("Details", "J"+strconv.Itoa(line), "J"+strconv.Itoa(line), styleNumber)
	xlsx.SetCellStyle("Details", "K"+strconv.Itoa(line), "K"+strconv.Itoa(line), styleNumber)
	xlsx.SetCellStyle("Details", "L"+strconv.Itoa(line+1), "L"+strconv.Itoa(line+1), styleNumber)
	xlsx.SetCellStyle("Details", "M"+strconv.Itoa(line+1), "O"+strconv.Itoa(line+1), styleNumber)

	xlsx.SetColWidth("Details", "B", "C", 22)
	xlsx.SetColWidth("Details", "E", "E", 22)
	xlsx.SetColWidth("Details", "F", "K", 12)
	xlsx.SetColWidth("Details", "L", "L", 15)
	xlsx.SetColWidth("Details", "M", "M", 15)
	xlsx.SetColWidth("Details", "N", "N", 25)
	xlsx.SetColWidth("Details", "O", "O", 25)
	xlsx.SetColWidth("Details", "P", "P", 25)

	xlsx.NewSheet("PayOuts")
	xlsx.SetCellValue("PayOuts", "A1", "Payout Date")
	xlsx.SetCellValue("PayOuts", "B1", "Bank")
	xlsx.SetCellValue("PayOuts", "C1", "Branch")
	xlsx.SetCellValue("PayOuts", "D1", "Account")
	xlsx.SetCellValue("PayOuts", "E1", "Code Transaction")
	xlsx.SetCellValue("PayOuts", "F1", "Batch")
	xlsx.SetCellValue("PayOuts", "G1", "Value")
	xlsx.SetCellValue("PayOuts", "H1", "Balance Transfer")
	xlsx.SetCellValue("PayOuts", "I1", "Flag")

	xlsx.SetPanes("PayOuts", `{"freeze":true,"split":false,"x_split":0,"y_split":1,"top_left_cell":"B1","active_pane":"topRight"}`)
	xlsx.SetCellStyle("PayOuts", "A1", "I1", styleHeader)

	// sort payouts
	sort.Slice(PayoutsF[:], func(i, j int) bool {
		return PayoutsF[i].Batch < PayoutsF[j].Batch
	})

	i := 0
	line = 2
	for i < len(PayoutsF) {
		if PayoutsF[i].Flag != "" {
			xlsx.SetCellValue("PayOuts", "A"+strconv.Itoa(line), PayoutsF[i].PayoutDate)
			xlsx.SetCellValue("PayOuts", "B"+strconv.Itoa(line), PayoutsF[i].Bank)
			xlsx.SetCellValue("PayOuts", "C"+strconv.Itoa(line), PayoutsF[i].Branch)
			xlsx.SetCellValue("PayOuts", "D"+strconv.Itoa(line), PayoutsF[i].Account)
			xlsx.SetCellValue("PayOuts", "E"+strconv.Itoa(line), PayoutsF[i].CodeTransaction)
			xlsx.SetCellValue("PayOuts", "F"+strconv.Itoa(line), PayoutsF[i].Batch)
			xlsx.SetCellValue("PayOuts", "G"+strconv.Itoa(line), PayoutsF[i].Value)
			xlsx.SetCellValue("PayOuts", "H"+strconv.Itoa(line), PayoutsF[i].BanlanceTransfer)
			xlsx.SetCellValue("PayOuts", "I"+strconv.Itoa(line), PayoutsF[i].Flag)

			xlsx.SetCellStyle("PayOuts", "B"+strconv.Itoa(line), "G"+strconv.Itoa(line), styleNumber)
			line++
		}
		i++
	}

	xlsx.SetColWidth("PayOuts", "A", "A", 20)
	xlsx.SetColWidth("PayOuts", "B", "D", 10)
	xlsx.SetColWidth("PayOuts", "F", "G", 12)
	xlsx.SetColWidth("PayOuts", "E", "E", 20)
	xlsx.SetColWidth("PayOuts", "H", "H", 20)

	xlsx.SetCellFormula("PayOuts", "G"+strconv.Itoa(line), "SUM(G1:G"+strconv.Itoa(line-1)+")")
	xlsx.SetCellFormula("PayOuts", "H"+strconv.Itoa(line), "SUM(H1:H"+strconv.Itoa(line-1)+")")

	xlsx.NewSheet("CheckDB")

	xlsx.SetCellValue("CheckDB", "A1", "PSP Reference")
	xlsx.SetCellValue("CheckDB", "B1", "PSP Modification")
	xlsx.SetCellValue("CheckDB", "C1", "Value")
	xlsx.SetCellValue("CheckDB", "D1", "Type")

	styleHeaderCheckDB, _ := xlsx.NewStyle(`{"fill":{"type":"pattern","color":["#81BEF7"],"pattern":1}}`)

	xlsx.SetPanes("CheckDB", `{"freeze":true,"split":false,"x_split":0,"y_split":1,"top_left_cell":"B1","active_pane":"topRight"}`)
	xlsx.SetCellStyle("CheckDB", "A1", "D1", styleHeaderCheckDB)

	i = 0
	line = 2

	for i < len(PSPcheckDB) {
		xlsx.SetCellValue("CheckDB", "A"+strconv.Itoa(line), PSPcheckDB[i].PSPReference)
		xlsx.SetCellValue("CheckDB", "B"+strconv.Itoa(line), PSPcheckDB[i].PSPModification)
		xlsx.SetCellValue("CheckDB", "C"+strconv.Itoa(line), PSPcheckDB[i].Value)
		xlsx.SetCellValue("CheckDB", "D"+strconv.Itoa(line), PSPcheckDB[i].Type)
		line++
		i++
	}
	i = 0
	for i < len(DBvaluesCheckDb) {
		xlsx.SetCellValue("CheckDB", "A"+strconv.Itoa(line), DBvaluesCheckDb[i].PSP)
		xlsx.SetCellValue("CheckDB", "B"+strconv.Itoa(line), "")
		xlsx.SetCellValue("CheckDB", "C"+strconv.Itoa(line), DBvaluesCheckDb[i].Value)
		xlsx.SetCellValue("CheckDB", "D"+strconv.Itoa(line), "Turbi")
		line++
		i++
	}
	xlsx.SetColWidth("CheckDB", "A", "C", 22)


	xlsx.NewSheet("Conciliation")
	xlsx.SetCellValue("Conciliation", "A1", "Advancement Code")
	xlsx.SetCellValue("Conciliation", "B1", "DB")
	xlsx.SetCellValue("Conciliation", "C1", "Adyen")
	xlsx.SetCellValue("Conciliation", "D1", "BS Turbi")
	xlsx.SetCellValue("Conciliation", "E1", "BS")

	//styleRed, _ := xlsx.NewStyle(`{"fill":{"type":"pattern","pattern":1,"fontcolor":["#FF0000"]}}`)
	styleRed, _ := xlsx.NewStyle(`{"font":{"color":"#FF0000"}}`)

	xlsx.SetPanes("Conciliation", `{"freeze":true,"split":false,"x_split":0,"y_split":1,"top_left_cell":"B1","active_pane":"topRight"}`)
	xlsx.SetCellStyle("Conciliation", "A1", "E1", styleHeader)

	// sort Conciliation
	sort.Slice(AnalisysRFS[:], func(i, j int) bool {
		return AnalisysRFS[i].AdvancementCode < AnalisysRFS[j].AdvancementCode
	})

	i = 0
	line = 2
	for i < len(AnalisysRFS) {
		if AnalisysRFS[i].AdvancementCode != 0 {
			xlsx.SetCellValue("Conciliation", "A"+strconv.Itoa(line), AnalisysRFS[i].AdvancementCode)
			xlsx.SetCellValue("Conciliation", "B"+strconv.Itoa(line), AnalisysRFS[i].DB)
			xlsx.SetCellValue("Conciliation", "C"+strconv.Itoa(line), AnalisysRFS[i].Adyen)
			xlsx.SetCellValue("Conciliation", "D"+strconv.Itoa(line), AnalisysRFS[i].BSTurbi)
			xlsx.SetCellValue("Conciliation", "E"+strconv.Itoa(line), AnalisysRFS[i].BSReal)	
			
			xlsx.SetCellStyle("Conciliation", "B"+strconv.Itoa(line), "E"+strconv.Itoa(line), styleNumber)

			if (AnalisysRFS[i].BSTurbi != AnalisysRFS[i].BSReal){
				xlsx.SetCellStyle("Conciliation", "A"+strconv.Itoa(line), "E"+strconv.Itoa(line), styleRed)
			}
			
			line++
		}
		i++
	}

	xlsx.SetColWidth("Conciliation", "A", "A", 20)
	xlsx.SetColWidth("Conciliation", "B", "E", 12)


	err := xlsx.SaveAs("./TurbiFinancial.xlsx")
	if err != nil {
		fmt.Println(err)
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

// MergeFiles ...
func MergeFiles() {
	entries, _ := ioutil.ReadDir("CSV/")

	f, _ := os.Create("merged.csv")

	defer f.Close()

	i := 0
	for i < len(entries) {
		if strings.Contains(entries[i].Name(), "report_batch") {
			name := "CSV/" + entries[i].Name()

			b, _ := ioutil.ReadFile(name)

			f.Write(b)
		}
		i++
	}
	f.Sync()

}

// GetPayOut ...
func GetPayOut() {
	entries, _ := ioutil.ReadDir("CSV/")

	i := 0

	for i < len(entries) {

		f, err := os.Open("CSV/" + entries[i].Name())
		if err != nil {
			panic(err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)

		line := 1

		var payout models.Payouts

		for scanner.Scan() {
			flagV := strings.Split(scanner.Text(), `,`)
			if flagV[4] != "" && !strings.Contains(flagV[4], "Bank=") && !strings.Contains(flagV[4], "Payment Method") {
				payout.Flag = flagV[4]
			}
			if strings.Contains(scanner.Text(), "Balancetransfer") {
				LineReadV := strings.Split(scanner.Text(), `,`)
				payout.BanlanceTransfer, _ = strconv.ParseFloat(LineReadV[15], 64)
				if payout.BanlanceTransfer == 0 {
					payout.BanlanceTransfer, _ = strconv.ParseFloat(LineReadV[14], 64)
				}
				if len(LineReadV) >= 24 {
					payout.Batch, _ = strconv.ParseInt(LineReadV[24], 10, 64)
					payout.PayoutDate = LineReadV[5]
					payout.CodeTransaction = "Balancetransfer"
					//if strings.Contains(payout.PayoutDate, ym) {
					PayoutsF = append(PayoutsF, payout)
					//}
				}
			}
			if strings.Contains(scanner.Text(), "Payout Date") {
				LineReadV := strings.Split(scanner.Text(), `,`)
				payout.PayoutDate = LineReadV[8]
				bankS := strings.Replace(LineReadV[4], "Bank=", "", 1)
				bankS = strings.Replace(bankS, " ", "", 1)
				payout.Bank, _ = strconv.ParseInt(bankS, 10, 64)
				branchS := strings.Replace(LineReadV[5], "Branch=", "", 1)
				branchS = strings.Replace(branchS, " ", "", 1)
				payout.Branch, _ = strconv.ParseInt(branchS, 10, 64)
				accountS := strings.Replace(LineReadV[6], "Account=", "", 1)
				accountS = strings.Replace(accountS, "\"", "", 1)
				accountS = strings.Replace(accountS, " ", "", 1)
				payout.Account, _ = strconv.ParseInt(accountS, 10, 64)
				payout.Value, _ = strconv.ParseFloat(LineReadV[19], 64)
				ln := strings.Split(LineReadV[11], " ")
				payout.CodeTransaction = ln[0]
				payout.CodeTransaction = strings.Replace(payout.CodeTransaction, "\"", "", 1)
				payout.CodeTransaction = strings.Replace(payout.CodeTransaction, "TX2", "", 1)
				payout.CodeTransaction = strings.Replace(payout.CodeTransaction, "XT", "", 1)
				payout.CodeTransaction = strings.Replace(payout.CodeTransaction, " ", "", 1)
				payout.Batch, _ = strconv.ParseInt(ln[2], 10, 64)
				//if strings.Contains(payout.PayoutDate, ym) {
				PayoutsF = append(PayoutsF, payout)
				//}
			}

			line++
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		i++
	}
}

// WgetCSV ...
func WgetCSV() {
	cmd := exec.Command("wget",
		"--http-user=report@Company.TurbiBR",
		"--http-password=FH4Z5sV[6a3mmk\\pH(x=Y!*u+",
		"--directory-prefix=CSV/",
		"--no-check-certificate",
		"-nc",
		"-r",
		"-l1",
		"-H",
		"-t1",
		"-nd",
		"-np",
		"-Asettle*",
		"-erobots=off",
		"https://ca-live.adyen.com/reports/download/MerchantAccount/TurbiBRCOM")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	cmd.Process.Kill()
}

// DownloadCSV ...
func DownloadCSV() {

	// AdyenUser = "report@Company.TurbiBR"
	// AdyenPass = "FH4Z5sV[6a3mmk\\pH(x=Y!*u+"
	// AdyenURL = "https://ca-live.adyen.com/reports/download/MerchantAccount/TurbiBRCOM"

	// var header = map[string]string{
	// 	"Content-Type":  "application/json",
	// 	"Authorization": "Basic " + BasicAuth(AdyenUser, AdyenPass),
	// }

	// req, err := http.NewRequest("GET", AdyenURL, bytes.NewBuffer(jsonStr))
	// req.Header.Set("Content-Type",  "application/json")
	// req.Header.Set("Authorization" ,  "Basic " + BasicAuth(AdyenUser, AdyenPass))

	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// 	return
	// }
	// defer resp.Body.Close()

	// fmt.Println("")
	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)

	// body, _ := ioutil.ReadAll(resp.Body)

}

// BasicAuth ...
func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// checkErr ...
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
