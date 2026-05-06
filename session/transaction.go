package session

func (s *Session) Begin() (err error) {
	s.tx, err = s.db.Begin()
	return
}

func (s *Session) Commit() (err error) {
	err = s.tx.Commit()
	return
}

func (s *Session) RollBack() (err error) {
	err = s.tx.Rollback()
	return
}
