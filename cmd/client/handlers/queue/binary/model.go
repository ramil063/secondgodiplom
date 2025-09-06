package binary

import "time"

type Request struct {
	GeneratedID   string    `json:"generated_id"`
	ID            int64     `json:"id"`
	Description   string    `json:"description"`
	MetaDataName  string    `json:"meta_data_name"`
	MetaDataValue string    `json:"meta_data_value"`
	CreatedAt     time.Time `json:"created_at"`
	RetryCount    int       `json:"retry_count"`
	Status        string    `json:"status"` // "pending", "processing", "completed", "failed"
}
