package mock

import (
	"errors"
	"reflect"
)

type Row struct {
	Values []interface{}
	Err    error
}

func (m *Row) Scan(dest ...interface{}) error {
	if m.Err != nil {
		return m.Err
	}
	for i := range dest {
		if i >= len(m.Values) {
			return errors.New("not enough values to scan")
		}
		val := reflect.ValueOf(dest[i]).Elem()
		val.Set(reflect.ValueOf(m.Values[i]))
	}
	return nil
}
