package reflection

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

var ErrTypeConvert = errors.New("convert")

var floatType = reflect.TypeOf(float64(0))
var int64Type = reflect.TypeOf(int64(0))

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

func ParseFloat(unk interface{}) (float64, error) {
	if strV, ok := unk.(string); ok {
		return strconv.ParseFloat(strV, 64)
	}

	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, fmt.Errorf("cannot convert %v to float64: %w", v.Type(), ErrTypeConvert)
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}

func ParseInt64(unk interface{}) (int64, error) {
	if strV, ok := unk.(string); ok {
		return strconv.ParseInt(strV, 10, 64)
	}

	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(int64Type) {
		return 0, fmt.Errorf("cannot convert %v to float64: %w", v.Type(), ErrTypeConvert)
	}
	fv := v.Convert(int64Type)
	return fv.Int(), nil
}
