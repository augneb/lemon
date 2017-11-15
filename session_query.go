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

func (s *Session) Query(sqlStr string, values ...interface{}) (*sql.Rows, error) {
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

func (s *Session) Get(obj interface{}, to ...*map[string]interface{}) error {
	defer s.reset()

	v := reflect.ValueOf(obj)

	// 检测是否指针
	if v.Kind() != reflect.Ptr {
		return errors.New("needs a pointer, given: " + v.Kind().String())
	}

	// 检测是否结构体
	if k := v.Elem().Kind(); k != reflect.Struct {
		return errors.New("element needs a struct, given: " + k.String())
	}

	s.Limit(1)

	if s.table == nil {
		s.Table(s.orm.GetTableInfo(obj).Name)
	}

	// 分析需要查询的字段
	if s.colIdx == nil {
		s.parseSelectFields(s.table)
	}

	var err error
	var cacheKey string

	// 是否开启缓存
	if s.enableCache {
		// 命中缓存条件，从缓存中获取数据
		cacheKey = s.getCleanKey()
		if cacheKey != "" {
			err = s.getFromCache(cacheKey, &v)
		}
	}

	if cacheKey == "" {
		pointers := make([]interface{}, len(s.colIdx))
		for i, idx := range s.colIdx {
			pointers[i] = v.Elem().Field(idx).Addr().Interface()
		}

		var rows *sql.Rows

		sqlStr, values := s.GetSelectSql()
		rows, err = s.Query(sqlStr, values...)
		if err != nil {
			return err
		}

		if rows.Next() {
			err = rows.Scan(pointers...)
		}

		rows.Close()
	}

	if err == nil && len(to) > 0 {
		for _, idx := range s.colIdx {
			(*to[0])[s.table.Fields[idx]] = v.Elem().Field(idx).Interface()
		}
	}

	return err
}

func (s *Session) Find(obj interface{}) error {
	defer s.reset()

	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		return errNeedPointer
	}

	sliceValue := reflect.Indirect(v)
	if sliceValue.Kind() != reflect.Slice {
		return errNeedPtrToSlice
	}

	elemType := sliceValue.Type().Elem()

	var isPointer bool
	if elemType.Kind() == reflect.Ptr {
		isPointer = true
		elemType = elemType.Elem()
	}

	ti := s.orm.GetTableInfo(elemType, true)

	if s.table == nil {
		s.Table(ti.Name)
	}

	if s.colIdx == nil {
		s.parseSelectFields(ti)
	}

	sqlStr, values := s.GetSelectSql()
	rows, err := s.Query(sqlStr, values...)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		newValue := reflect.New(elemType)
		newIndirect := reflect.Indirect(newValue)

		pointers := []interface{}{}
		for _, field := range s.colIdx {
			pointers = append(pointers, newIndirect.Field(field).Addr().Interface())
		}

		err := rows.Scan(pointers...)
		if err != nil {
			return err
		}

		if isPointer {
			sliceValue.Set(reflect.Append(sliceValue, newValue.Elem().Addr()))
		} else {
			sliceValue.Set(reflect.Append(sliceValue, newValue.Elem()))
		}
	}

	return nil
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
