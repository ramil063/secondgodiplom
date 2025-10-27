package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ramil063/secondgodiplom/internal/storage/db"
)

func TestNewDBStorage(t *testing.T) {
	tests := []struct {
		name string
		want Storager
	}{
		{
			name: "test 1",
			want: &db.Storage{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDBStorage()
			assert.Equal(t, tt.want, got)
		})
	}
}
