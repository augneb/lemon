package lemon

import (
	"time"
	"database/sql"
)

func (s *Session) Update() (int64, error) {
	return s.updateDelete(s.GetUpdateSql())
}

func (s *Session) Delete() (int64, error) {
	return s.updateDelete(s.GetDeleteSql())
}

func (s *Session) updateDelete(sqlStr string, values []interface{}) (int64, error) {
	defer s.reset()

	keys, _ := s.makeCleanKey()
	if keys != nil {
		s.cleanCache(keys)
	}

	res, err := s.Statement(sqlStr, values...)
	if err != nil {
		return 0, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	if keys != nil {
		s.cleanCache(keys)
	}

	return n, nil
}

func (s *Session) Insert() (int64, error) {
	defer s.reset()

	sqlStr, values := s.GetInsertSql()

	res, err := s.Statement(sqlStr, values...)
	if err != nil {
		return 0, err
	}

	n, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if s.orm.cacheEmpty {
		t := s.table.Name
		s.reset()
		s.Table(t).Where(s.table.PrimaryKey, n)

		if keys, _ := s.makeCleanKey(); keys != nil {
			s.cleanCache(keys)
		}
	}

	return n, nil
}

func (s *Session) Statement(sqlStr string, values ...interface{}) (sql.Result, error) {
	s.queryStart = time.Now()

	var r sql.Result
	var e error
	if s.tx != nil {
		r, e = s.tx.Exec(sqlStr, values...)
	} else {
		r, e = s.orm.db.Exec(sqlStr, values...)
	}

	s.after(e == nil)

	return r, e
}
