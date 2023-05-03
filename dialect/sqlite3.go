package dialect

import (
	"fmt"
	"reflect"
	"time"

	_ "github.com/mattn/go-sqlite3" //内置sqlite3
)

type sqlite3 struct{}

// 在编译期检测sqlite3结构体是否实现了Dialect接口
var _ Dialect = (*sqlite3)(nil)

func init() {
	RegisterDialect("sqlite3", &sqlite3{})
}

// DataType将go的数据类型转化为sqlite3的数据类型
func (s *sqlite3) DataType(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint16, reflect.Uint32:
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

// TableExistSql 生成表是否存在的sql语句
// 因为每个数据库判断表存在的语句不同
func (s *sqlite3) TableExistSql(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	return "SELECT name FROM sqlite_master WHERE type = 'table' and name = ?", args
}
