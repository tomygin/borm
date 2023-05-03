package clause

import (
	"fmt"
	"strings"
)

// generator类型的函数专门用于生成构造sql子句，这里定义出来方便统一标准
type generator func(values ...interface{}) (string, []interface{})

// generators装载各种构建sql子句的函数，供Set生成子句保存在Clause里面
var generators map[Type]generator

// genBuildVars构建出 `?, ?, ?`
func genBuildVars(num int) string {
	var vars []string
	for i := 0; i < num; i++ {
		vars = append(vars, "?")
	}
	return strings.Join(vars, ", ")
}

// _insert的第一个参数是表名，后面的参数分别是数据库字段名
// 最后生成 INSTERT INTO TableName (col1,col2)
func _insert(values ...interface{}) (string, []interface{}) {
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("INSERT INTO %s (%v) ", tableName, fields), []interface{}{}
}

// _values的入参是一个二维空接口切片，一级切片确定每一行，二级切片确定每一行的字段
// 最后生成 VALUES (?,?...),(?,?...)	vars
func _values(values ...interface{}) (string, []interface{}) {
	var BuildStr string
	var sql strings.Builder
	var vars []interface{}

	sql.WriteString("VALUES ")

	for i, value := range values {
		v := value.([]interface{})

		if len(BuildStr) == 0 {
			BuildStr = genBuildVars(len(v))
		}

		// (?,?,?...)
		sql.WriteString(fmt.Sprintf("(%v)", BuildStr))
		// 如果不是最后一行就添加 ，
		if i+1 != len(values) {
			sql.WriteString(", ")
		}

		vars = append(vars, v...)
	}
	return sql.String(), vars
}

// _select第一个参数是表名，后面的参数分别是数据库字段名
// 最后生成 SELECT col1,col2 ... FROM TableName
func _select(values ...interface{}) (string, []interface{}) {
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("SELECT %v FROM %s ", fields, tableName), []interface{}{}
}

func _limit(values ...interface{}) (string, []interface{}) {
	return "LIMIT ?", values
}

func _offset(values ...interface{}) (string, []interface{}) {
	return "OFFSET ?", values
}

func _where(values ...interface{}) (string, []interface{}) {
	desc, vars := values[0], values[1:]
	return fmt.Sprintf("WHERE %s", desc), vars
}

// _orderBy只识别第一个参数作为排序的标准
// 最后生成 ORDER BY arg
func _orderBy(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("ORDER BY %s", values[0]), []interface{}{}
}

// _update的第一个参数是表名，第二个是map[string]interface{}保存需要更新的键值对
// 最后生成 UPDATE TableName SET field1 = ?,field2 =?   var1,var2
func _update(values ...interface{}) (string, []interface{}) {
	tableName := values[0]
	m := values[1].(map[string]interface{})
	var keys []string
	var vars []interface{}
	for k, v := range m {
		keys = append(keys, k+" = ?")
		vars = append(vars, v)
	}
	return fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(keys, ", ")), vars
}

// _delete只识别第一个参数作为从某个表删除
// 最后生成 DELETE FROM TableName
func _delete(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("DELETE FROM %s", values[0]), []interface{}{}
}

// _count唯一一个参数是表名
// 最后生成 SELECT TableName count(*)
func _count(values ...interface{}) (string, []interface{}) {
	return _select(values[0], []string{"count(*)"})
}

func init() {
	generators = make(map[Type]generator)
	generators[INSERT] = _insert
	generators[VALUES] = _values
	generators[LIMIT] = _limit
	generators[OFFSET] = _offset
	generators[WHERE] = _where
	generators[ORDERBY] = _orderBy
	generators[SELECT] = _select
	generators[UPDATE] = _update
	generators[DELETE] = _delete
	generators[COUNT] = _count
}
