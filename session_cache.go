package lemon

import (
	"fmt"
	"reflect"
	"strings"
	"strconv"
	"time"
	"bytes"
	"database/sql"

	"github.com/vmihailenco/msgpack"
	"github.com/augneb/util"
)

func (s *Session) getFromCache(cacheKey string, v *reflect.Value) (err error) {
	defer func() {
		s.queryTime = time.Since(s.queryStart).Seconds()
		str := "\033[42"
		if err != nil {
			str = "\033[41"
		}

		util.Debug(fmt.Sprintf(str + ";37;1mOrm\033[0m %s [%fs]", cacheKey, s.queryTime), s.queryStart)
	}()

	s.queryStart = time.Now()

	// get from cache
	res := s.orm.cacheHandler.Get(cacheKey)

	if res != nil && len(res) > 0 {
		// empty cache
		if bytes.Equal(res, []byte(emptyCacheString)) {
			return
		}

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
		useSlave:    s.useSlave,
		enableCache: false,
		table:       s.table,
		where:       s.where,
		columns:     s.table.Fields,
		limit:       1,
	}

	var rows *sql.Rows

	sqlStr, values := sn.GetSelectSql()
	rows, err = sn.Query(sqlStr, values...)
	if err != nil {
		return
	}

	sn = nil

	l := len(s.table.Fields)

	d := make([]interface{}, l)
	p := make([]interface{}, l)
	for i := 0; i < l; i++ {
		p[i] = &d[i]
	}

	var find bool
	if rows.Next() {
		find = true
		err = rows.Scan(p...)
	}

	rows.Close()

	if err != nil {
		return
	}

	if !find {
		if s.orm.cacheEmpty {
			s.orm.cacheHandler.Set(cacheKey, []byte(emptyCacheString), s.orm.cacheTime)
		}

		return
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
	switch s := src.(type) {
	case string:
		switch d := dest.(type) {
		case *string:
			*d = s
		case *[]byte:
			*d = []byte(s)
		default:
			return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, dest)
		}

		return nil
	case []byte:
		switch d := dest.(type) {
		case *string:
			*d = string(s)
		case *interface{}:
			*d = util.CloneBytes(s)
		case *[]byte:
			*d = util.CloneBytes(s)
		default:
			return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, dest)
		}

		return nil
	case nil:
		switch d := dest.(type) {
		case *interface{}:
			*d = nil
		case *[]byte:
			*d = nil
		default:
			return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, dest)
		}

		return nil
	}

	switch d := dest.(type) {
	case *string:
		fmt.Println("->", src)
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

func (s *Session) makeCleanKey() (keys []string, err error) {
	usedFields := s.getUsedFields()

	columns := []string{}
	for k := range usedFields {
		columns = append(columns, k)
	}

	sn := &Session{
		orm:         s.orm,
		tx:          s.tx,
		useSlave:    s.useSlave,
		enableCache: false,
		options:     s.options,
		table:       s.table,
		orderBy:     s.orderBy,
		where:       s.where,
		columns:     columns,
		limit:       s.limit,
		offset:      s.offset,
	}

	var rows *sql.Rows

	sqlStr, values := sn.GetSelectSql()
	rows, err = sn.Query(sqlStr, values...)
	if err != nil {
		return
	}

	sn = nil

	defer rows.Close()

	l := len(columns)

	res := []map[string]interface{}{}
	if rows.Next() {
		d := map[string]interface{}{}

		t := make([]interface{}, l)
		p := make([]interface{}, l)
		for i := 0; i < l; i++ {
			p[i] = &t[i]
		}

		err = rows.Scan(p...)
		if err != nil {
			return
		}

		for i, f := range columns {
			d[f] = t[i]
		}

		res = append(res, d)
	}

	// 尝试从 where 条件中找到主键或者唯一
	if len(res) == 0 && len(s.where) > 0 {
		d := map[string]interface{}{}
		for _, item := range s.where {
			if item.operator != "=" {
				continue
			}

			if v, ok := usedFields[item.column]; ok {
				d[item.column] = v
			}
		}

		res = append(res, d)
	}

	if len(res) == 0 {
		return
	}

	for _, items := range res {
		if v, ok := items[s.table.PrimaryKey]; ok {
			pk := fmt.Sprintf(s.orm.primaryCacheKey, s.table.Name, v)
			keys = append(keys, pk)
		}

		if len(s.table.UniqueKeys) == 0 {
			continue
		}

		for k, v := range s.table.UniqueKeys {
			uniques := make([]interface{}, len(v))
			for i, f := range v {
				v, ok := items[f]
				if !ok {
					uniques = nil
					break
				}

				uniques[i] = v
			}

			if uniques == nil {
				continue
			}

			t := k + ":"
			for i := 0; i < l; i++ {
				if t != "" {
					t += "&"
				}

				t += "%v"
			}

			uk := fmt.Sprintf(s.orm.uniqueCacheKey, s.table.Name, fmt.Sprintf(t, uniques...))
			keys = append(keys, uk)
		}
	}

	return
}

func (s *Session) cleanCache(keys []string) {
	for _, key := range keys {
		s.orm.cacheHandler.Del(key)
	}
}

func strconvErr(src interface{}, s string, kind reflect.Kind, err error) error {
	if ne, ok := err.(*strconv.NumError); ok {
		err = ne.Err
	}

	return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, kind, err)
}
