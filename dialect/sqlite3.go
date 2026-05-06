package dialect

import (
	"fmt"
	"reflect"
	"time"

	_ "modernc.org/sqlite"
)

type sqlite3 struct{}

var _ Dialect = (*sqlite3)(nil)

func init() {
	RegisterDialect("sqlite", &sqlite3{})
}

// DataType 将 go 的数据类型转化为 sqlite3 的数据类型
func (s *sqlite3) DataType(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return "integer"
	case reflect.Int64, reflect.Uint64:
		return "bigint"
	case reflect.Float32, reflect.Float64:
		return "real"
	case reflect.String:
		return "text"
	case reflect.Array, reflect.Slice:
		return "blob"
	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "datetime"
		}
	}
	panic(fmt.Sprintf("invalid sql type %s (%s) ", typ.Type().Name(), typ.Kind()))
}

// TableExistSql 生成表是否存在的 sql 语句
func (s *sqlite3) TableExistSql(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	return "SELECT name FROM sqlite_master WHERE type = 'table' and name = ?", args
}

// TableColumnsSql 查询 sqlite 表现有列名和类型
// 第一列：列名；第二列：列类型
func (s *sqlite3) TableColumnsSql(tableName string) (string, []interface{}) {
	// PRAGMA 不支持占位符，表名拼接；表名来自 Model，可信
	return "SELECT name, type FROM pragma_table_info('" + tableName + "')", nil
}

// AlterColumnSql sqlite 不支持直接修改列类型，SupportsAlterColumn 返回 false
// 这里保留占位实现
func (s *sqlite3) AlterColumnSql(tableName, column, newType string) string {
	return ""
}

// DropColumnSql sqlite 3.35+ 支持 DROP COLUMN
func (s *sqlite3) DropColumnSql(tableName, column string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, column)
}

// SupportsAlterColumn sqlite 不支持修改列类型
func (s *sqlite3) SupportsAlterColumn() bool { return false }
