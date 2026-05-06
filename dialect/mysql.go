package dialect

import (
	"fmt"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type mysql struct{}

var _ Dialect = (*mysql)(nil)

func init() {
	RegisterDialect("mysql", &mysql{})
}

// DataType 将 go 类型转为 mysql 类型
func (m *mysql) DataType(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "tinyint(1)"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return "int unsigned"
	case reflect.Int64:
		return "bigint"
	case reflect.Uint64:
		return "bigint unsigned"
	case reflect.Float32:
		return "float"
	case reflect.Float64:
		return "double"
	case reflect.String:
		return "varchar(255)"
	case reflect.Array, reflect.Slice:
		return "blob"
	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "datetime"
		}
	}
	panic(fmt.Sprintf("invalid sql type %s (%s)", typ.Type().Name(), typ.Kind()))
}

// TableExistSql 表是否存在
func (m *mysql) TableExistSql(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	return "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?", args
}

// TableColumnsSql 查询列名和类型
func (m *mysql) TableColumnsSql(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	return "SELECT COLUMN_NAME, COLUMN_TYPE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?", args
}

// AlterColumnSql 修改列类型
func (m *mysql) AlterColumnSql(tableName, column, newType string) string {
	return fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s %s", tableName, column, newType)
}

// DropColumnSql 删除列
func (m *mysql) DropColumnSql(tableName, column string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, column)
}

// SupportsAlterColumn mysql 支持修改列类型
func (m *mysql) SupportsAlterColumn() bool { return true }
