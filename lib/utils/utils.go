package utils

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"strconv"
)

func RandInt(length int) string {
	return fmt.Sprintf("%0"+strconv.Itoa(length)+"d", rand.IntN(10^length))
}

func CopyNonDefaultValues(src, dst interface{}) {
	srcVal := reflect.ValueOf(src).Elem()
	dstVal := reflect.ValueOf(dst).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		dstField := dstVal.Field(i)

		if !reflect.DeepEqual(srcField.Interface(), reflect.Zero(srcField.Type()).Interface()) {
			dstField.Set(srcField)
		}
	}
}
