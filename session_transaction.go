package lemon

import "errors"

func (s *Session) Begin() error {
	if s.tx != nil {
		return errors.New("transaction has exist")
	}

	tx, err := s.orm.db.Begin()
	if err != nil {
		return err
	}

	s.tx = tx

	return nil
}

func (s *Session) Rollback() error {
	return s.transaction("rollback")
}

func (s *Session) Commit() error {
	return s.transaction("commit")
}

func (s *Session) transaction(t string) error {
	if s.tx == nil {
		return errors.New("transaction not start")
	}

	var err error
	if t == "rollback" {
		err = s.tx.Rollback()
	} else {
		err = s.tx.Commit()
	}

	s.tx = nil

	return err
}

