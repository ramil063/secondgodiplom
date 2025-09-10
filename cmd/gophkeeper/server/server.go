package server

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	serverConfig "github.com/ramil063/secondgodiplom/cmd/gophkeeper/config"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/interceptors"
	authServer "github.com/ramil063/secondgodiplom/cmd/gophkeeper/server/auth"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/server/items/bankcard"
	binary2 "github.com/ramil063/secondgodiplom/cmd/gophkeeper/server/items/binary"
	passwordServer "github.com/ramil063/secondgodiplom/cmd/gophkeeper/server/items/password"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/server/items/text"
	regServer "github.com/ramil063/secondgodiplom/cmd/gophkeeper/server/registration"
	localStorage "github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/items"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/items/binary"
	"github.com/ramil063/secondgodiplom/internal/logger"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/auth"
	itemsBankcard "github.com/ramil063/secondgodiplom/internal/proto/gen/items/bankcard"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/binarydata"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/password"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/textdata"
	"github.com/ramil063/secondgodiplom/internal/security/crypto"
	"github.com/ramil063/secondgodiplom/internal/storage/db"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
)

// PrepareServerEnvironment подготавливает окружение для работы сервера
func PrepareServerEnvironment() (*serverConfig.ServerConfig, localStorage.Storager, *crypto.Manager, error) {
	config, err := serverConfig.GetConfig()
	if err != nil {
		logger.WriteErrorLog(err.Error())
	}
	if config == nil {
		return nil, nil, nil, fmt.Errorf("error in getting config")
	}

	var grpcStorage = localStorage.NewDBStorage()

	if config.DatabaseURI != "" {
		rep, errRepo := repository.NewRepository(config)
		if errRepo != nil {
			logger.WriteErrorLog(errRepo.Error())
			return nil, nil, nil, errRepo
		}

		grpcStorage.SetRepository(rep)
		err = db.Init(*rep)
		if err != nil {
			log.Println(err.Error())
			logger.WriteErrorLog("Init DB: " + err.Error())
			return nil, nil, nil, err
		}
	}

	manager := crypto.NewCryptoManager()
	if config.CryptoKey != "" {
		decryptor, err := crypto.NewAes256gcmDecryptor([]byte(config.CryptoKey))
		if err != nil {
			logger.WriteErrorLog(err.Error())
		}
		manager.SetGRPCDecryptor(decryptor)

		encryptor, err := crypto.NewAes256gcmEncryptor([]byte(config.CryptoKey))
		if err != nil {
			logger.WriteErrorLog(err.Error())
		}
		manager.SetGRPCEncryptor(encryptor)
	}
	return config, grpcStorage, manager, nil
}

// GetGRPCServer возвращает настроенный и запущенный gRPC сервер
func GetGRPCServer(config *serverConfig.ServerConfig) (*grpc.Server, net.Listener, error) {
	var err error

	lis, err := net.Listen("tcp", config.Address)
	if err != nil {
		logger.WriteErrorLog(err.Error())
		return nil, nil, err
	}

	authInterceptor := interceptors.NewAuthInterceptors(config.Secret)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary),
		grpc.StreamInterceptor(authInterceptor.Stream),
	)
	return grpcServer, lis, nil
}

// RegisterServiceServers регистрация сервисов в сервере
func RegisterServiceServers(
	grpcServer *grpc.Server,
	storage localStorage.Storager,
	config *serverConfig.ServerConfig,
	manager *crypto.Manager,
) {
	regStorage := localStorage.NewRegistrationStorage(storage.GetRepository())
	authStorage := localStorage.NewAuthStorage(storage.GetRepository())
	newStorage := items.NewStorage(storage.GetRepository())
	newBinaryStorage := binary.NewStorage(storage.GetRepository())

	passServer := passwordServer.NewServer(newStorage, manager.GetGRPCEncryptor(), manager.GetGRPCDecryptor())
	textDataServer := text.NewServer(newStorage, manager.GetGRPCEncryptor(), manager.GetGRPCDecryptor())
	bankcardServer := bankcard.NewServer(newStorage, manager.GetGRPCEncryptor(), manager.GetGRPCDecryptor())
	binaryServer := binary2.NewServer(newBinaryStorage, manager.GetGRPCEncryptor(), manager.GetGRPCDecryptor(), config)

	auth.RegisterRegistrationServiceServer(grpcServer, regServer.NewRegistrationServer(regStorage))
	auth.RegisterAuthServiceServer(grpcServer, authServer.NewAuthServer(authStorage, config.Secret))
	password.RegisterServiceServer(grpcServer, passServer)
	textdata.RegisterServiceServer(grpcServer, textDataServer)
	itemsBankcard.RegisterServiceServer(grpcServer, bankcardServer)
	binarydata.RegisterServiceServer(grpcServer, binaryServer)
}
