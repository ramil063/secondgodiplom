package binary

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgconn"
	"github.com/pashagolub/pgxmock"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/items"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/binarydata"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/mock"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
	repository2 "github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository/mocks"
	"github.com/stretchr/testify/assert"
)

func TestItem_CreateFileRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx      context.Context
		userID   int
		metadata *binarydata.FileMetadata
	}
	tests := []struct {
		name   string
		query  string
		args   args
		fileID int64
		want   int64
	}{
		{
			name: "test 1",
			args: args{
				ctx:    context.Background(),
				userID: 1,
				metadata: &binarydata.FileMetadata{
					Filename:     "filename",
					MimeType:     "mime_type",
					OriginalSize: 1,
					Description:  "description",
					ChunkSize:    1,
					TotalChunks:  1,
				},
			},
			fileID: 1,
			want:   1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)
			i := &Item{
				Repository: &repository.Repository{Pool: poolMock},
			}

			poolMock.EXPECT().
				QueryRow(
					tt.args.ctx,
					gomock.Any(),
					tt.args.userID,
					tt.args.metadata.Filename,
					tt.args.metadata.MimeType,
					tt.args.metadata.OriginalSize,
					tt.args.metadata.Description,
					tt.args.metadata.ChunkSize,
					tt.args.metadata.TotalChunks,
				).
				Return(&mock.Row{
					Values: []interface{}{
						tt.fileID,
					},
				})

			got, err := i.CreateFileRecord(tt.args.ctx, tt.args.userID, tt.args.metadata)
			assert.NoError(t, err)
			if got != tt.want {
				t.Errorf("CreateFileRecord() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItem_DeleteFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx    context.Context
		userID int64
		fileID int64
	}
	tests := []struct {
		name  string
		query string
		args  args
	}{
		{
			name:  "success",
			query: `UPDATE binary_file SET is_deleted=TRUE WHERE id = $1 AND user_id = $2`,
			args: args{
				ctx:    context.Background(),
				userID: 1,
				fileID: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)
			i := &Item{
				Repository: &repository.Repository{Pool: poolMock},
			}

			expectedCommandTag := pgconn.CommandTag("UPDATE 0 1")
			poolMock.EXPECT().
				Exec(
					tt.args.ctx,
					tt.query,
					tt.args.fileID,
					tt.args.userID).
				Return(expectedCommandTag, nil)
			err := i.DeleteFile(tt.args.ctx, tt.args.userID, tt.args.fileID)
			assert.NoError(t, err)
		})
	}
}

func TestItem_GetChunksInRange(t *testing.T) {
	type args struct {
		ctx    context.Context
		fileID int64
		start  int32
		end    int32
	}
	tests := []struct {
		name string
		args args
		want []*items.ChunkData
	}{
		{
			name: "test 1",
			args: args{
				ctx:    context.Background(),
				fileID: 1,
				start:  1,
				end:    2,
			},
			want: []*items.ChunkData{
				{
					ChunkIndex:          1,
					EncryptedData:       []byte("test"),
					EncryptionAlgorithm: "AES-256-GCM",
					IV:                  []byte("test"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			i := &Item{
				Repository: &repository.Repository{Pool: mock},
			}

			rows := mock.NewRows([]string{
				"chunk_index",
				"encrypted_data",
				"encryption_algorithm",
				"iv",
			}).AddRow(
				tt.want[0].ChunkIndex,
				tt.want[0].EncryptedData,
				tt.want[0].EncryptionAlgorithm,
				tt.want[0].IV,
			)

			mock.ExpectQuery("SELECT.*chunk_index.*encrypted_data").
				WithArgs(tt.args.fileID, tt.args.start, tt.args.end-1).
				WillReturnRows(rows)

			got, err := i.GetChunksInRange(tt.args.ctx, tt.args.fileID, tt.args.start, tt.args.end)
			assert.NoError(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetChunksInRange() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItem_GetFileInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx    context.Context
		fileID int64
		userID int64
	}
	tests := []struct {
		name         string
		args         args
		want         *items.FileInfo
		metaDataJSON []byte
	}{
		{
			name: "test 1",
			args: args{
				ctx:    context.Background(),
				fileID: 1,
				userID: 1,
			},
			want:         &items.FileInfo{},
			metaDataJSON: []byte(`[]`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)
			i := &Item{
				Repository: &repository.Repository{Pool: poolMock},
			}

			var metadata []*items.MetaData
			err := json.Unmarshal(tt.metaDataJSON, &metadata)
			assert.NoError(t, err)
			tt.want.MetaDataItems = metadata

			poolMock.EXPECT().
				QueryRow(
					tt.args.ctx,
					gomock.Any(),
					tt.args.fileID,
					tt.args.userID,
				).
				Return(&mock.Row{
					Values: []interface{}{
						tt.want.ID,
						tt.want.Filename,
						tt.want.MimeType,
						tt.want.OriginalSize,
						tt.want.Description,
						tt.want.ChunkSize,
						tt.want.TotalChunks,
						tt.want.CreatedAt,
						tt.metaDataJSON,
					},
				})

			got, err := i.GetFileInfo(tt.args.ctx, tt.args.fileID, tt.args.userID)
			assert.NoError(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFileInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItem_GetListFiles(t *testing.T) {
	type args struct {
		ctx     context.Context
		userID  int64
		page    int32
		perPage int32
		filter  string
	}
	tests := []struct {
		name  string
		args  args
		want  []*items.FileInfo
		want1 int32
	}{
		{
			name: "test 1",
			args: args{
				ctx:     context.Background(),
				userID:  1,
				page:    1,
				perPage: 1,
				filter:  "",
			},
			want: []*items.FileInfo{
				{
					ID:           1,
					Filename:     "test1.txt",
					MimeType:     "text/plain",
					OriginalSize: 0,
					Description:  "",
					CreatedAt:    time.Now(),
				},
			},
			want1: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock, err := pgxmock.NewPool()
			i := &Item{
				Repository: &repository.Repository{Pool: poolMock},
			}

			rows := poolMock.NewRows([]string{
				"id",
				"filename",
				"mime_type",
				"original_size",
				"description",
				"created_at",
			}).AddRow(
				tt.want[0].ID,
				tt.want[0].Filename,
				tt.want[0].MimeType,
				tt.want[0].OriginalSize,
				tt.want[0].Description,
				tt.want[0].CreatedAt,
			)

			rows1 := poolMock.NewRows([]string{
				"count",
			}).AddRow(
				tt.want1,
			)

			poolMock.ExpectQuery("SELECT.*bf.id.*bf.filename.*").
				WithArgs(tt.args.userID, tt.args.perPage, (tt.args.page-1)*tt.args.perPage).
				WillReturnRows(rows)
			poolMock.ExpectQuery("SELECT.*COUNT.*").
				WithArgs(tt.args.userID).
				WillReturnRows(rows1)

			got, got1, err := i.GetListFiles(tt.args.ctx, tt.args.userID, tt.args.page, tt.args.perPage, tt.args.filter)
			assert.NoError(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetListFiles() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetListFiles() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestItem_GetTotalCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx    context.Context
		query  string
		userID int64
		filter string
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			name: "test 1",
			args: args{
				ctx:    context.Background(),
				query:  "",
				userID: 1,
				filter: "",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)
			i := &Item{
				Repository: &repository.Repository{Pool: poolMock},
			}

			poolMock.EXPECT().
				QueryRow(
					tt.args.ctx,
					gomock.Any(),
					tt.args.userID,
				).
				Return(&mock.Row{
					Values: []interface{}{
						tt.want,
					},
				})

			got, err := i.GetTotalCount(tt.args.ctx, tt.args.query, tt.args.userID, tt.args.filter)
			assert.NoError(t, err)
			if got != tt.want {
				t.Errorf("GetTotalCount() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItem_MarkFileComplete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx        context.Context
		fileID     int64
		totalBytes int64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test 1",
			args: args{
				ctx:        context.Background(),
				fileID:     1,
				totalBytes: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)
			i := &Item{
				Repository: &repository.Repository{Pool: poolMock},
			}

			expectedCommandTag := pgconn.CommandTag("UPDATE 0 1")
			poolMock.EXPECT().
				Exec(
					tt.args.ctx,
					gomock.Any(),
					tt.args.totalBytes,
					tt.args.fileID).
				Return(expectedCommandTag, nil)
			err := i.MarkFileComplete(tt.args.ctx, tt.args.fileID, tt.args.totalBytes)
			assert.NoError(t, err)
		})
	}
}

func TestItem_SaveChunk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		Repository *repository.Repository
	}
	type args struct {
		ctx           context.Context
		fileID        int64
		chunkIndex    int32
		encryptedData []byte
		algorithm     string
		iv            []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test 1",
			args: args{
				ctx:           context.Background(),
				fileID:        1,
				chunkIndex:    0,
				encryptedData: []byte("test"),
				algorithm:     "AES-256-GCM",
				iv:            nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolMock := repository2.NewMockPooler(ctrl)
			i := &Item{
				Repository: &repository.Repository{Pool: poolMock},
			}

			expectedCommandTag := pgconn.CommandTag("UPDATE 0 1")
			poolMock.EXPECT().
				Exec(
					tt.args.ctx,
					gomock.Any(),
					tt.args.fileID,
					tt.args.chunkIndex,
					tt.args.encryptedData,
					tt.args.algorithm,
					tt.args.iv).
				Return(expectedCommandTag, nil)
			err := i.SaveChunk(tt.args.ctx, tt.args.fileID, tt.args.chunkIndex, tt.args.encryptedData, tt.args.algorithm, tt.args.iv)
			assert.NoError(t, err)
		})
	}
}
