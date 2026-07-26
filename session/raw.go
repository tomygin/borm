package session

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/tomygin/borm/clause"
	"github.com/tomygin/borm/dialect"
	"github.com/tomygin/borm/schema"
)

// Session 是一次会话
// 钩子函数默认开启，若钩子返回非 nil 的 error，则后续 SQL 自动终止
type Session struct {
	db      *sql.DB //从 engine 那里获取来
	sql     strings.Builder
	sqlVars []interface{}

	dialect dialect.Dialect //适配不同的 sql 语言
	clause  clause.Clause   //构造 sql 语句

	refTable *schema.Schema //不同结构体反射的 Schema 对象

	tx *sql.Tx //事务

	// 钩子函数返回的错误（非 nil 即终止后续 sql 执行）
	hookErr error
}

// CommonDB 事务与普通连接的公共接口
type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

// DB 如果有事务就返回 *sql.Tx，否则返回 *sql.DB
func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

// New 生成一个新的 Session
func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{
		db:      db,
		dialect: dialect,
	}
}

// Clear 将会把一个 Session 还原为新的 Session，但保留基本配置
func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
	s.clause = clause.Clause{}
	s.hookErr = nil
}

// Raw 将 sql 语句和变量追加到当前会话的 SQL 缓冲区
// 多次调用会以空格连接，适合在 ORM 无法优雅表达的地方手写/拼接 SQL
//
// 使用方式：
//
//	s.Raw("UPDATE User SET Age = Age + 1 WHERE Name = ?", "tomygin").Query()
//	s.Raw("SELECT Name, Age FROM User WHERE Name = ?", "tomygin").Query(&u)
//	s.Raw("SELECT Name, Age FROM User WHERE Age > ?", 18).Query(&users)
func (s *Session) Raw(sql string, values ...interface{}) *Session {
	if s.sql.Len() > 0 {
		// 在片段之间自动加一个空格，避免粘连
		s.sql.WriteString(" ")
	}
	s.sql.WriteString(strings.TrimSpace(sql))
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

// exec 执行 Session 中的 sql 语句和变量
// 执行结束后会清理 Session 中的 sql 缓冲区
// 若钩子 Hook 返回了 error，则直接返回该错误而不执行
func (s *Session) exec() (result sql.Result, err error) {
	defer s.Clear()
	if s.hookErr != nil {
		return nil, s.hookErr
	}
	result, err = s.DB().Exec(s.sql.String(), s.sqlVars...)
	return
}

// queryRow 查询单行。若钩子返回错误则返回 nil
func (s *Session) queryRow() *sql.Row {
	defer s.Clear()
	if s.hookErr != nil {
		return nil
	}
	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

// queryRows 查询多行。若钩子返回错误则直接返回该错误
func (s *Session) queryRows() (rows *sql.Rows, err error) {
	defer s.Clear()
	if s.hookErr != nil {
		return nil, s.hookErr
	}
	rows, err = s.DB().Query(s.sql.String(), s.sqlVars...)
	return
}

// Query 是统一的终结方法，根据传入参数自动决定行为：
//
//  1. 不传参数 → 执行写操作（INSERT / UPDATE / DELETE / DDL），只返回 error
//     s.Raw("UPDATE User SET Age = Age + 1 WHERE Name = ?", "x").Query()
//     s.Raw("INSERT INTO User (Name, Age) VALUES (?, ?)", "a", 1).Query()
//
//  2. 传入切片指针 *[]T → 取多条
//     s.Where("Age > ?", 10).Query(&users)
//     s.Raw("SELECT Name, Age FROM User").Query(&users)
//
//  3. 传入其它指针 *Struct / *基础类型 → 取一条
//     s.Where("Name = ?", "tomygin").Query(&u)
//     s.Raw("SELECT COUNT(*) FROM User").Query(&count)
//
// 查询时 ORM / Raw 两种模式同样自动判断（是否调用过 Raw）。
//
// dest 支持：
//   - 切片指针：*[]Struct 或 *[]基础类型
//   - 结构体指针：按列名匹配字段（大小写不敏感）
//   - 基础类型指针：*int / *string / *float64 等
func (s *Session) Query(dest ...interface{}) error {
	// 不带参数：执行写操作
	if len(dest) == 0 {
		_, err := s.exec()
		return err
	}
	if len(dest) > 1 {
		return errors.New("Query: expects at most one dest argument")
	}

	d := dest[0]
	destV := reflect.ValueOf(d)
	if destV.Kind() != reflect.Ptr || destV.IsNil() {
		return errors.New("Query: dest must be a non-nil pointer")
	}

	// 根据 dest 的反射类型自动判别：切片指针取多条，否则取一条
	if destV.Elem().Kind() == reflect.Slice {
		return s.queryAll(d)
	}
	return s.queryOne(d)
}

// queryOne 查询并把单条记录写入 dest（Query 的取一条实现）
func (s *Session) queryOne(dest interface{}) error {
	// ORM 模式：自动触发钩子并构造 SELECT
	ormMode := !s.hasRaw()
	if ormMode {
		s.callMethod(beforeQuery, nil)
		defer s.callMethod(afterQuery, s.refTable.Model)

		// 如果用户传入的是结构体指针且没有 Model，就用它当 Model
		if s.refTable == nil {
			dv := reflect.ValueOf(dest)
			if dv.Kind() == reflect.Ptr && dv.Elem().Kind() == reflect.Struct {
				s.Model(reflect.New(dv.Elem().Type()).Interface())
			}
		}
		sqlStr, vars := s.buildSelect(nil)
		// 对 ORM 模式统一 Limit 1
		s.Raw(sqlStr, vars...)
		// 重新补上 LIMIT 1（Build 时 LIMIT 可能未设置）
		// 这里简单处理：若原 sql 中没有 LIMIT，则追加
		if !strings.Contains(strings.ToUpper(sqlStr), "LIMIT") {
			s.sql.WriteString(" LIMIT 1")
		}
	}

	rows, err := s.queryRows()
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	return scanInto(rows, cols, dest)
}

// queryAll 查询并把多条记录写入 dest（Query 的取多条实现）
//
// dest 必须是切片指针：
//   - *[]Struct
//   - *[]基础类型
func (s *Session) queryAll(dest interface{}) error {
	destV := reflect.ValueOf(dest)
	if destV.Kind() != reflect.Ptr || destV.Elem().Kind() != reflect.Slice {
		return errors.New("Query: dest must be a pointer to a slice")
	}
	sliceV := destV.Elem()
	elemT := sliceV.Type().Elem()

	ormMode := !s.hasRaw()
	if ormMode {
		// 自动触发钩子
		s.callMethod(beforeQuery, nil)
		defer s.callMethod(afterQuery, s.refTable.Model)

		// 若还没 Model，用切片元素类型当 Model
		if s.refTable == nil && elemT.Kind() == reflect.Struct {
			s.Model(reflect.New(elemT).Interface())
		}
		sqlStr, vars := s.buildSelect(nil)
		s.Raw(sqlStr, vars...)
	}

	rows, err := s.queryRows()
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	for rows.Next() {
		elemPtr := reflect.New(elemT)
		if err := scanInto(rows, cols, elemPtr.Interface()); err != nil {
			return err
		}
		sliceV.Set(reflect.Append(sliceV, elemPtr.Elem()))
	}
	return rows.Err()
}

// scanInto 将 rows 当前行的数据写入 dest
// dest 可以是 *struct 或 *基础类型
func scanInto(rows *sql.Rows, cols []string, dest interface{}) error {
	destV := reflect.ValueOf(dest)
	if destV.Kind() != reflect.Ptr || destV.IsNil() {
		return errors.New("scan target must be a non-nil pointer")
	}
	elem := destV.Elem()

	switch elem.Kind() {
	case reflect.Struct:
		// 按列名匹配结构体字段（大小写不敏感）
		scanArgs := make([]interface{}, len(cols))
		holders := make([]interface{}, len(cols)) // 兜底丢弃
		for i, col := range cols {
			f := findFieldByName(elem, col)
			if f.IsValid() && f.CanAddr() {
				scanArgs[i] = f.Addr().Interface()
			} else {
				var placeholder interface{}
				holders[i] = &placeholder
				scanArgs[i] = holders[i]
			}
		}
		return rows.Scan(scanArgs...)

	default:
		// 基础类型
		if len(cols) != 1 {
			return fmt.Errorf("scan into %s expects 1 column, got %d", elem.Kind(), len(cols))
		}
		return rows.Scan(dest)
	}
}

// findFieldByName 在结构体中根据列名找字段（忽略大小写）
func findFieldByName(structV reflect.Value, name string) reflect.Value {
	t := structV.Type()
	lname := strings.ToLower(name)
	for i := 0; i < t.NumField(); i++ {
		if strings.ToLower(t.Field(i).Name) == lname {
			return structV.Field(i)
		}
	}
	return reflect.Value{}
}
