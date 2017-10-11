package lemon

import (
	"fmt"
	"reflect"
	"strings"
	"strconv"

	"github.com/vmihailenco/msgpack"
	"time"
	"database/sql"
)

func (s *Session) getFromCache(cacheKey string, obj interface{}) (err error) {
	err = nil

	defer func() {
		s.queryTime = time.Since(s.queryStart).Seconds()
		str := "\033[42"
		if err != nil {
			str = "\033[41"
		}

		fmt.Printf("[%s] " + str + ";37;1mOrm\033[0m %s [%fs]\n", date("m-d H:i:s", s.queryStart), cacheKey, s.queryTime)
	}()

	s.queryStart = time.Now()

	v := reflect.ValueOf(obj)
	n := reflect.New(s.table.Type)
	i := reflect.Indirect(n)

	// get from cache
	res := s.orm.cacheHandler.Get(cacheKey)
	if res != nil && len(res) > 0 {
		var d []interface{}
		err = msgpack.Unmarshal(res, &d)

		if err != nil {
			return
		}

		for _, idx := range s.colIdx {
			convertAssign(v.Elem().Field(idx).Addr().Interface(), d[idx])
		}

		return
	}

	// get from database
	sn := &Session{
		orm:         s.orm,
		tx:          s.tx,
		enableCache: false,
		options:     s.options,
		table:       s.table,
		groupBy:     s.groupBy,
		orderBy:     s.orderBy,
		where:       s.where,
		columns:     s.table.Fields,
		limit:       1,
	}

	var rows *sql.Rows
	rows, err = sn.Query()
	if err != nil {
		return
	}

	sn = nil

	defer rows.Close()

	pointers := make([]interface{}, i.NumField())
	for j := 0; j < i.NumField(); j++ {
		pointers[j] = i.Field(j).Addr().Interface()
	}

	var find bool
	if rows.Next() {
		find = true
		rows.Scan(pointers...)
	}

	// TODO: empty data cache
	if !find {
		return
	}

	d := make([]interface{}, i.NumField())
	for j := 0; j < i.NumField(); j++ {
		d[j] = i.Field(j).Interface()
	}

	for _, idx := range s.colIdx {
		convertAssign(v.Elem().Field(idx).Addr().Interface(), d[idx])
	}

	var val []byte
	val, err = msgpack.Marshal(d)
	if err == nil {
		s.orm.cacheHandler.Set(cacheKey, val, s.orm.cacheTime)
	}

	return
}

func (s *Session) isQueryAllFields() bool {
	if len(s.columns) > 1 {
		return false
	}

	return s.columns[0] == "*"
}

func (s *Session) getCleanKey() string {
	// has raw
	for _, v := range s.columns {
		if strings.Contains(v, "(") || strings.Contains(v, " ") {
			return ""
		}
	}

	if s.options != nil || s.orderBy != nil || s.groupBy != nil || s.having != nil || s.where == nil {
		return ""
	}

	// 获取到全部主键和唯一字段
	usedFields := s.getUsedFields()

	where := map[string]interface{}{}
	for _, ws := range s.where {
		if ws.operator != "=" {
			return ""
		}

		// 存在非主键或唯一字段
		if _, ok := usedFields[ws.column]; !ok {
			return ""
		}

		where[ws.column] = ws.value
	}

	l := len(where)

	// 主键
	if v, ok := where[s.table.PrimaryKey]; ok {
		if l > 1 {
			return ""
		}

		return fmt.Sprintf(s.orm.primaryCacheKey, s.table.Name, v)
	}

	// 唯一
	if len(s.table.UniqueKeys) == 0 {
		return ""
	}

	for k, v := range s.table.UniqueKeys {
		if len(v) != l {
			continue
		}

		uniques := []interface{}{}

		for _, val := range v {
			if value, ok := where[val]; ok {
				uniques = append(uniques, value)
				continue
			}

			uniques = []interface{}{}
			break
		}

		if len(uniques) == 0 {
			continue
		}

		t := k + ":"
		for i := 0; i < l; i++ {
			if t != "" {
				t += "&"
			}

			t += "%v"
		}

		return fmt.Sprintf(s.orm.uniqueCacheKey, s.table.Name, fmt.Sprintf(t, uniques...))
	}

	return ""
}

func convertAssign(dest, src interface{}) error {
	switch d := dest.(type) {
	case *string:
		*d = src.(string)
	case *[]byte:
		*d = src.([]byte)
	case *bool:
		*d = src.(bool)
	default:
		dv := reflect.Indirect(reflect.ValueOf(dest))

		switch dv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			s := strconv.FormatInt(reflect.ValueOf(src).Int(), 10)
			i64, err := strconv.ParseInt(s, 10, dv.Type().Bits())
			if err != nil {
				return strconvErr(src, s, dv.Kind(), err)
			}
			dv.SetInt(i64)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			s := strconv.FormatUint(reflect.ValueOf(src).Uint(), 10)
			u64, err := strconv.ParseUint(s, 10, dv.Type().Bits())
			if err != nil {
				return strconvErr(src, s, dv.Kind(), err)
			}
			dv.SetUint(u64)
		case reflect.Float32, reflect.Float64:
			var s string
			rv := reflect.ValueOf(src)
			switch rv.Kind() {
			case reflect.Float64:
				s = strconv.FormatFloat(rv.Float(), 'g', -1, 64)
			case reflect.Float32:
				s = strconv.FormatFloat(rv.Float(), 'g', -1, 32)
			}

			f64, err := strconv.ParseFloat(s, dv.Type().Bits())
			if err != nil {
				return strconvErr(src, s, dv.Kind(), err)
			}
			dv.SetFloat(f64)
		default:
			return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, dest)
		}
	}

	return nil
}

func strconvErr(src interface{}, s string, kind reflect.Kind, err error) error {
	if ne, ok := err.(*strconv.NumError); ok {
		err = ne.Err
	}

	return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, kind, err)
}

func (s *Session) getUsedFields() map[string]bool {
	usedFields := map[string]bool{s.table.PrimaryKey: true}
	if len(s.table.UniqueKeys) > 0 {
		for _, keys := range s.table.UniqueKeys {
			for _, key := range keys {
				usedFields[key] = true
			}
		}
	}

	return usedFields
}

func (s *Session) makeCleanKey() {

}

func (s *Session) cleanCache() {

}
