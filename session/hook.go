package session

import (
	"reflect"
)

const (
	BeforeQuery  = "BeforeQuery"
	AfterQuery   = "AfterQuery"
	BeforeUpdate = "BeforeUpdate"
	AfterUpdate  = "AfterUpdate"
	BeforeDelete = "BeforeDelete"
	AfterDelete  = "AfterDelete"
	BeforeInsert = "BeforeInsert"
	AfterInsert  = "AfterInsert"
)

// CallMethod 会调用 Before, After 系列的钩子方法
// value 为 nil 时调用表 Model 自身的方法，否则调用 value 的方法
//
// 钩子函数签名: func (x *T) BeforeQuery(s *Session) error
// 只要返回的 error 不为 nil，就会自动终止后续 SQL 执行。
//
// 钩子函数默认开启，无需额外开关。
func (s *Session) CallMethod(method string, value interface{}) {
	// 若先前的钩子已经返回错误，则跳过后续所有钩子
	if s.hookErr != nil {
		return
	}

	if s.refTable == nil {
		return
	}

	// 找到当前表结构体的 method 方法
	fm := reflect.ValueOf(s.RefTable().Model).MethodByName(method)

	// 如果有自定义对象则优先在它上面查找
	if value != nil {
		fm = reflect.ValueOf(value).MethodByName(method)
	}

	if !fm.IsValid() {
		return
	}

	param := []reflect.Value{reflect.ValueOf(s)}
	ret := fm.Call(param)
	if len(ret) == 0 {
		return
	}
	if err, ok := ret[0].Interface().(error); ok && err != nil {
		// 一旦钩子返回错误，记录下来，后续 SQL 将自动终止
		s.hookErr = err
	}
}
