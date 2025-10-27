package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRow_Scan(t *testing.T) {
	type fields struct {
		Values []interface{}
		Err    error
	}
	type args struct {
		dest []interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test 1",
			fields: fields{
				Values: []interface{}{},
			},
			args: args{
				dest: []interface{}{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Row{
				Values: tt.fields.Values,
				Err:    tt.fields.Err,
			}
			err := m.Scan(tt.args.dest...)
			assert.NoError(t, err)
		})
	}
}
