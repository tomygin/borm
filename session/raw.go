package session

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/tomygin/borm/clause"
	"github.com/tomygin/borm/dialect"
	"github.com/tomygin/borm/log"
	"github.com/tomygin/borm/schema"
)

type Session struct {
	db      *sql.DB //从engine那里获取来
	sql     strings.Builder
	sqlVars []interface{}

	dialect dialect.Dialect //适配不同的sql语言
	clause  clause.Clause   //构造sql语句

	refTable *schema.Schema //不同结构体反射的Schema对象

	history strings.Builder //用于记录历史执行了的sql语句
	tx      *sql.Tx         //事务

	// 在钩子函数中关闭后续操作
	Abort bool
	// 开启sql语句历史记录，默认关闭
	EnableHistory bool
	// 开启钩子函数，默认关闭
	EnableHook bool
}

// 为了对事务的支持

type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

// DB如果有事务就返回 *sql.Tx ，否者返回*sql.DB
func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

// New生成一个新的Session
func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{
		db:      db,
		dialect: dialect,
	}
}

// Clear将会把一个Session还原为新的Session，但保留基本配置
func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
	s.clause = clause.Clause{}
	s.Abort = false
}

// Raw将sql语句和变量保存在Session中
func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

// Exec会打印日志，然后执行Session中的sql语句和变量
// 最后会清理Session中的sql语句和变量
func (s *Session) Exec() (resout sql.Result, err error) {
	defer s.Clear()
	if s.Abort {
		err = errors.New("Abort")
		log.Error("Abort: ", s.sql.String(), s.sqlVars)
		return
	}
	log.Info(s.sql.String(), s.sqlVars)
	if s.EnableHistory {
		s.recordSql(s.sql.String(), s.sqlVars)
	}
	if resout, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}

func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	if s.Abort {

		log.Error("Abort: ", s.sql.String(), s.sqlVars)
		return nil
	}
	log.Info(s.sql.String(), s.sqlVars)
	if s.EnableHistory {
		s.recordSql(s.sql.String(), s.sqlVars)
	}
	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()
	if s.Abort {
		err = errors.New("Abort")
		log.Error("Abort: ", s.sql.String(), s.sqlVars)
		return
	}
	log.Info(s.sql.String(), s.sqlVars)
	if s.EnableHistory {
		s.recordSql(s.sql.String(), s.sqlVars)
	}
	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}
