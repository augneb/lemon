package lemon

import (
	"errors"
	"reflect"
	"time"
	"database/sql"
)

func (s *Session) Query() (*sql.Rows, error) {
	sqlStr, values := s.GetSelectSql()

	s.queryStart = time.Now()

	var r *sql.Rows
	var e error
	if s.tx != nil {
		r, e = s.tx.Query(sqlStr, values...)
	} else {
		r, e = s.orm.db.Query(sqlStr, values...)
	}

	s.after(e == nil)

	return r, e
}

func (s *Session) Get(obj interface{}) error {
	defer s.reset()

	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		return errors.New("needs a pointer, given:" + v.Kind().String())
	}

	s.Limit(1)

	if s.table == nil {
		s.Table(s.orm.GetTableInfo(obj).Name)
	}

	if s.colIdx == nil {
		s.parseSelectFields(s.table)
	}

	if s.enableCache {
		if cacheKey := s.getCleanKey(); cacheKey != "" {
			return s.getFromCache(cacheKey, obj)
		}
	}

	rows, err := s.Query()
	if err != nil {
		return err
	}

	defer rows.Close()

	pointers := make([]interface{}, len(s.colIdx))
	for i, idx := range s.colIdx {
		pointers[i] = v.Elem().Field(idx).Addr().Interface()
	}

	if rows.Next() {
		err = rows.Scan(pointers...)
	}

	return err
}

func (s *Session) Find(obj interface{}) error {
	defer s.reset()

	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		return errors.New("needs a pointer")
	}

	sliceValue := reflect.Indirect(v)
	if sliceValue.Kind() != reflect.Slice {
		return errors.New("needs a pointer to a slice")
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

	rows, err := s.Query()
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		newValue    := reflect.New(elemType)
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
