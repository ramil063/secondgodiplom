package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"

	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/auth"
	"github.com/ramil063/secondgodiplom/internal/hash"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
	repositoryMock "github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository/mocks"
)

func TestAuth_GetRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx          context.Context
		refreshToken string
	}
	tests := []struct {
		name string
		args args
		want *auth.RefreshToken
	}{
		{
			name: "test 1",
			args: args{
				ctx:          context.Background(),
				refreshToken: "refresh_token_hash",
			},
			want: &auth.RefreshToken{
				TokenHash:       "refresh_token_hash",
				AccessTokenHash: "access_token_hash",
				UserID:          1,
				ExpiresAt:       time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repositoryMock.NewMockPooler(ctrl)

			poolMock.EXPECT().
				QueryRow(
					tt.args.ctx,
					gomock.Any(),
					hash.GetTokenHash(tt.args.refreshToken),
				).
				Return(&mockRow{
					values: []interface{}{
						[]byte(tt.want.TokenHash),
						[]byte(tt.want.AccessTokenHash),
						tt.want.UserID,
						tt.want.ExpiresAt,
					},
				})

			s := &Auth{
				Repository: &repository.Repository{Pool: poolMock},
			}

			_, err := s.GetRefreshToken(tt.args.ctx, tt.args.refreshToken)
			assert.NoError(t, err)
		})
	}
}

func TestAuth_RevokeRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx          context.Context
		query        string
		refreshToken string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test 1",
			args: args{
				ctx:          context.Background(),
				query:        `UPDATE oauth_refresh_token SET is_revoked=TRUE WHERE token_hash = $1`,
				refreshToken: "refresh_token_hash",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repositoryMock.NewMockPooler(ctrl)

			s := &Auth{
				Repository: &repository.Repository{Pool: poolMock},
			}

			expectedCommandTag := pgconn.CommandTag("UPDATE 0 1")
			poolMock.EXPECT().
				Exec(
					tt.args.ctx,
					tt.args.query,
					hash.GetTokenHash(tt.args.refreshToken)).
				Return(expectedCommandTag, nil)
			err := s.RevokeRefreshToken(tt.args.ctx, tt.args.refreshToken)
			assert.NoError(t, err)
		})
	}
}

func TestAuth_SaveAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx    context.Context
		userID int
		query  string
		token  string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test 1",
			args: args{
				ctx:    context.Background(),
				query:  "INSERT INTO oauth_access_token (token_hash, user_id, expires_at) VALUES ($1, $2, $3) RETURNING id",
				userID: 1,
				token:  "access_token_hash",
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repositoryMock.NewMockPooler(ctrl)
			s := &Auth{
				Repository: &repository.Repository{Pool: poolMock},
			}
			poolMock.EXPECT().
				QueryRow(
					tt.args.ctx,
					tt.args.query,
					hash.GetTokenHash(tt.args.token),
				).
				Return(&mockRow{
					values: []interface{}{
						tt.want,
					},
				})

			got, err := s.SaveAccessToken(tt.args.ctx, tt.args.userID, tt.args.token)
			assert.NoError(t, err)
			if got != tt.want {
				t.Errorf("SaveAccessToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuth_SaveRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx           context.Context
		query         string
		accessTokenId int
		token         string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test 1",
			args: args{
				ctx:           context.Background(),
				query:         "INSERT INTO oauth_refresh_token (token_hash, access_token_id, expires_at) VALUES ($1, $2, $3)",
				accessTokenId: 1,
				token:         "refresh_token_hash",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repositoryMock.NewMockPooler(ctrl)
			s := &Auth{
				Repository: &repository.Repository{Pool: poolMock},
			}
			expectedCommandTag := pgconn.CommandTag("UPDATE 0 1")
			poolMock.EXPECT().
				Exec(
					tt.args.ctx,
					tt.args.query,
					hash.GetTokenHash(tt.args.token),
					tt.args.accessTokenId,
					gomock.Any()).
				Return(expectedCommandTag, nil)
			if err := s.SaveRefreshToken(tt.args.ctx, tt.args.accessTokenId, tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("SaveRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
