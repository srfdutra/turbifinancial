package main

import (
	"fmt"
	"strings"
	"strconv"
	"sort"
	"github.com/360EntSecGroup-Skylar/excelize"
	"turbi.com.br/financial/models"
)

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
	xlsx.SetCellValue("Conciliation", "C1", "Commission Adyen")
	xlsx.SetCellValue("Conciliation", "D1", "Adyen")
	xlsx.SetCellValue("Conciliation", "E1", "Commission BS")
	xlsx.SetCellValue("Conciliation", "F1", "BS Turbi")
	xlsx.SetCellValue("Conciliation", "G1", "BS")
	xlsx.SetCellValue("Conciliation", "H1", "Advancement Date")
	xlsx.SetCellValue("Conciliation", "I1", "Advancement Pay")

	//styleRed, _ := xlsx.NewStyle(`{"fill":{"type":"pattern","pattern":1,"fontcolor":["#FF0000"]}}`)
	styleRed, _ := xlsx.NewStyle(`{"font":{"color":"#FF0000"}}`)
	styleCenter, _ := xlsx.NewStyle(`{"alignment":{"horizontal":"center"}}`)
	styleRedCenter, _ := xlsx.NewStyle(`{"alignment":{"horizontal":"center"},"font":{"color":"#FF0000"}}`)

	xlsx.SetPanes("Conciliation", `{"freeze":true,"split":false,"x_split":0,"y_split":1,"top_left_cell":"B1","active_pane":"topRight"}`)
	xlsx.SetCellStyle("Conciliation", "A1", "I1", styleHeader)

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
			xlsx.SetCellValue("Conciliation", "C"+strconv.Itoa(line), AnalisysRFS[i].CommissionAdyen)
			xlsx.SetCellValue("Conciliation", "D"+strconv.Itoa(line), AnalisysRFS[i].Adyen)
			xlsx.SetCellValue("Conciliation", "E"+strconv.Itoa(line), AnalisysRFS[i].CommissionBS)
			xlsx.SetCellValue("Conciliation", "F"+strconv.Itoa(line), AnalisysRFS[i].BSTurbi)
			xlsx.SetCellValue("Conciliation", "G"+strconv.Itoa(line), AnalisysRFS[i].BSReal)	
			xlsx.SetCellValue("Conciliation", "H"+strconv.Itoa(line), AnalisysRFS[i].DataAnt)	
			xlsx.SetCellValue("Conciliation", "I"+strconv.Itoa(line), AnalisysRFS[i].DataPag)	
			
			xlsx.SetCellStyle("Conciliation", "B"+strconv.Itoa(line), "G"+strconv.Itoa(line), styleNumber)
			xlsx.SetCellStyle("Conciliation", "H"+strconv.Itoa(line), "I"+strconv.Itoa(line), styleCenter)
			
			if (AnalisysRFS[i].BSTurbi != AnalisysRFS[i].BSReal){
				xlsx.SetCellStyle("Conciliation", "A"+strconv.Itoa(line), "I"+strconv.Itoa(line), styleRed)
				xlsx.SetCellStyle("Conciliation", "H"+strconv.Itoa(line), "I"+strconv.Itoa(line), styleRedCenter)
			}
			
			
			line++
		}
		i++
	}

	xlsx.SetColWidth("Conciliation", "A", "A", 20)
	xlsx.SetColWidth("Conciliation", "B", "B", 12)
	xlsx.SetColWidth("Conciliation", "D", "D", 12)
	xlsx.SetColWidth("Conciliation", "C", "C", 20)
	xlsx.SetColWidth("Conciliation", "E", "E", 20)
	xlsx.SetColWidth("Conciliation", "F", "G", 12)
	xlsx.SetColWidth("Conciliation", "H", "I", 22)


	err := xlsx.SaveAs("files/TurbiFinancial.xlsx")
	if err != nil {
		fmt.Println(err)
	}
}