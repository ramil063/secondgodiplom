package text

import (
	"context"
	"math"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/items"
	itemModel "github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/items"
	itemsConstants "github.com/ramil063/secondgodiplom/internal/constants/items"
	textDataPb "github.com/ramil063/secondgodiplom/internal/proto/gen/items/textdata"
	"github.com/ramil063/secondgodiplom/internal/security/crypto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	textDataPb.UnimplementedServiceServer

	storage   items.Itemer
	Encryptor crypto.Encryptor
	Decryptor crypto.Decryptor
}

func NewServer(storage items.Itemer, encryptor crypto.Encryptor, decryptor crypto.Decryptor) *Server {
	return &Server{
		storage:   storage,
		Encryptor: encryptor,
		Decryptor: decryptor,
	}
}

func (s *Server) CreateTextData(ctx context.Context, req *textDataPb.CreateTextDataRequest) (*textDataPb.TextDataItem, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication")
	}

	// 3. Шифруем всю структуру
	encryptedData, algorithm, iv, err := s.Encryptor.Encrypt([]byte(req.TextData))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to encrypt data")
	}

	// 4. Сохраняем в основную таблицу
	itemID, err := s.storage.SaveEncryptedData(ctx, &itemModel.EncryptedItem{
		UserID:              int64(userID),
		Type:                itemsConstants.TypePasswords,
		Data:                encryptedData,
		Description:         req.Description,
		EncryptionAlgorithm: algorithm,
		Iv:                  iv,
	})

	// 4. Сохраняем метаданные в отдельную таблицу
	err = s.storage.SaveMetadata(ctx, &itemModel.MetaData{
		ItemID: itemID,
		Name:   req.MetaDataName,
		Value:  req.MetaDataValue,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to save meta data")
	}

	// 5. Возвращаем ответ
	return &textDataPb.TextDataItem{
		Id:          itemID,
		TextData:    req.TextData,
		Description: req.Description,
		MetaData: []*textDataPb.MetaData{
			{
				Name:  req.MetaDataName,
				Value: req.MetaDataValue,
			},
		},
	}, nil
}

func (s *Server) ListTextDataItems(
	ctx context.Context,
	req *textDataPb.ListTextDataRequest,
) (*textDataPb.ListTextDataResponse, error) {
	// Извлекаем userID из контекста
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication")
	}

	// Валидация параметров
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 {
		req.PerPage = 10 // Значение по умолчанию
	}

	// Получаем данные из репозитория
	passwordsList, totalCount, err := s.storage.GetListItems(
		ctx, int64(userID), req.Page, req.PerPage, itemsConstants.TypePasswords, req.Filter,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list text data")
	}

	// Конвертируем в proto-сообщения
	var pbPasswords []*textDataPb.TextDataItem
	var pbMetaData []*textDataPb.MetaData

	for i := range passwordsList {
		p := passwordsList[i]

		decryptedData, err := s.Decryptor.Decrypt(p.Data, p.IV)
		if err != nil {
			continue
		}

		for _, val := range p.MetaDataItems {
			pbMetaDataItem := textDataPb.MetaData{
				Id:    val.ID,
				Name:  val.Name,
				Value: val.Value,
			}
			pbMetaData = append(pbMetaData, &pbMetaDataItem)
		}

		pbPasswords = append(pbPasswords, &textDataPb.TextDataItem{
			Id:          p.ID,
			TextData:    string(decryptedData),
			Description: p.Description,
			CreatedAt:   p.CreatedAt.String(),
			MetaData:    pbMetaData,
		})
	}
	totalPages := int32(math.Ceil(float64(totalCount) / float64(req.PerPage)))

	return &textDataPb.ListTextDataResponse{
		TextDataItems: pbPasswords,
		TotalCount:    totalCount,
		TotalPages:    totalPages,
		CurrentPage:   req.Page,
	}, nil
}

func (s *Server) GetTextData(ctx context.Context, req *textDataPb.GetTextDataRequest) (*textDataPb.TextDataItem, error) {
	// Извлекаем userID из контекста
	if _, ok := ctx.Value("userID").(int); !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication")
	}

	// Получаем данные из репозитория
	password, err := s.storage.GetItem(ctx, req.Id)
	if err != nil || password == nil {
		return nil, status.Error(codes.Internal, "failed to get password")
	}

	decryptedData, err := s.Decryptor.Decrypt(password.Data, password.IV)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to decrypt password")
	}

	var metaDataList []*textDataPb.MetaData

	for _, val := range password.MetaDataItems {
		metaDataList = append(metaDataList, &textDataPb.MetaData{
			Id:    val.ID,
			Name:  val.Name,
			Value: val.Value,
		})
	}

	return &textDataPb.TextDataItem{
		Id:          password.ID,
		TextData:    string(decryptedData),
		Description: password.Description,
		CreatedAt:   password.CreatedAt.String(),
		MetaData:    metaDataList,
	}, nil
}

func (s *Server) DeleteTextData(ctx context.Context, req *textDataPb.DeleteTextDataRequest) (*empty.Empty, error) {
	// Извлекаем userID из контекста
	if _, ok := ctx.Value("userID").(int); !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication")
	}
	err := s.storage.DeleteItem(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete password")
	}
	return &empty.Empty{}, nil
}

func (s *Server) UpdateTextData(
	ctx context.Context,
	req *textDataPb.UpdateTextDataRequest,
) (*textDataPb.TextDataItem, error) {
	if _, ok := ctx.Value("userID").(int); !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication")
	}
	password, err := s.storage.GetItem(ctx, req.Id)
	if err != nil || password == nil {
		return nil, status.Error(codes.Internal, "failed to get password")
	}

	// 3. Шифруем всю структуру
	encryptedData, algorithm, iv, err := s.Encryptor.Encrypt([]byte(req.TextData))
	if err != nil {
		return nil, err
	}

	// 4. Сохраняем в основную таблицу
	itemID, err := s.storage.UpdateItem(ctx, req.Id, &itemModel.EncryptedItem{
		Data:                encryptedData,
		Description:         req.Description,
		EncryptionAlgorithm: algorithm,
		Iv:                  iv,
	})

	// 5. Возвращаем ответ
	return &textDataPb.TextDataItem{
		Id:          itemID,
		Description: req.Description,
	}, nil
}
