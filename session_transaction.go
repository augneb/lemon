package lemon

import "errors"

var (
	errTransExist    = errors.New("transaction has exist")
	errTransNotExist = errors.New("transaction not exist")
)

func (s *Session) Begin() error {
	if s.tx != nil {
		return errTransExist
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
		return errTransNotExist
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
