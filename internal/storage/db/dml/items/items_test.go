package items

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgconn"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"

	itemModel "github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/items"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/mock"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
	repository2 "github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository/mocks"
)

func TestItem_DeleteItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx    context.Context
		itemID int64
	}
	tests := []struct {
		name  string
		query string
		args  args
	}{
		{
			name:  "success",
			query: "UPDATE encrypted_item SET is_deleted=TRUE WHERE id = $1",
			args: args{
				ctx:    context.Background(),
				itemID: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)

			pi := &Item{
				Repository: &repository.Repository{Pool: poolMock},
			}

			expectedCommandTag := pgconn.CommandTag("UPDATE 0 1")
			poolMock.EXPECT().
				Exec(
					tt.args.ctx,
					tt.query,
					tt.args.itemID).
				Return(expectedCommandTag, nil)
			err := pi.DeleteItem(tt.args.ctx, tt.args.itemID)
			assert.NoError(t, err)
		})
	}
}

func TestItem_GetItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx    context.Context
		itemID int64
	}
	tests := []struct {
		name         string
		args         args
		want         *itemModel.ItemData
		metaDataJSON []byte
	}{
		{
			name: "test 1",
			args: args{
				ctx:    context.Background(),
				itemID: 1,
			},
			want:         &itemModel.ItemData{},
			metaDataJSON: []byte(`[]`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)

			pi := &Item{
				Repository: &repository.Repository{Pool: poolMock},
			}
			var metadata []*itemModel.MetaData
			err := json.Unmarshal(tt.metaDataJSON, &metadata)
			assert.NoError(t, err)
			tt.want.MetaDataItems = metadata

			poolMock.EXPECT().
				QueryRow(
					tt.args.ctx,
					gomock.Any(),
					tt.args.itemID).
				Return(&mock.Row{
					Values: []interface{}{
						tt.want.ID,
						tt.want.Data,
						tt.want.Description,
						tt.want.CreatedAt,
						tt.want.EncryptionAlgorithm,
						tt.want.IV,
						tt.metaDataJSON,
					},
				})

			got, err := pi.GetItem(tt.args.ctx, tt.args.itemID)
			assert.NoError(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetItem() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItem_GetListItems(t *testing.T) {
	type args struct {
		ctx      context.Context
		userID   int64
		page     int32
		perPage  int32
		itemType string
		filter   string
	}
	tests := []struct {
		name         string
		args         args
		want         []*itemModel.ItemData
		want1        int32
		metadataJSON []byte
	}{
		{
			name: "test 1",
			args: args{
				ctx:      context.Background(),
				userID:   1,
				page:     1,
				perPage:  1,
				itemType: "list",
				filter:   "",
			},
			want: []*itemModel.ItemData{
				{
					ID:                  1,
					Data:                []byte("1"),
					Description:         "",
					CreatedAt:           time.Now(),
					EncryptionAlgorithm: "AES-256-GCM",
					IV:                  []byte("1"),
					MetaDataItems:       []*itemModel.MetaData{},
				},
			},
			want1:        2,
			metadataJSON: []byte(`[]`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock, err := pgxmock.NewPool()
			pi := &Item{
				Repository: &repository.Repository{Pool: poolMock},
			}
			err = json.Unmarshal(tt.metadataJSON, &tt.want[0].MetaDataItems)
			assert.NoError(t, err)

			rows := poolMock.NewRows([]string{
				"id",
				"encrypted_data",
				"description",
				"created_at",
				"encryption_algorithm",
				"iv",
				"metadata",
			}).AddRow(
				tt.want[0].ID,
				tt.want[0].Data,
				tt.want[0].Description,
				tt.want[0].CreatedAt,
				tt.want[0].EncryptionAlgorithm,
				tt.want[0].IV,
				tt.metadataJSON,
			)

			rows1 := poolMock.NewRows([]string{
				"count",
			}).AddRow(
				tt.want1,
			)

			poolMock.ExpectQuery("SELECT.*ei.id.*ei.encrypted_data.*").
				WithArgs(
					tt.args.userID,
					tt.args.itemType,
					tt.args.perPage,
					(tt.args.page-1)*tt.args.perPage).
				WillReturnRows(rows)
			poolMock.ExpectQuery("SELECT.*COUNT.*").
				WithArgs(tt.args.userID, tt.args.itemType).
				WillReturnRows(rows1)

			got, got1, err := pi.GetListItems(tt.args.ctx, tt.args.userID, tt.args.page, tt.args.perPage, tt.args.itemType, tt.args.filter)

			assert.NoError(t, err)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetListItems() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetListItems() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestItem_GetMetaDataList(t *testing.T) {
	type args struct {
		ctx    context.Context
		itemId int64
	}
	tests := []struct {
		name string
		args args
		want []*itemModel.MetaData
	}{
		{
			name: "test 1",
			args: args{
				ctx:    context.Background(),
				itemId: 1,
			},
			want: []*itemModel.MetaData{
				{
					ID:        1,
					Name:      "name",
					Value:     "value",
					CreatedAt: time.Now(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock, err := pgxmock.NewPool()
			pi := &Item{
				Repository: &repository.Repository{Pool: poolMock},
			}

			rows := poolMock.NewRows([]string{
				"id",
				"name",
				"value",
				"created_at",
			}).AddRow(
				tt.want[0].ID,
				tt.want[0].Name,
				tt.want[0].Value,
				tt.want[0].CreatedAt,
			)

			poolMock.ExpectQuery("SELECT.*id.*item_metadata.*").
				WithArgs(tt.args.itemId).
				WillReturnRows(rows)

			got, err := pi.GetMetaDataList(tt.args.ctx, tt.args.itemId)
			assert.NoError(t, err)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMetaDataList() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItem_GetTotalCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx      context.Context
		query    string
		userID   int64
		itemType string
		filter   string
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			name: "test 1",
			args: args{
				ctx:      context.Background(),
				userID:   1,
				itemType: "list",
				filter:   "",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)
			pi := &Item{
				Repository: &repository.Repository{Pool: poolMock},
			}

			poolMock.EXPECT().
				QueryRow(
					tt.args.ctx,
					gomock.Any(),
					tt.args.userID,
				).
				Return(&mock.Row{
					Values: []interface{}{
						tt.want,
					},
				})
			got, err := pi.GetTotalCount(tt.args.ctx, tt.args.query, tt.args.userID, tt.args.itemType, tt.args.filter)
			assert.NoError(t, err)
			if got != tt.want {
				t.Errorf("GetTotalCount() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItem_SaveEncryptedData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx           context.Context
		encryptedItem *itemModel.EncryptedItem
	}
	tests := []struct {
		name       string
		args       args
		wantTypeId int64
		want       int64
	}{
		{
			name: "test 1",
			args: args{
				ctx: context.Background(),
				encryptedItem: &itemModel.EncryptedItem{
					UserID:              1,
					Type:                "password",
					Data:                []byte("password"),
					Description:         "123",
					EncryptionAlgorithm: "AES-256-GCM",
					Iv:                  []byte("iv"),
				},
			},
			wantTypeId: 1,
			want:       1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)
			pi := &Item{
				Repository: &repository.Repository{Pool: poolMock},
			}

			poolMock.EXPECT().
				QueryRow(
					tt.args.ctx,
					gomock.Any(),
					tt.args.encryptedItem.Type,
				).
				Return(&mock.Row{
					Values: []interface{}{
						tt.wantTypeId,
					},
				})
			poolMock.EXPECT().
				QueryRow(
					tt.args.ctx,
					gomock.Any(),
					tt.args.encryptedItem.Data,
					tt.args.encryptedItem.Description,
					tt.args.encryptedItem.UserID,
					tt.wantTypeId,
					tt.args.encryptedItem.EncryptionAlgorithm,
					tt.args.encryptedItem.Iv,
				).
				Return(&mock.Row{
					Values: []interface{}{
						tt.want,
					},
				})
			got, err := pi.SaveEncryptedData(tt.args.ctx, tt.args.encryptedItem)
			assert.NoError(t, err)
			if got != tt.want {
				t.Errorf("SaveEncryptedData() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItem_SaveMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx      context.Context
		metadata *itemModel.MetaData
	}
	tests := []struct {
		name  string
		query string
		args  args
	}{
		{
			name:  "test 1",
			query: `INSERT INTO item_metadata (item_id, name, value) VALUES ($1, $2, $3)`,
			args: args{
				ctx: context.Background(),
				metadata: &itemModel.MetaData{
					ItemID: 1,
					Name:   "name",
					Value:  "value",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)
			pi := &Item{
				Repository: &repository.Repository{Pool: poolMock},
			}

			expectedCommandTag := pgconn.CommandTag("INSERT 0 1")
			poolMock.EXPECT().
				Exec(
					tt.args.ctx,
					tt.query,
					tt.args.metadata.ItemID,
					tt.args.metadata.Name,
					tt.args.metadata.Value).
				Return(expectedCommandTag, nil)
			err := pi.SaveMetadata(tt.args.ctx, tt.args.metadata)
			assert.NoError(t, err)
		})
	}
}

func TestItem_UpdateItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx           context.Context
		itemId        int64
		encryptedItem *itemModel.EncryptedItem
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "test 1",
			args: args{
				ctx:    context.Background(),
				itemId: 1,
				encryptedItem: &itemModel.EncryptedItem{
					UserID:              1,
					Type:                "password",
					Data:                []byte("password"),
					Description:         "123",
					EncryptionAlgorithm: "AES256-GCM",
					Iv:                  []byte("iv"),
				},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)
			pi := &Item{
				Repository: &repository.Repository{Pool: poolMock},
			}

			poolMock.EXPECT().
				QueryRow(
					tt.args.ctx,
					gomock.Any(),
					tt.args.itemId,
					tt.args.encryptedItem.Data,
					tt.args.encryptedItem.Description,
					tt.args.encryptedItem.EncryptionAlgorithm,
					tt.args.encryptedItem.Iv,
				).
				Return(&mock.Row{
					Values: []interface{}{
						tt.want,
					},
				})
			got, err := pi.UpdateItem(tt.args.ctx, tt.args.itemId, tt.args.encryptedItem)
			assert.NoError(t, err)
			if got != tt.want {
				t.Errorf("UpdateItem() got = %v, want %v", got, tt.want)
			}
		})
	}
}
