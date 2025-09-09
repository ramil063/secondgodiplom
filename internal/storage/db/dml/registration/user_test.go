package registration

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/user"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/mock"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
	repository2 "github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository/mocks"
	"github.com/stretchr/testify/assert"
)

func TestReg_RegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx  context.Context
		user *user.User
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test 1",
			args: args{
				ctx: context.Background(),
				user: &user.User{
					Login:        "test",
					PasswordHash: "test",
					FirstName:    "test",
					LastName:     "test",
				},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)
			s := &Reg{
				Repository: &repository.Repository{Pool: poolMock},
			}

			poolMock.EXPECT().
				QueryRow(
					tt.args.ctx,
					gomock.Any(),
					tt.args.user.Login,
					tt.args.user.PasswordHash,
					tt.args.user.FirstName,
					tt.args.user.LastName,
					true,
				).
				Return(&mock.Row{
					Values: []interface{}{
						tt.want,
					},
				})
			got, err := s.RegisterUser(tt.args.ctx, tt.args.user)
			assert.Nil(t, err)
			if got != tt.want {
				t.Errorf("RegisterUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}
