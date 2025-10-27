package auth

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage"
	storageMock "github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/mocks"
	modelAuth "github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/auth"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/user"
	"github.com/ramil063/secondgodiplom/internal/hash"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/auth"
	storageAuth "github.com/ramil063/secondgodiplom/internal/storage/db/dml/auth"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthServer(t *testing.T) {
	type args struct {
		storage storage.Authenticator
		secret  string
	}
	tests := []struct {
		name string
		args args
		want *Server
	}{
		{
			name: "TestNewAuthServer",
			args: args{
				storage: &storageAuth.Auth{
					Repository: &repository.Repository{},
				},
				secret: "secret",
			},
			want: &Server{
				storage: &storageAuth.Auth{
					Repository: &repository.Repository{},
				},
				Secret: "secret",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAuthServer(tt.args.storage, tt.args.secret); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAuthServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx context.Context
		req *auth.LoginRequest
	}
	tests := []struct {
		name          string
		userID        int
		accessTokenID int
		args          args
	}{
		{
			name:          "TestLogin",
			userID:        1,
			accessTokenID: 1,
			args: args{
				ctx: context.Background(),
				req: &auth.LoginRequest{
					Login:    "test",
					Password: "test",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := storageMock.NewMockAuthenticator(ctrl)
			s := &Server{
				storage: mockStorage,
				Secret:  "secret",
			}

			pHash, err := hash.GetPasswordHash(tt.args.req.Password)
			assert.NoError(t, err)

			mockStorage.EXPECT().
				GetUserByLogin(tt.args.ctx, tt.args.req.Login).
				Return(&user.User{
					ID:           tt.userID,
					Login:        tt.args.req.Login,
					PasswordHash: pHash,
					FirstName:    "test",
					LastName:     "test",
				}, nil)
			mockStorage.EXPECT().
				SaveAccessToken(tt.args.ctx, tt.userID, gomock.Any()).
				Return(tt.accessTokenID, nil)
			mockStorage.EXPECT().
				SaveRefreshToken(tt.args.ctx, tt.accessTokenID, gomock.Any()).
				Return(nil)

			got, err := s.Login(tt.args.ctx, tt.args.req)
			assert.NoError(t, err)
			assert.NotEmpty(t, got.AccessToken)
			assert.NotEmpty(t, got.RefreshToken)
			assert.NotEmpty(t, got.ExpiresIn)
		})
	}
}

func TestServer_Refresh(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx context.Context
		req *auth.RefreshRequest
	}
	tests := []struct {
		name          string
		userID        int
		accessTokenID int
		args          args
	}{
		{
			name:          "TestRefresh",
			userID:        1,
			accessTokenID: 1,
			args: args{
				ctx: context.Background(),
				req: &auth.RefreshRequest{
					RefreshToken: "test",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := storageMock.NewMockAuthenticator(ctrl)
			s := &Server{
				storage: mockStorage,
				Secret:  "secret",
			}

			mockStorage.EXPECT().
				GetRefreshToken(tt.args.ctx, tt.args.req.RefreshToken).
				Return(&modelAuth.RefreshToken{
					UserID:    tt.userID,
					ExpiresAt: time.Now().Add(time.Minute),
				}, nil)
			mockStorage.EXPECT().
				RevokeRefreshToken(tt.args.ctx, tt.args.req.RefreshToken).
				Return(nil)
			mockStorage.EXPECT().
				SaveAccessToken(tt.args.ctx, tt.userID, gomock.Any()).
				Return(tt.accessTokenID, nil)

			mockStorage.EXPECT().
				SaveRefreshToken(tt.args.ctx, tt.accessTokenID, gomock.Any()).
				Return(nil)

			got, err := s.Refresh(tt.args.ctx, tt.args.req)
			assert.NoError(t, err)
			assert.NotEmpty(t, got.AccessToken)
			assert.NotEmpty(t, got.RefreshToken)
			assert.NotEmpty(t, got.ExpiresIn)
		})
	}
}
