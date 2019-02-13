package main

import (
	"encoding/base64"
	"os"
	"os/exec"
)

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