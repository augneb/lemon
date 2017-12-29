package lemon

import (
	"reflect"
	"strings"
	"github.com/augneb/util"
)

type TableName interface {
	TableName() string
}

type structCache struct {
	Name          string
	PrimaryKey    string
	Fields        []string
	AutoIncrement map[int]string
	UniqueKeys    map[string][]string
}

// 获取表结构信息
func (o *Orm) GetTableInfo(table interface{}, isType ...bool) *structCache {
	switch table.(type) {
	case string:
		s, _ := o.structCache.Load(table.(string))
		return s.(*structCache)
	}

	var v reflect.Value
	var t reflect.Type

	if len(isType) > 0 && isType[0] == true {
		t = table.(reflect.Type)
	} else {
		v = reflect.ValueOf(table)

		if v.Kind() == reflect.Ptr {
			t = v.Elem().Type()
		} else {
			t = v.Type()
		}
	}

	// 结构体名
	structName := t.Name()

	if !v.IsValid() {
		v = reflect.New(t)
	}

	// 表名
	tableName := ""
	if fn, ok := v.Interface().(TableName); ok {
		tableName = fn.TableName()
	} else {
		tableName = util.ToSnakeCase(structName, true)
	}

	if s, ok := o.structCache.Load(tableName); ok {
		return s.(*structCache)
	}

	return o.cacheTableInfo(t, tableName)
}

// 缓存表结构信息
func (o *Orm) cacheTableInfo(t reflect.Type, tName string) *structCache {
	newVal := new(structCache)
	newVal.Name = tName
	newVal.Fields = make([]string, t.NumField())
	newVal.AutoIncrement = map[int]string{}
	newVal.UniqueKeys = map[string][]string{}

	for i := 0; i < t.NumField(); i++ {
		// 取 Tag
		tag := t.Field(i).Tag.Get("db")

		// 表结构字段名
		n := t.Field(i).Name

		// 没有 Tag，则数据库字段名为表结构字段名
		if tag == "" {
			newVal.Fields[i] = util.ToSnakeCase(n, true)
			continue
		}

		// 去空格，并且按空格分割成 []string
		tags := strings.Fields(tag)

		// 第一个 tag，如果不是字段名，则数据库字段名为结构体字段名
		f := tags[0][:2]
		if f == "ai" || f == "pk" || f == "un" {
			f = util.ToSnakeCase(n, true)
		} else {
			f = strings.Trim(tags[0], "'")
		}

		newVal.Fields[i] = f

		for _, v := range tags {
			if v == "ai" {
				newVal.AutoIncrement[i] = v
				continue
			}

			if v == "pk" {
				newVal.PrimaryKey = f
				continue
			}

			if v[:2] == "un" {
				if u := strings.Replace(v, "unique:", "", -1); u == "" {
					newVal.UniqueKeys[f] = append(newVal.UniqueKeys[f], f)
				} else {
					newVal.UniqueKeys[u] = append(newVal.UniqueKeys[u], f)
				}

				continue
			}
		}
	}

	o.structCache.Store(tName, newVal)

	return newVal
}
