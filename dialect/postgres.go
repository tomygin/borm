package dialect

import (
	"fmt"
	"reflect"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type postgres struct{}

var _ Dialect = (*postgres)(nil)

func init() {
	RegisterDialect("pgx", &postgres{})
	RegisterDialect("postgres", &postgres{})
}

// DataType 将 go 类型转为 postgres 类型
func (p *postgres) DataType(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return "integer"
	case reflect.Int64, reflect.Uint64:
		return "bigint"
	case reflect.Float32:
		return "real"
	case reflect.Float64:
		return "double precision"
	case reflect.String:
		return "text"
	case reflect.Array, reflect.Slice:
		return "bytea"
	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "timestamp"
		}
	}
	panic(fmt.Sprintf("invalid sql type %s (%s)", typ.Type().Name(), typ.Kind()))
}

// TableExistSql 表是否存在
func (p *postgres) TableExistSql(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	return "SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename = $1", args
}

// TableColumnsSql 查询列名和类型
func (p *postgres) TableColumnsSql(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	return "SELECT column_name, data_type FROM information_schema.columns WHERE table_schema = 'public' AND table_name = $1", args
}

// AlterColumnSql 修改列类型
func (p *postgres) AlterColumnSql(tableName, column, newType string) string {
	return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s", tableName, column, newType)
}

// DropColumnSql 删除列
func (p *postgres) DropColumnSql(tableName, column string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, column)
}

// SupportsAlterColumn postgres 支持修改列类型
func (p *postgres) SupportsAlterColumn() bool { return true }
