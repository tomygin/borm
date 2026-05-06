package session

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/tomygin/borm/schema"
)

// Model 如果当前对象没有被解析为 Schema 就解析
func (s *Session) Model(value interface{}) *Session {
	if s.refTable == nil || reflect.TypeOf(value) != reflect.TypeOf(s.refTable.Model) {
		s.refTable = schema.Parse(value, s.dialect)
	}
	return s
}

// RefTable 用于获取 Session 中的 Schema
func (s *Session) RefTable() *schema.Schema {
	if s.refTable == nil {
		panic(errors.New("Model is not set"))
	}
	return s.refTable
}

// CreateTable 根据 Model 创建表
func (s *Session) CreateTable() error {
	table := s.RefTable()
	var col []string
	for _, field := range table.Fields {
		col = append(col, fmt.Sprintf("%s %s %s", field.Name, field.Type, field.Tag))
	}

	desc := strings.Join(col, ",")
	_, err := s.Raw(fmt.Sprintf("CREATE TABLE %s (%s);", table.Name, desc)).Exec()
	return err
}

// DropTable 删除表（若存在）
func (s *Session) DropTable() error {
	_, err := s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s", s.RefTable().Name)).Exec()
	return err
}

// IsExistTable 判断表是否存在
func (s *Session) IsExistTable() bool {
	sqlStr, values := s.dialect.TableExistSql(s.RefTable().Name)
	row := s.Raw(sqlStr, values...).QueryRow()
	if row == nil {
		return false
	}
	var tmp string
	_ = row.Scan(&tmp)
	return tmp == s.RefTable().Name
}

// existingColumns 读取当前数据库中此表已有的列：name -> type（列类型字符串）
// key 统一小写便于比对
func (s *Session) existingColumns() (map[string]string, error) {
	sqlStr, values := s.dialect.TableColumnsSql(s.RefTable().Name)
	rows, err := s.Raw(sqlStr, values...).QueryRows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := map[string]string{}
	for rows.Next() {
		var name, typ string
		if err := rows.Scan(&name, &typ); err != nil {
			return nil, err
		}
		cols[strings.ToLower(name)] = strings.ToLower(typ)
	}
	return cols, rows.Err()
}

// Sync 自动同步表结构
//
//   - 表不存在：自动 CreateTable
//   - 表已存在：
//     1) 补齐 Model 中新增的列（ALTER TABLE ADD COLUMN）
//     2) 列类型不一致则修改（方言支持时；sqlite 会跳过）
//     3) 删除数据库中存在而 Model 中不存在的多余列
func (s *Session) Sync() error {
	if s.refTable == nil {
		return errors.New("model is not set, please call Model() first")
	}

	if !s.IsExistTable() {
		return s.CreateTable()
	}

	existing, err := s.existingColumns()
	if err != nil {
		return err
	}

	table := s.RefTable()
	modelCols := map[string]struct{}{}

	// 1. 加列 / 改类型
	for _, f := range table.Fields {
		lname := strings.ToLower(f.Name)
		modelCols[lname] = struct{}{}

		curType, ok := existing[lname]
		if !ok {
			// 缺失：补加
			addSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table.Name, f.Name, f.Type)
			if _, err := s.Raw(addSQL).Exec(); err != nil {
				return fmt.Errorf("sync add column %s failed: %w", f.Name, err)
			}
			continue
		}

		// 已存在：若方言支持修改列类型，且类型明显不同，则 ALTER
		if s.dialect.SupportsAlterColumn() {
			want := strings.ToLower(f.Type)
			if !columnTypeEqual(curType, want) {
				alterSQL := s.dialect.AlterColumnSql(table.Name, f.Name, f.Type)
				if alterSQL == "" {
					continue
				}
				if _, err := s.Raw(alterSQL).Exec(); err != nil {
					return fmt.Errorf("sync alter column %s failed: %w", f.Name, err)
				}
			}
		}
	}

	// 2. 删除 Model 中已经不存在的列
	for colName := range existing {
		if _, ok := modelCols[colName]; ok {
			continue
		}
		dropSQL := s.dialect.DropColumnSql(table.Name, colName)
		if dropSQL == "" {
			continue
		}
		if _, err := s.Raw(dropSQL).Exec(); err != nil {
			return fmt.Errorf("sync drop column %s failed: %w", colName, err)
		}
	}

	return nil
}

// columnTypeEqual 对比两个列类型字符串是否"等价"
// 由于不同数据库反射出来的类型和我们声明的类型不完全一致，比如
//
//	mysql: int(11) vs int
//	postgres: "double precision" vs "double precision"
//
// 这里做一个宽松匹配：前缀去掉长度修饰即可
func columnTypeEqual(a, b string) bool {
	norm := func(s string) string {
		s = strings.ToLower(strings.TrimSpace(s))
		// 去掉括号修饰, 比如 int(11), varchar(255)
		if i := strings.Index(s, "("); i >= 0 {
			s = s[:i]
		}
		return strings.TrimSpace(s)
	}
	return norm(a) == norm(b)
}
