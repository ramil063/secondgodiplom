package items

import (
	"time"
)

type EncryptedItem struct {
	UserID              int64
	Type                string
	Data                []byte
	Description         string
	EncryptionAlgorithm string
	Iv                  []byte
}

type MetaData struct {
	ID        int64
	ItemID    int64
	Name      string
	Value     string
	CreatedAt time.Time
}

type ItemData struct {
	ID                  int64
	Data                []byte
	Description         string
	CreatedAt           time.Time
	EncryptionAlgorithm string
	IV                  []byte
	MetaDataItems       []*MetaData
}
