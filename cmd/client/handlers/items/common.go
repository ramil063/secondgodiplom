package items

import (
	"context"
	"log"

	"google.golang.org/grpc/metadata"

	cookieContants "github.com/ramil063/secondgodiplom/internal/constants/cookie"
	"github.com/ramil063/secondgodiplom/internal/security/cookie"
)

// CreateAuthContext создание специального авторизационного контекста для отправки на gRPC сервер
func CreateAuthContext() context.Context {
	accessToken, _, _, err := cookie.LoadTokens(cookieContants.FileToSaveCookie)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	md := metadata.Pairs("authorization", "Bearer "+accessToken)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx
}
