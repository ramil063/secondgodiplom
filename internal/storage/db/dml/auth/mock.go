package auth

import (
	"errors"
	"reflect"
)

type mockRow struct {
	values []interface{}
	err    error
}

func (m *mockRow) Scan(dest ...interface{}) error {
	if m.err != nil {
		return m.err
	}
	for i := range dest {
		if i >= len(m.values) {
			return errors.New("not enough values to scan")
		}
		val := reflect.ValueOf(dest[i]).Elem()
		val.Set(reflect.ValueOf(m.values[i]))
	}
	return nil
}
