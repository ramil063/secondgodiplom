package items

import "encoding/json"

// SensitivePasswordData чувствительные данные данных о пароле
type SensitivePasswordData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Target   string `json:"target"`
}

// ToJSON конвертация структуры в json формат
func (d *SensitivePasswordData) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

// SensitivePasswordDataFromJSON конвертация json строки в структуру
func SensitivePasswordDataFromJSON(data []byte) (*SensitivePasswordData, error) {
	var result SensitivePasswordData
	err := json.Unmarshal(data, &result)
	return &result, err
}
