package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTokenHash(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "TestGetTokenHash",
			args: args{
				token: "123",
			},
			want: "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GetTokenHash(tt.args.token), "GetTokenHash(%v)", tt.args.token)
		})
	}
}
