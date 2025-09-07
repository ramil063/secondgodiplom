package auth

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/user"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
	repository2 "github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAuth_GetUserByLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx   context.Context
		login string
	}
	tests := []struct {
		name  string
		query string
		args  args
		want  *user.User
	}{
		{
			name:  "success",
			query: `SELECT id, password_hash FROM users WHERE login = $1`,
			args: args{
				ctx:   context.Background(),
				login: "test",
			},
			want: &user.User{
				ID:           1,
				PasswordHash: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)
			s := &Auth{
				Repository: &repository.Repository{Pool: poolMock},
			}

			poolMock.EXPECT().
				QueryRow(
					tt.args.ctx,
					tt.query,
					tt.args.login,
				).
				Return(&mockRow{
					values: []interface{}{
						tt.want.ID,
						[]byte(tt.want.PasswordHash),
					},
				})
			got, err := s.GetUserByLogin(tt.args.ctx, tt.args.login)
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, got, "GetUserByLogin(%v, %v)", tt.args.ctx, tt.args.login)
		})
	}
}
