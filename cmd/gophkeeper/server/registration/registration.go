package registration

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/user"
	"github.com/ramil063/secondgodiplom/internal/hash"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/auth"
)

// RegServer надстройка над стандартным gRPC сервером(регистрация)
type RegServer struct {
	auth.UnimplementedRegistrationServiceServer

	storage storage.Registerer
}

// NewRegistrationServer инициализация сервера регистрации и его хранилища
func NewRegistrationServer(storage storage.Registerer) *RegServer {
	return &RegServer{
		storage: storage,
	}
}

// Register зарегистрировать пользователя
// сохранение основных данных о пользователе
// применяется хеширование пароля
func (s *RegServer) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	// 1. Хеширование пароля
	hashedPassword, err := hash.GetPasswordHash(req.Password)
	if err != nil {
		return nil, err
	}
	// 2. Создание пользователя в БД
	userID, err := s.storage.RegisterUser(ctx, &user.User{
		Login:        req.Login,
		PasswordHash: hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &auth.RegisterResponse{
		UserId: strconv.Itoa(userID),
	}, nil
}
