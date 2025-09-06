package items

import (
	"context"

	itemModel "github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/items"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/items"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
)

type Itemer interface {
	SaveEncryptedData(ctx context.Context, encryptedPassword *itemModel.EncryptedItem) (int64, error)
	SaveMetadata(ctx context.Context, metadata *itemModel.MetaData) error
	GetListItems(ctx context.Context, userID int64, page int32, perPage int32, itemType, filter string) ([]*itemModel.ItemData, int32, error)
	GetItem(ctx context.Context, passwordID int64) (*itemModel.ItemData, error)
	DeleteItem(ctx context.Context, passwordID int64) error
	UpdateItem(ctx context.Context, itemId int64, encryptedPassword *itemModel.EncryptedItem) (int64, error)
	GetMetaDataList(ctx context.Context, itemId int64) ([]*itemModel.MetaData, error)
}

func NewStorage(rep repository.Repository) Itemer {
	return &items.Item{
		Repository: &rep,
	}
}
