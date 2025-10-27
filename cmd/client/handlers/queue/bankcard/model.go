package bankcard

import "time"

// Request общая структура для отправки отложенных запросов на сервер
type Request struct {
	GeneratedID     string    `json:"generatedId"`
	ID              int64     `json:"id"`
	Number          string    `json:"number"`
	ValidUntilYear  int32     `json:"valid_until_year"`
	ValidUntilMonth int32     `json:"valid_until_month"`
	Cvv             int32     `json:"cvv"`
	Holder          string    `json:"holder"`
	Description     string    `json:"description"`
	MetaDataName    string    `json:"meta_data_name"`
	MetaDataValue   string    `json:"meta_data_value"`
	CreatedAt       time.Time `json:"created_at"`
	RetryCount      int       `json:"retry_count"`
	Status          string    `json:"status"` // "pending", "processing", "completed", "failed"
}
