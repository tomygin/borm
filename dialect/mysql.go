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

func (m *mysql) DataType(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "tinyint(1)" // MySQL 通常用 tinyint(1) 表示布尔
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return "int" // 32位整数
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return "int unsigned" // 无符号32位整数
	case reflect.Int64:
		return "bigint" // 64位整数
	case reflect.Uint64:
		return "bigint unsigned" // 无符号64位整数
	case reflect.Float32:
		return "float" // 单精度浮点数
	case reflect.Float64:
		return "double" // 双精度浮点数
	case reflect.String:
		return "varchar(255)" // 默认字符串长度
	case reflect.Array, reflect.Slice:
		return "blob" // 二进制数据
	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "datetime" // 时间类型
		}
	}

	panic(fmt.Sprintf("invalid sql type %s (%s)", typ.Type().Name(), typ.Kind()))
}

// 修正后的表存在判断语句
func (m *mysql) TableExistSql(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	// 使用 MySQL 的 information_schema 系统表
	return "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?", args
}
