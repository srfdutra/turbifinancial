package models

// XLSX ...
type XLSX struct {
	Batch            string
	PSPModification  string
	PSPReference     string
	DateTransaction  string
	Advancement      float64
	AdvancementBatch string
	AdvancementCode  int64
	GrossCredit      float64
	GrossDebit       float64
	NetCredit        float64
	NetDebit         float64
	Commission       float64
	DBvalue          float64
	Flag             string
}

// AnalysisLot ...
type AnalysisLot struct {
	AdvancementCode int64
	DB              float64
	CommissionAdyen float64
	Adyen           float64
	CommissionBS    float64
	BSTurbi         float64	
	BSReal          float64
	DataAnt         string
	DataPag         string
}

// SumXLSX ...
type SumXLSX struct {
	Gross      float64
	Commission float64
	Net        float64
	GrossDB    float64
}

// DBValues ...
type DBValues struct {
	PSP          string
	State        string
	ResponseCode string
	Verb         string
	Value        float64
}

// DBCheckPSP ...
type DBCheckPSP struct {
	PSPModification string
	PSPReference    string
	Value           float64
	Type            string
}

// Payouts ...
type Payouts struct {
	PayoutDate       string
	Bank             int64
	Branch           int64
	Account          int64
	CodeTransaction  string
	Batch            int64
	Value            float64
	BanlanceTransfer float64
	Flag             string
}

// Transaction ...
type Transaction struct {
	ID            int64
	TransactionID string
	BookingID     int64
	State         string
	Token         string
	Verb          string
}
