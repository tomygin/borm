package dialect

import (
	"fmt"
	"reflect"
	"time"
)

// TODO内置驱动

type mysql struct{}

var _ Dialect = (*mysql)(nil)

func init() {
	RegisterDialect("mysql", &mysql{})
}

func (m *mysql) DataType(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
		return "int"
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

func (m *mysql) TableExistSql(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}

	// TODOMySQL的查询表存在的语句
	return "SELECT name FROM mysql_master WHERE type = 'table' and name = ?", args
}
