package schema

import (
	"go/ast"
	"reflect"

	"github.com/tomygin/box/dialect"
)

// 数据库里面的字段
type Field struct {
	Name string
	Type string
	Tag  string
}

type Schema struct {
	Model interface{} //数据库表原型
	Name  string      //模型名字

	Fields     []*Field
	FieldNames []string //为了加快查找Field
	FieldMap   map[string]*Field
}

func (s *Schema) GetField(name string) *Field {
	return s.FieldMap[name]
}

// 将任意对象转化为 Schema
// dest为要转化为Schema的结构体
// dialect为每个字段提供数据类型转换服务
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()

	schema := &Schema{
		Model:    dest,
		Name:     modelType.Name(),
		FieldMap: map[string]*Field{},
	}
	// 将所有的字段都记录在Schema
	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		if !p.Anonymous && ast.IsExported(p.Name) {

			field := &Field{
				Name: p.Name,
				Type: d.DataType(reflect.Indirect(reflect.New(p.Type))),
			}

			if v, ok := p.Tag.Lookup("box"); ok {
				field.Tag = v
			}

			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.FieldMap[p.Name] = field
		}
	}
	return schema
}

// RecordValues将一个结构体的数据库字段的所有值获取，入参就是这个被获取字段的结构体
func (s *Schema) RecordValues(dest interface{}) interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	for _, field := range s.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}
