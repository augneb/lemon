package lemon

import (
	"errors"
	"reflect"
	"time"
	"database/sql"
)

var (
	errSqlEmpty       = errors.New("sql is empty")
	errNeedPointer    = errors.New("needs a pointer")
	errNeedPtrToSlice = errors.New("needs a pointer to a slice")
)

func (s *Session) Query(sqlStr string, values []interface{}) (*sql.Rows, error) {
	if sqlStr == "" {
		return nil, errSqlEmpty
	}

	if s.sql == "" {
		s.sql = sqlStr
		s.args = values
	}

	s.queryStart = time.Now()

	var r *sql.Rows
	var e error
	if s.tx != nil {
		r, e = s.tx.Query(sqlStr, values...)
	} else {
		if s.useSlave == 0 {
			if s.orm.dbSlaveLen > 0 {
				s.useSlave = 1
			} else {
				s.useSlave = 2
			}
		}

		var db *sql.DB
		if s.useSlave == 2 {
			db = s.orm.db
		} else {
			db = s.orm.getSlave()
		}

		r, e = db.Query(sqlStr, values...)
	}

	s.after(e == nil)

	return r, e
}

func (s *Session) Get(obj interface{}, to ...*map[string]interface{}) (find bool, err error) {
	defer s.reset()

	v := reflect.ValueOf(obj)

	// 检测是否指针
	if v.Kind() != reflect.Ptr {
		return find, errors.New("need a pointer, given: " + v.Kind().String())
	}

	// 检测是否结构体
	if k := v.Elem().Kind(); k != reflect.Struct {
		return find, errors.New("element need a struct, given: " + k.String())
	}

	s.Limit(1)

	if s.table == nil {
		s.Table(s.orm.GetTableInfo(obj).Name)
	}

	// 分析需要查询的字段
	if s.colIdx == nil {
		s.parseSelectFields(s.table)
	}

	var cacheKey string

	// 是否开启缓存
	if s.enableCache {
		cacheKey = s.getCleanKey()
	}

	if cacheKey != "" {
		find, err = s.getFromCache(cacheKey, &v)
	} else {
		pointers := make([]interface{}, len(s.colIdx))
		for i, idx := range s.colIdx {
			pointers[i] = v.Elem().Field(idx).Addr().Interface()
		}

		var rows *sql.Rows
		rows, err = s.Query(s.GetSelectSql())
		if err != nil {
			return
		}

		if rows.Next() {
			err  = rows.Scan(pointers...)
			find = _if(err != nil, true, false).(bool)
		}

		rows.Close()
	}

	if err == nil && len(to) > 0 {
		for _, idx := range s.colIdx {
			(*to[0])[s.table.Fields[idx]] = v.Elem().Field(idx).Interface()
		}
	}

	return
}

func (s *Session) Find(obj interface{}) (find bool, err error) {
	defer s.reset()

	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		return find, errNeedPointer
	}

	sliceValue := reflect.Indirect(v)
	if sliceValue.Kind() != reflect.Slice {
		return find, errNeedPtrToSlice
	}

	elemType := sliceValue.Type().Elem()

	var isPtr bool
	if elemType.Kind() == reflect.Ptr {
		isPtr = true
		elemType = elemType.Elem()
	}

	ti := s.orm.GetTableInfo(elemType, true)

	if s.table == nil {
		s.Table(ti.Name)
	}

	if s.colIdx == nil {
		s.parseSelectFields(ti)
	}

	var rows *sql.Rows
	rows, err = s.Query(s.GetSelectSql())
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		find = true

		newValue := reflect.New(elemType)
		newIndirect := reflect.Indirect(newValue)

		pointers := []interface{}{}
		for _, field := range s.colIdx {
			pointers = append(pointers, newIndirect.Field(field).Addr().Interface())
		}

		err = rows.Scan(pointers...)
		if err != nil {
			return false, err
		}

		if isPtr {
			sliceValue.Set(reflect.Append(sliceValue, newValue.Elem().Addr()))
		} else {
			sliceValue.Set(reflect.Append(sliceValue, newValue.Elem()))
		}
	}

	return
}

func (s *Session) parseSelectFields(ti *structCache) {
	s.colIdx = []int{}

	check := true
	if s.columns == nil {
		check = false
	}

	columns := map[string]bool{}
	for _, v := range s.columns {
		columns[v] = true
	}

	for i, f := range ti.Fields {
		if check {
			if _, ok := columns[f]; !ok {
				continue
			}
		} else {
			s.columns = append(s.columns, f)
		}

		s.colIdx = append(s.colIdx, i)
	}
}
