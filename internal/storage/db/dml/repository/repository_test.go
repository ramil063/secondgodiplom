package repository

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"

	repositoryMock "github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository/mocks"
)

func TestRepository_ExecContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name            string
		ctx             context.Context
		query           string
		commandTagQuery string
		token           string
		ID              int64
		login           string
	}{
		{
			name:            "success",
			ctx:             context.Background(),
			query:           "UPDATE encrypted_item SET is_deleted=TRUE WHERE id = $1",
			commandTagQuery: "UPDATE 0 1",
			ID:              10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repositoryMock.NewMockPooler(ctrl)
			dbr := Repository{Pool: poolMock}

			expectedCommandTag := pgconn.CommandTag(tt.commandTagQuery)
			poolMock.EXPECT().
				Exec(
					context.Background(),
					tt.query,
					tt.ID).
				Return(expectedCommandTag, nil)

			_, err := dbr.ExecContext(tt.ctx, tt.query, tt.ID)
			assert.NoError(t, err)
		})
	}
}

func TestRepository_PingContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name string
	}{
		{"test 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repositoryMock.NewMockPooler(ctrl)
			dbr := Repository{Pool: poolMock}

			poolMock.EXPECT().
				Ping(context.Background()).
				Return(nil)
			err := dbr.PingContext(context.Background())
			assert.NoError(t, err)
		})
	}
}
