package items

import "time"

type FileInfo struct {
	ID            int64     `json:"id"`
	Filename      string    `json:"filename"`
	MimeType      string    `json:"mime_type"`
	OriginalSize  int64     `json:"original_size"`
	Description   string    `json:"description"`
	ChunkSize     int32     `json:"chunk_size"`
	TotalChunks   int32     `json:"total_chunks"`
	CreatedAt     time.Time `json:"created_at"`
	MetaDataItems []*MetaData
}

type ChunkData struct {
	ID                  int64     `json:"id"`
	FileId              int64     `json:"file_id"`
	ChunkIndex          int32     `json:"chunk_index"`
	EncryptedData       []byte    `json:"encrypted_data"`
	EncryptionAlgorithm string    `json:"encryption_algorithm"`
	IV                  []byte    `json:"iv"`
	CreatedAt           time.Time `json:"created_at"`
}
