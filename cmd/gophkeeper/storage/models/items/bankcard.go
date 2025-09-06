package items

import "encoding/json"

type SensitiveBankCardData struct {
	Number          string `json:"number"`
	ValidUntilYear  int32  `json:"valid_until_year"`
	ValidUntilMonth int32  `json:"valid_until_month"`
	Cvv             int32  `json:"cvv"`
	Holder          string `json:"holder"`
}

func (d *SensitiveBankCardData) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

func SensitiveBankCardDataFromJSON(data []byte) (*SensitiveBankCardData, error) {
	var result SensitiveBankCardData
	err := json.Unmarshal(data, &result)
	return &result, err
}
