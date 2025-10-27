package bankcard

import (
	"context"
	"math"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/items"
	itemModel "github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/items"
	itemsConstants "github.com/ramil063/secondgodiplom/internal/constants/items"
	bankcardsPb "github.com/ramil063/secondgodiplom/internal/proto/gen/items/bankcard"
	"github.com/ramil063/secondgodiplom/internal/security/crypto"
)

// Server надстройка над стандартным gRPC сервером(логика работы с банковскими картами)
// шифрование/дешифровка/хранение данных
type Server struct {
	bankcardsPb.UnimplementedServiceServer

	storage   items.Itemer
	Encryptor crypto.Encryptor
	Decryptor crypto.Decryptor
}

// NewServer инициализация сервера, шифровальщика, дешифровщика и структуры для работы с хранилищем
func NewServer(storage items.Itemer, encryptor crypto.Encryptor, decryptor crypto.Decryptor) *Server {
	return &Server{
		storage:   storage,
		Encryptor: encryptor,
		Decryptor: decryptor,
	}
}

// CreateCardData создание записи о банковской карте
// так же сохранение метаданных о ней
// все специфичные данные шифруются
func (s *Server) CreateCardData(ctx context.Context, req *bankcardsPb.CreateCardDataRequest) (*bankcardsPb.CardDataItem, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication")
	}
	// 1. Создаем структуру для шифрования
	sensitiveData := &itemModel.SensitiveBankCardData{
		Number:          req.Number,
		ValidUntilYear:  req.ValidUntilYear,
		ValidUntilMonth: req.ValidUntilMonth,
		Cvv:             req.Cvv,
		Holder:          req.Holder,
	}

	// 2. Сериализуем в JSON
	jsonData, err := sensitiveData.ToJSON()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to serialize data")
	}

	// 3. Шифруем всю структуру
	encryptedData, algorithm, iv, err := s.Encryptor.Encrypt(jsonData)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to encrypt data")
	}

	// 4. Сохраняем в основную таблицу
	itemID, err := s.storage.SaveEncryptedData(ctx, &itemModel.EncryptedItem{
		UserID:              int64(userID),
		Type:                itemsConstants.TypeCard,
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
	return &bankcardsPb.CardDataItem{
		Id:              itemID,
		Number:          req.Number,
		ValidUntilYear:  req.ValidUntilYear,
		ValidUntilMonth: req.ValidUntilMonth,
		Cvv:             req.Cvv,
		Holder:          req.Holder,
		Description:     req.Description,
		MetaData: []*bankcardsPb.MetaData{
			{
				Name:  req.MetaDataName,
				Value: req.MetaDataValue,
			},
		},
	}, nil
}

// ListCardsData листинг данных о банковской карте
// так же берутся мета данные карты
func (s *Server) ListCardsData(
	ctx context.Context,
	req *bankcardsPb.ListCardsDataRequest,
) (*bankcardsPb.ListCardsDataResponse, error) {
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
	list, totalCount, err := s.storage.GetListItems(
		ctx, int64(userID), req.Page, req.PerPage, itemsConstants.TypeCard, req.Filter,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list cards data")
	}

	// Конвертируем в proto-сообщения
	var dataItems []*bankcardsPb.CardDataItem
	var pbMetaData []*bankcardsPb.MetaData

	for i := range list {
		p := list[i]

		decryptedData, err := s.Decryptor.Decrypt(p.Data, p.IV)
		if err != nil {
			continue
		}
		sensitiveData, err := itemModel.SensitiveBankCardDataFromJSON(decryptedData)
		if err != nil {
			continue
		}

		for _, val := range p.MetaDataItems {
			pbMetaDataItem := bankcardsPb.MetaData{
				Id:    val.ID,
				Name:  val.Name,
				Value: val.Value,
			}
			pbMetaData = append(pbMetaData, &pbMetaDataItem)
		}

		dataItems = append(dataItems, &bankcardsPb.CardDataItem{
			Id:              p.ID,
			Number:          sensitiveData.Number,
			ValidUntilYear:  sensitiveData.ValidUntilYear,
			ValidUntilMonth: sensitiveData.ValidUntilMonth,
			Cvv:             sensitiveData.Cvv,
			Holder:          sensitiveData.Holder,
			Description:     p.Description,
			CreatedAt:       p.CreatedAt.String(),
			MetaData:        pbMetaData,
		})
	}
	totalPages := int32(math.Ceil(float64(totalCount) / float64(req.PerPage)))

	return &bankcardsPb.ListCardsDataResponse{
		Cards:       dataItems,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		CurrentPage: req.Page,
	}, nil
}

// GetCardData получение данных об одной карте
func (s *Server) GetCardData(ctx context.Context, req *bankcardsPb.GetCardDataRequest) (*bankcardsPb.CardDataItem, error) {
	// Извлекаем userID из контекста
	if _, ok := ctx.Value("userID").(int); !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication")
	}

	// Получаем данные из репозитория
	cardData, err := s.storage.GetItem(ctx, req.Id)
	if err != nil || cardData == nil {
		return nil, status.Error(codes.Internal, "failed to get cardData")
	}

	decryptedData, err := s.Decryptor.Decrypt(cardData.Data, cardData.IV)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to decrypt cardData")
	}
	sensitiveData, err := itemModel.SensitiveBankCardDataFromJSON(decryptedData)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to parse sensitive cardData")
	}

	var metaDataList []*bankcardsPb.MetaData

	for _, val := range cardData.MetaDataItems {
		metaDataList = append(metaDataList, &bankcardsPb.MetaData{
			Id:    val.ID,
			Name:  val.Name,
			Value: val.Value,
		})
	}

	return &bankcardsPb.CardDataItem{
		Id:              cardData.ID,
		Number:          sensitiveData.Number,
		ValidUntilYear:  sensitiveData.ValidUntilYear,
		ValidUntilMonth: sensitiveData.ValidUntilMonth,
		Cvv:             sensitiveData.Cvv,
		Holder:          sensitiveData.Holder,
		Description:     cardData.Description,
		CreatedAt:       cardData.CreatedAt.String(),
		MetaData:        metaDataList,
	}, nil
}

// DeleteCardData удаление данных о карте
func (s *Server) DeleteCardData(ctx context.Context, req *bankcardsPb.DeleteCardDataRequest) (*empty.Empty, error) {
	// Извлекаем userID из контекста
	if _, ok := ctx.Value("userID").(int); !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication")
	}
	err := s.storage.DeleteItem(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete card data")
	}
	return &empty.Empty{}, nil
}

// UpdateCardData обновляем данные о карте
func (s *Server) UpdateCardData(
	ctx context.Context,
	req *bankcardsPb.UpdateCardDataRequest,
) (*bankcardsPb.CardDataItem, error) {
	if _, ok := ctx.Value("userID").(int); !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication")
	}
	itemData, err := s.storage.GetItem(ctx, req.Id)
	if err != nil || itemData == nil {
		return nil, status.Error(codes.Internal, "failed to get itemData")
	}

	// 1. Создаем структуру для шифрования
	sensitiveData := &itemModel.SensitiveBankCardData{
		Number:          req.Number,
		ValidUntilYear:  req.ValidUntilYear,
		ValidUntilMonth: req.ValidUntilMonth,
		Cvv:             req.Cvv,
		Holder:          req.Holder,
	}

	// 2. Сериализуем в JSON
	jsonData, err := sensitiveData.ToJSON()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to serialize data")
	}

	// 3. Шифруем всю структуру
	encryptedData, algorithm, iv, err := s.Encryptor.Encrypt(jsonData)
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
	return &bankcardsPb.CardDataItem{
		Id:          itemID,
		Description: req.Description,
	}, nil
}
