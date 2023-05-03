package session

import (
	"fmt"
)

func (s *Session) History() string {
	return s.history.String()
}

func (s *Session) recordSql(sql string, vars interface{}) {

	s.history.WriteString(sql)
	s.history.WriteString(fmt.Sprintf("%v", vars))
	s.history.WriteString("\n")
}
