package dialect

import "reflect"

var dialectsMap = map[string]Dialect{}

// Dialect 各种数据库方言接口
type Dialect interface {
	// DataType 将 go 类型转换为对应数据库的列类型
	DataType(typ reflect.Value) string

	// TableExistSql 判断表是否存在的 SQL 及参数
	TableExistSql(tableName string) (string, []interface{})

	// TableColumnsSql 查询表当前已有列（列名放在结果第一列，列类型放在第二列）
	TableColumnsSql(tableName string) (string, []interface{})

	// AlterColumnSql 生成修改列类型的 SQL
	AlterColumnSql(tableName, column, newType string) string

	// DropColumnSql 生成删除列的 SQL
	DropColumnSql(tableName, column string) string

	// SupportsAlterColumn 是否支持修改列类型
	// 若为 false，Sync 将跳过列类型变更
	SupportsAlterColumn() bool
}

// RegisterDialect 将方言注册进全局字典 dialectsMap
func RegisterDialect(name string, dialet Dialect) {
	dialectsMap[name] = dialet
}

// GetDialect 从全局字典 dialectsMap 获取方言
func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectsMap[name]
	return
}
