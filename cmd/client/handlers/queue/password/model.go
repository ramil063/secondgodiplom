package password

import "time"

type Request struct {
	ID            string    `json:"id"`
	Login         string    `json:"login"`
	Password      string    `json:"password"`
	Target        string    `json:"target"`
	Description   string    `json:"description"`
	MetaDataName  string    `json:"meta_data_name"`
	MetaDataValue string    `json:"meta_data_value"`
	CreatedAt     time.Time `json:"created_at"`
	RetryCount    int       `json:"retry_count"`
	Status        string    `json:"status"` // "pending", "processing", "completed", "failed"
}
