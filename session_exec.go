package lemon

import (
	"time"
	"database/sql"
)

func (s *Session) Update() (int64, error) {
	defer s.reset()

	sqlStr, values := s.GetUpdateSql()

	res, err := s.Statement(sqlStr, values...)
	if err != nil {
		return 0, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (s *Session) Delete() (int64, error) {
	defer s.reset()

	sqlStr, values := s.GetDeleteSql()

	res, err := s.Statement(sqlStr, values...)
	if err != nil {
		return 0, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
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
