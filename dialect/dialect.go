package dialect

import "reflect"

var dialectsMap = map[string]Dialect{}

type Dialect interface {
	DataType(typ reflect.Value) string
	TableExistSql(tableName string) (string, []interface{})
}

// RegisterDialect 将方言注册进全局字典dialectsMap
func RegisterDialect(name string, dialet Dialect) {
	dialectsMap[name] = dialet
}

// GetDialect 从全局字典dialectsMap获取方言
func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectsMap[name]
	return
}
