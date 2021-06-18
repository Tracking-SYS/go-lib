package reflection

import "reflect"

func Coalesce(vals ...interface{}) interface{} {
	if len(vals) == 0 {
		return nil
	}

	var t reflect.Type
	for _, v := range vals {
		if v == nil {
			continue
		}
		if t == nil {
			t = reflect.TypeOf(v)
		} else if t.Kind() != reflect.TypeOf(v).Kind() {
			panic("args must be same type")
		}
		if !reflect.ValueOf(v).IsZero() {
			return v
		}
	}
	if t == nil {
		return nil
	}
	return reflect.Zero(t).Interface()
}
