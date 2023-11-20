package session

import (
	"reflect"

	"github.com/tomygin/borm/log"
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

// CallMethod会调用Before,After系列的方法
// 如果value为nil调用的对象就是当前数据库的那个对象
// 否者是value对象作为调用的对象
func (s *Session) CallMethod(method string, value interface{}) {

	if !s.EnableHook {
		return
	}

	//找到当前表 结构体 的 method 方法
	fm := reflect.ValueOf(s.RefTable().Model).MethodByName(method)

	//如果有自定义结构体就不用表结构体
	if value != nil {
		fm = reflect.ValueOf(value).MethodByName(method)
	}

	param := []reflect.Value{reflect.ValueOf(s)}

	if fm.IsValid() {
		if v := fm.Call(param); len(v) > 0 {
			if err, ok := v[0].Interface().(error); ok {
				// panic(err)
				log.Error(err)
			}
		}
	}

}

// 在hook中如果执行失败可以调用来结束hook后相关的 增删查改
// func (s *Session) Abort() {
// 	s.abort = true
// }
