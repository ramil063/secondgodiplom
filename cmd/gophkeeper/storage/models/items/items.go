package items

import (
	"time"
)

// EncryptedItem структура работы с зашифрованными данными
type EncryptedItem struct {
	UserID              int64
	Type                string
	Data                []byte
	Description         string
	EncryptionAlgorithm string
	Iv                  []byte
}

// MetaData структура для работы с метаданными
type MetaData struct {
	ID        int64
	ItemID    int64
	Name      string
	Value     string
	CreatedAt time.Time
}

// ItemData структура работы с зашифрованными данными и метаданными
type ItemData struct {
	ID                  int64
	Data                []byte
	Description         string
	CreatedAt           time.Time
	EncryptionAlgorithm string
	IV                  []byte
	MetaDataItems       []*MetaData
}
