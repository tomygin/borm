package clause

import "strings"

type Type int

const (
	INSERT Type = iota + 1
	VALUES
	SELECT
	LIMIT
	OFFSET
	WHERE
	ORDERBY
	UPDATE
	DELETE
	COUNT
)

// Clause用于记录生成的子sql语句
// 比如 Limit 1
// 这里的sql是map的原因可以生成  VALUES （？，？，？） ， a,b,c
type Clause struct {
	sql     map[Type]string
	sqlVars map[Type][]interface{}
}

// Set用于给Clause里面添加子句
func (c *Clause) Set(name Type, vars ...interface{}) {
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVars = make(map[Type][]interface{})
	}
	sql, vars := generators[name](vars...)
	c.sql[name] = sql
	c.sqlVars[name] = vars
}

// Build的作用是将所有的子句sql拼接为一个完整的sql语句
// oeders是需要提取的子句sql，并且生成的完整sql也是按照这个顺序生成的
// 比如 INSERT VALUES 最后生成 INSET INTO TABLENAME (col1,col2) , (vaule1_1,vaule2_1),(value1_2,value2_2)
// Build执行完后会清空Clause，然后实现复用
func (c *Clause) Build(orders ...Type) (string, []interface{}) {

	defer func() {
		c.sql = nil
		c.sqlVars = nil
	}()

	var sqls []string
	var vars []interface{}
	for _, order := range orders {
		if sql, ok := c.sql[order]; ok {
			sqls = append(sqls, sql)
			vars = append(vars, c.sqlVars[order]...)
		}
	}
	return strings.Join(sqls, " "), vars
}
