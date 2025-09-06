package grpc

import (
	"fmt"

	"github.com/ramil063/secondgodiplom/internal/proto/gen/auth"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/bankcard"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/binarydata"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/password"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/textdata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	conn               *grpc.ClientConn
	AuthClient         auth.AuthServiceClient
	RegistrationClient auth.RegistrationServiceClient
	PasswordsClient    password.ServiceClient
	TextDataClient     textdata.ServiceClient
	BankCardDataClient bankcard.ServiceClient
	BinaryDataClient   binarydata.ServiceClient
}

func NewGRPCClients(serverAddr string) (*Clients, error) {
	credentials := insecure.NewCredentials()
	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(credentials),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	return &Clients{
		conn:               conn,
		AuthClient:         auth.NewAuthServiceClient(conn),
		RegistrationClient: auth.NewRegistrationServiceClient(conn),
		PasswordsClient:    password.NewServiceClient(conn),
		TextDataClient:     textdata.NewServiceClient(conn),
		BankCardDataClient: bankcard.NewServiceClient(conn),
		BinaryDataClient:   binarydata.NewServiceClient(conn),
	}, nil
}
