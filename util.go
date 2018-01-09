package lemon

import (
	"strings"
	"regexp"
	"reflect"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")

func _if(cond bool, v1, v2 interface{}) interface{} {
	if cond {
		return v1
	}

	return v2
}

func sliceFilter(items []string) []string {
	result := []string{}
	for _, text := range items {
		if strings.Trim(text, " ") != "" {
			result = append(result, text)
		}
	}

	return result
}

func sliceIn(elt, slice interface{}) bool {
	v := reflect.Indirect(reflect.ValueOf(slice))

	for i := 0; i < v.Len(); i++ {
		if reflect.DeepEqual(v.Index(i).Interface(), elt) {
			return true
		}
	}

	return false
}

func bytesClone(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)

	return c
}

func toSnakeCase(str string, toLower bool) string {
	snake := matchAllCap.ReplaceAllString(matchFirstCap.ReplaceAllString(str, "${1}_${2}"), "${1}_${2}")

	if toLower {
		return strings.ToLower(snake)
	}

	return snake
}