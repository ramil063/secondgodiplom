package items

import (
	"context"
	"log"

	"github.com/ramil063/secondgodiplom/internal/security/cookie"
	"google.golang.org/grpc/metadata"
)

func CreateAuthContext() context.Context {
	accessToken, _, _, err := cookie.LoadTokens()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	md := metadata.Pairs("authorization", "Bearer "+accessToken)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx
}
