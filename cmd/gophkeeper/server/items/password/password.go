package password

import (
	"context"
	"math"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/items"
	passwordsModel "github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/items"
	itemsConstants "github.com/ramil063/secondgodiplom/internal/constants/items"
	passwordPb "github.com/ramil063/secondgodiplom/internal/proto/gen/items/password"
	"github.com/ramil063/secondgodiplom/internal/security/crypto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server надстройка над стандартным gRPC сервером(логика работы с паролями)
type Server struct {
	passwordPb.UnimplementedServiceServer

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

// CreatePassword создание записи о пароле
// так же сохранение метаданных о нем
// все специфичные данные шифруются
func (s *Server) CreatePassword(ctx context.Context, req *passwordPb.CreatePasswordRequest) (*passwordPb.PasswordItem, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication")
	}
	// 1. Создаем структуру для шифрования
	sensitiveData := &passwordsModel.SensitivePasswordData{
		Login:    req.Login,
		Password: req.Password,
		Target:   req.Target,
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
	itemID, err := s.storage.SaveEncryptedData(ctx, &passwordsModel.EncryptedItem{
		UserID:              int64(userID),
		Type:                itemsConstants.TypePasswords,
		Data:                encryptedData,
		Description:         req.Description,
		EncryptionAlgorithm: algorithm,
		Iv:                  iv,
	})

	// 4. Сохраняем метаданные в отдельную таблицу
	err = s.storage.SaveMetadata(ctx, &passwordsModel.MetaData{
		ItemID: itemID,
		Name:   req.MetaDataName,
		Value:  req.MetaDataValue,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to save meta data")
	}

	// 5. Возвращаем ответ
	return &passwordPb.PasswordItem{
		Id:          itemID,
		Login:       req.Login,
		Password:    req.Password,
		Target:      req.Target,
		Description: req.Description,
		MetaData: []*passwordPb.MetaData{
			{
				Name:  req.MetaDataName,
				Value: req.MetaDataValue,
			},
		},
	}, nil
}

// ListPasswords листинг данных о паролях
// так же показываются метаданные по каждому
func (s *Server) ListPasswords(
	ctx context.Context,
	req *passwordPb.ListPasswordsRequest,
) (*passwordPb.ListPasswordsResponse, error) {
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
		return nil, status.Error(codes.Internal, "failed to list passwords")
	}

	// Конвертируем в proto-сообщения
	var pbPasswords []*passwordPb.PasswordItem
	var pbMetaData []*passwordPb.MetaData

	for i := range passwordsList {
		p := passwordsList[i]

		decryptedData, err := s.Decryptor.Decrypt(p.Data, p.IV)
		if err != nil {
			continue
		}
		sensitiveData, err := passwordsModel.SensitivePasswordDataFromJSON(decryptedData)
		if err != nil {
			continue
		}

		for _, val := range p.MetaDataItems {
			pbMetaDataItem := passwordPb.MetaData{
				Id:    val.ID,
				Name:  val.Name,
				Value: val.Value,
			}
			pbMetaData = append(pbMetaData, &pbMetaDataItem)
		}

		pbPasswords = append(pbPasswords, &passwordPb.PasswordItem{
			Id:          p.ID,
			Login:       sensitiveData.Login,
			Password:    sensitiveData.Password,
			Target:      sensitiveData.Target,
			Description: p.Description,
			CreatedAt:   p.CreatedAt.String(),
			MetaData:    pbMetaData,
		})
	}
	totalPages := int32(math.Ceil(float64(totalCount) / float64(req.PerPage)))

	return &passwordPb.ListPasswordsResponse{
		Passwords:   pbPasswords,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		CurrentPage: req.Page,
	}, nil
}

// GetPassword получение данных по 1 паролю
// так же возвращаются и метаданные
func (s *Server) GetPassword(ctx context.Context, req *passwordPb.GetPasswordRequest) (*passwordPb.PasswordItem, error) {
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
	sensitiveData, err := passwordsModel.SensitivePasswordDataFromJSON(decryptedData)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to parse sensitive password")
	}

	var metaDataList []*passwordPb.MetaData

	for _, val := range password.MetaDataItems {
		metaDataList = append(metaDataList, &passwordPb.MetaData{
			Id:    val.ID,
			Name:  val.Name,
			Value: val.Value,
		})
	}

	return &passwordPb.PasswordItem{
		Id:          password.ID,
		Login:       sensitiveData.Login,
		Password:    sensitiveData.Password,
		Target:      sensitiveData.Target,
		Description: password.Description,
		CreatedAt:   password.CreatedAt.String(),
		MetaData:    metaDataList,
	}, nil
}

// DeletePassword удаление данных о пароле
// используется мягкое удаление
func (s *Server) DeletePassword(ctx context.Context, req *passwordPb.DeletePasswordRequest) (*empty.Empty, error) {
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

// UpdatePassword обновление данных по паролю
func (s *Server) UpdatePassword(
	ctx context.Context,
	req *passwordPb.UpdatePasswordRequest,
) (*passwordPb.PasswordItem, error) {
	if _, ok := ctx.Value("userID").(int); !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication")
	}
	password, err := s.storage.GetItem(ctx, req.Id)
	if err != nil || password == nil {
		return nil, status.Error(codes.Internal, "failed to get password")
	}

	// 1. Создаем структуру для шифрования
	sensitiveData := &passwordsModel.SensitivePasswordData{
		Login:    req.Login,
		Password: req.Password,
		Target:   req.Target,
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
	itemID, err := s.storage.UpdateItem(ctx, req.Id, &passwordsModel.EncryptedItem{
		Data:                encryptedData,
		Description:         req.Description,
		EncryptionAlgorithm: algorithm,
		Iv:                  iv,
	})

	// TODO Сохранять метаданные в отдельном реквесте
	//err = s.storage.SaveMetadata(ctx, &passwordsModel.MetaData{
	//	ItemID: itemID,
	//	Name:   req.MetaDataName,
	//	Value:  req.MetaDataValue,
	//})

	// 5. Возвращаем ответ
	return &passwordPb.PasswordItem{
		Id:          itemID,
		Description: req.Description,
	}, nil
}
