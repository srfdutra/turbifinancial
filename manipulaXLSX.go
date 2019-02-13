package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"turbi.com.br/financial/models"
)

// CreateDetails ...
func CreateDetails() (xlsx []models.XLSX, sumXLSX models.SumXLSX, dbValues []models.DBValues) {

	dbValues = DirtyStruct()

	filename := "files/merged.csv"

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

// ReadBS ...
func ReadBS(advancementBatch string) (commissionBS float64) {
	xlsx, err := excelize.OpenFile("files/LotesAntecipados.xlsx")
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
func ReadBSOperations(batchCod string) (valueBS float64, da string, dp string) {
	xlsx, err := excelize.OpenFile("files/OperacoesPedidoExterno.xlsx")
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
			dp = rows[x][4]
			da = rows[x][3]
			return valueBS, da, dp
		}
		x++
	}
	return valueBS, da, dp
}
