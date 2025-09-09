package db

import (
	"context"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgconn"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
	repository2 "github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCheckPing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name string
	}{
		{
			name: "Ping",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)
			rep := &repository.Repository{Pool: poolMock}

			poolMock.EXPECT().
				Ping(gomock.Any()).
				Return(nil)

			err := CheckPing(*rep)
			assert.NoError(t, err)
		})
	}
}

func TestCreateTables(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name string
	}{
		{
			name: "CreateTables",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)
			rep := &repository.Repository{Pool: poolMock}

			expectedCommandTag := pgconn.CommandTag("INSERT 0 1")
			poolMock.EXPECT().
				Exec(
					context.Background(),
					gomock.Any()).
				Return(expectedCommandTag, nil)

			err := CreateTables(*rep)
			assert.NoError(t, err)
		})
	}
}

func TestInit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name string
	}{
		{
			name: "Init",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)
			rep := &repository.Repository{Pool: poolMock}

			poolMock.EXPECT().
				Ping(gomock.Any()).
				Return(nil)

			expectedCommandTag := pgconn.CommandTag("INSERT 0 1")
			poolMock.EXPECT().
				Exec(
					context.Background(),
					gomock.Any()).
				Return(expectedCommandTag, nil)

			err := Init(*rep)
			assert.NoError(t, err)
		})
	}
}

func TestStorage_GetRepository(t *testing.T) {
	type fields struct {
		Repository *repository.Repository
	}
	tests := []struct {
		name   string
		fields fields
		want   repository.Repository
	}{
		{
			name: "GetRepository",
			fields: fields{
				Repository: &repository.Repository{},
			},
			want: repository.Repository{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{
				Repository: tt.fields.Repository,
			}
			if got := s.GetRepository(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_SetRepository(t *testing.T) {
	type fields struct {
		Repository *repository.Repository
	}
	type args struct {
		repository *repository.Repository
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "SetRepository",
			fields: fields{
				Repository: &repository.Repository{},
			},
			args: args{
				repository: &repository.Repository{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{
				Repository: tt.fields.Repository,
			}
			s.SetRepository(tt.args.repository)
			if got := s.GetRepository(); !reflect.DeepEqual(got, *tt.fields.Repository) {
				t.Errorf("GetRepository() = %v, want %v", got, *tt.fields.Repository)
			}
		})
	}
}
