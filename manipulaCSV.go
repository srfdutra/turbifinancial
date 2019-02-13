package main

import (
	"strconv"
	"turbi.com.br/financial/models"
	"bufio"
	"strings"
	"os"
	"io/ioutil"
)

// MergeFiles ...
func MergeFiles() {
	entries, _ := ioutil.ReadDir("CSV/")

	f, _ := os.Create("files/merged.csv")

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
