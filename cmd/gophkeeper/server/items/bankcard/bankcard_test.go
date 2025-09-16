package bankcard

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/items"
	itemsMock "github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/items/mocks"
	itemModel "github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/items"
	itemsConstants "github.com/ramil063/secondgodiplom/internal/constants/items"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/bankcard"
	"github.com/ramil063/secondgodiplom/internal/security/crypto"
	cryptoMock "github.com/ramil063/secondgodiplom/internal/security/crypto/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := itemsMock.NewMockItemer(ctrl)
	encryptorMock := cryptoMock.NewMockEncryptor(ctrl)
	decryptorMock := cryptoMock.NewMockDecryptor(ctrl)

	type args struct {
		storage   items.Itemer
		encryptor crypto.Encryptor
		decryptor crypto.Decryptor
	}
	tests := []struct {
		name string
		args args
		want *Server
	}{
		{
			name: "TestNewServer",
			args: args{
				storage:   storageMock,
				encryptor: encryptorMock,
				decryptor: decryptorMock,
			},
			want: &Server{
				storage:   storageMock,
				Encryptor: encryptorMock,
				Decryptor: decryptorMock,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewServer(tt.args.storage, tt.args.encryptor, tt.args.decryptor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_CreateCardData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := itemsMock.NewMockItemer(ctrl)
	encryptorMock := cryptoMock.NewMockEncryptor(ctrl)
	decryptorMock := cryptoMock.NewMockDecryptor(ctrl)

	type args struct {
		ctx context.Context
		req *bankcard.CreateCardDataRequest
	}
	tests := []struct {
		name          string
		userID        int
		itemID        int64
		encryptedData []byte
		algorithm     string
		iv            []byte
		args          args
		want          *bankcard.CardDataItem
	}{
		{
			name:          "TestCreateCardData",
			encryptedData: []byte("test"),
			algorithm:     "aes-256-gcm",
			iv:            []byte("test"),
			userID:        1,
			itemID:        1,
			args: args{
				ctx: context.WithValue(context.Background(), "userID", 1),
				req: &bankcard.CreateCardDataRequest{
					Number:          "1234",
					ValidUntilYear:  1,
					ValidUntilMonth: 1,
					Cvv:             1,
					Holder:          "test",
					Description:     "test",
					MetaDataName:    "test",
					MetaDataValue:   "test",
				},
			},
			want: &bankcard.CardDataItem{
				Id:              1,
				Number:          "1234",
				ValidUntilYear:  1,
				ValidUntilMonth: 1,
				Cvv:             1,
				Holder:          "test",
				Description:     "test",
				MetaData: []*bankcard.MetaData{
					{
						Name:  "test",
						Value: "test",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				storage:   storageMock,
				Encryptor: encryptorMock,
				Decryptor: decryptorMock,
			}

			encryptorMock.EXPECT().
				Encrypt(gomock.Any()).
				Return(tt.encryptedData, tt.algorithm, tt.iv, nil)

			storageMock.EXPECT().SaveEncryptedData(tt.args.ctx, &itemModel.EncryptedItem{
				UserID:              int64(tt.userID),
				Type:                itemsConstants.TypeCard,
				Data:                tt.encryptedData,
				Description:         tt.args.req.Description,
				EncryptionAlgorithm: tt.algorithm,
				Iv:                  tt.iv,
			}).Return(tt.itemID, nil)

			storageMock.EXPECT().SaveMetadata(tt.args.ctx, &itemModel.MetaData{
				ItemID: tt.itemID,
				Name:   tt.args.req.MetaDataName,
				Value:  tt.args.req.MetaDataValue,
			}).Return(nil)

			got, err := s.CreateCardData(tt.args.ctx, tt.args.req)
			assert.NoError(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateCardData() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_DeleteCardData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := itemsMock.NewMockItemer(ctrl)
	encryptorMock := cryptoMock.NewMockEncryptor(ctrl)
	decryptorMock := cryptoMock.NewMockDecryptor(ctrl)

	type args struct {
		ctx context.Context
		req *bankcard.DeleteCardDataRequest
	}
	tests := []struct {
		name string
		args args
		want *empty.Empty
	}{
		{
			name: "TestDeleteCardData",
			args: args{
				ctx: context.WithValue(context.Background(), "userID", 1),
				req: &bankcard.DeleteCardDataRequest{
					Id: 1,
				},
			},
			want: &empty.Empty{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				storage:   storageMock,
				Encryptor: encryptorMock,
				Decryptor: decryptorMock,
			}
			storageMock.EXPECT().DeleteItem(tt.args.ctx, tt.args.req.Id).Return(nil)
			got, err := s.DeleteCardData(tt.args.ctx, tt.args.req)
			assert.NoError(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteCardData() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_GetCardData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := itemsMock.NewMockItemer(ctrl)
	encryptorMock := cryptoMock.NewMockEncryptor(ctrl)
	decryptorMock := cryptoMock.NewMockDecryptor(ctrl)

	timeStr := "2025-09-16 06:29:40.129907335 +0300 MSK"
	parsedTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", timeStr)
	assert.NoError(t, err)

	type args struct {
		ctx context.Context
		req *bankcard.GetCardDataRequest
	}
	tests := []struct {
		name     string
		args     args
		itemData *itemModel.ItemData
		sbcData  *itemModel.SensitiveBankCardData
		want     *bankcard.CardDataItem
	}{
		{
			name: "TestGetCardData",
			args: args{
				ctx: context.WithValue(context.Background(), "userID", 1),
				req: &bankcard.GetCardDataRequest{
					Id: 1,
				},
			},
			itemData: &itemModel.ItemData{
				ID:            1,
				Data:          []byte("test"),
				Description:   "test",
				CreatedAt:     parsedTime,
				MetaDataItems: []*itemModel.MetaData{},
			},
			sbcData: &itemModel.SensitiveBankCardData{
				Number:          "1234",
				ValidUntilYear:  1,
				ValidUntilMonth: 1,
				Cvv:             1,
				Holder:          "test",
			},
			want: &bankcard.CardDataItem{
				Id:              1,
				Number:          "1234",
				ValidUntilYear:  1,
				ValidUntilMonth: 1,
				Cvv:             1,
				Holder:          "test",
				Description:     "test",
				CreatedAt:       timeStr,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				storage:   storageMock,
				Encryptor: encryptorMock,
				Decryptor: decryptorMock,
			}

			storageMock.EXPECT().
				GetItem(tt.args.ctx, tt.args.req.Id).
				Return(tt.itemData, nil)

			dData, err := tt.sbcData.ToJSON()
			assert.NoError(t, err)

			decryptorMock.EXPECT().
				Decrypt(tt.itemData.Data, tt.itemData.IV).
				Return(dData, nil)

			got, err := s.GetCardData(tt.args.ctx, tt.args.req)
			assert.NoError(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCardData() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_ListCardsData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := itemsMock.NewMockItemer(ctrl)
	encryptorMock := cryptoMock.NewMockEncryptor(ctrl)
	decryptorMock := cryptoMock.NewMockDecryptor(ctrl)

	timeStr := "2025-09-16 06:29:40.129907335 +0300 MSK"
	parsedTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", timeStr)
	assert.NoError(t, err)

	type args struct {
		ctx context.Context
		req *bankcard.ListCardsDataRequest
	}
	tests := []struct {
		name     string
		userID   int
		args     args
		itemData *itemModel.ItemData
		sbcData  *itemModel.SensitiveBankCardData
		want     *bankcard.ListCardsDataResponse
	}{
		{
			name:   "TestListCardsData",
			userID: 1,
			args: args{
				ctx: context.WithValue(context.Background(), "userID", 1),
				req: &bankcard.ListCardsDataRequest{
					Page:    1,
					PerPage: 10,
					Filter:  "",
				},
			},
			itemData: &itemModel.ItemData{
				ID:            1,
				Data:          []byte("test"),
				Description:   "test",
				CreatedAt:     parsedTime,
				MetaDataItems: []*itemModel.MetaData{},
			},
			sbcData: &itemModel.SensitiveBankCardData{
				Number:          "1234",
				ValidUntilYear:  1,
				ValidUntilMonth: 1,
				Cvv:             1,
				Holder:          "test",
			},
			want: &bankcard.ListCardsDataResponse{
				Cards: []*bankcard.CardDataItem{
					{
						Id:              1,
						Number:          "1234",
						ValidUntilYear:  1,
						ValidUntilMonth: 1,
						Cvv:             1,
						Holder:          "test",
						Description:     "test",
						CreatedAt:       timeStr,
					},
				},
				TotalCount:  1,
				TotalPages:  1,
				CurrentPage: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				storage:   storageMock,
				Encryptor: encryptorMock,
				Decryptor: decryptorMock,
			}
			storageMock.EXPECT().
				GetListItems(
					tt.args.ctx,
					int64(tt.userID),
					tt.args.req.Page,
					tt.args.req.PerPage,
					itemsConstants.TypeCard,
					tt.args.req.Filter,
				).
				Return(
					[]*itemModel.ItemData{
						tt.itemData,
					},
					tt.want.TotalCount,
					nil)

			dData, err := tt.sbcData.ToJSON()
			assert.NoError(t, err)
			decryptorMock.EXPECT().
				Decrypt(tt.itemData.Data, tt.itemData.IV).
				Return(dData, nil)

			got, err := s.ListCardsData(tt.args.ctx, tt.args.req)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Cards, got.Cards)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListCardsData() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_UpdateCardData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := itemsMock.NewMockItemer(ctrl)
	encryptorMock := cryptoMock.NewMockEncryptor(ctrl)
	decryptorMock := cryptoMock.NewMockDecryptor(ctrl)

	timeStr := "2025-09-16 06:29:40.129907335 +0300 MSK"
	parsedTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", timeStr)
	assert.NotEmpty(t, parsedTime)
	assert.NoError(t, err)

	type args struct {
		ctx context.Context
		req *bankcard.UpdateCardDataRequest
	}
	tests := []struct {
		name          string
		args          args
		encryptedData []byte
		algorithm     string
		iv            []byte
		itemData      *itemModel.ItemData
		sbcData       *itemModel.SensitiveBankCardData
		itemID        int64
		want          *bankcard.CardDataItem
	}{
		{
			name:      "TestUpdateCardData",
			algorithm: "aes-256-gcm",
			iv:        []byte("test"),
			args: args{
				ctx: context.WithValue(context.Background(), "userID", 1),
				req: &bankcard.UpdateCardDataRequest{
					Id:              1,
					Number:          "1234",
					ValidUntilYear:  2000,
					ValidUntilMonth: 1,
					Cvv:             1,
					Holder:          "test",
					Description:     "test",
				},
			},
			itemData: &itemModel.ItemData{
				ID:            1,
				Data:          []byte("test"),
				Description:   "test",
				CreatedAt:     parsedTime,
				MetaDataItems: []*itemModel.MetaData{},
			},
			sbcData: &itemModel.SensitiveBankCardData{
				Number:          "1234",
				ValidUntilYear:  1,
				ValidUntilMonth: 1,
				Cvv:             1,
				Holder:          "test",
			},
			itemID: 1,
			want: &bankcard.CardDataItem{
				Id:          1,
				Description: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				storage:   storageMock,
				Encryptor: encryptorMock,
				Decryptor: decryptorMock,
			}

			storageMock.
				EXPECT().
				GetItem(tt.args.ctx, tt.args.req.Id).
				Return(tt.itemData, nil)

			sensitiveData := &itemModel.SensitiveBankCardData{
				Number:          tt.args.req.Number,
				ValidUntilYear:  tt.args.req.ValidUntilYear,
				ValidUntilMonth: tt.args.req.ValidUntilMonth,
				Cvv:             tt.args.req.Cvv,
				Holder:          tt.args.req.Holder,
			}
			jsonData, err := sensitiveData.ToJSON()
			assert.NoError(t, err)
			encryptorMock.
				EXPECT().
				Encrypt(jsonData).
				Return(tt.itemData.Data, tt.algorithm, tt.iv, nil)

			storageMock.EXPECT().
				UpdateItem(tt.args.ctx, tt.args.req.Id, &itemModel.EncryptedItem{
					Data:                tt.itemData.Data,
					Description:         tt.args.req.Description,
					EncryptionAlgorithm: tt.algorithm,
					Iv:                  tt.iv,
				}).
				Return(tt.itemID, nil)

			got, err := s.UpdateCardData(tt.args.ctx, tt.args.req)
			assert.NoError(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateCardData() got = %v, want %v", got, tt.want)
			}
		})
	}
}
