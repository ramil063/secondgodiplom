package items

import "encoding/json"

// SensitiveBankCardData чувствительная часть данных о банковских картах
// чувствительная - значит требуешь шифрования
type SensitiveBankCardData struct {
	Number          string `json:"number"`
	ValidUntilYear  int32  `json:"valid_until_year"`
	ValidUntilMonth int32  `json:"valid_until_month"`
	Cvv             int32  `json:"cvv"`
	Holder          string `json:"holder"`
}

// ToJSON функция перевода данных в json формат
func (d *SensitiveBankCardData) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

// SensitiveBankCardDataFromJSON перевод данных из формата json в формат структуры
func SensitiveBankCardDataFromJSON(data []byte) (*SensitiveBankCardData, error) {
	var result SensitiveBankCardData
	err := json.Unmarshal(data, &result)
	return &result, err
}
