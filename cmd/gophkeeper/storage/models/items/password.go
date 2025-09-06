package items

import "encoding/json"

type SensitivePasswordData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Target   string `json:"target"`
}

func (d *SensitivePasswordData) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

func SensitivePasswordDataFromJSON(data []byte) (*SensitivePasswordData, error) {
	var result SensitivePasswordData
	err := json.Unmarshal(data, &result)
	return &result, err
}
