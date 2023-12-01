<img src="logo.png" style="zoom:15%;" />

## borm 介绍

这是一款轻量级的数据持久化库，还在递归更新中，相信你能3分钟内上手，默认使用sqlite3数据库

## 更新或下载

```go
go get -u github.com/tomygin/borm@latest
```

## 快速上手

```go
package main

import (
	"github.com/tomygin/borm"
	"github.com/tomygin/borm/log"
	"github.com/tomygin/borm/session"
)

type User struct {
	Name string `borm:"PRIMARY KEY"`
	Age  int
}

func main() {
	engine, _ := borm.NewEngine("test.db")
	defer engine.Close()

	s := engine.NewSession().Model(&User{})

	// 开启钩子函数
	s.EnableHook = true
	// 增删表
	s.CreateTable()
	defer s.DropTable()

	// 判断表存在
	if s.IsExistTable() {
		log.Info("表存在")
	}

	// 插入操作
	if affect, err := s.Insert(
		&User{Name: "tomygin", Age: 20},
		&User{Name: "ice", Age: 19},
		&User{Name: "test", Age: 18},
		&User{Name: "t0", Age: 100},
		&User{Name: "t1", Age: 101},
		&User{Name: "t2", Age: 102},
		&User{Name: "t3", Age: 103},
		&User{Name: "t4", Age: 104},
		&User{Name: "t5", Age: 105},
		&User{Name: "t6", Age: 106}); err == nil {
		log.Info("成功插入", affect, "条数据")
	}

	// 单条查询
	tmp := User{}
	if err := s.Where("Name = ?", "tomygin").First(&tmp); err != nil {
		log.Error(err)
	}

	// 多条查询
	tmps := []User{}
	if err := s.Where("Age > 10").Find(&tmps); err == nil {
		log.Info("拿到数据", tmps)
	}

	// 分页查询
	// Page 仅仅是封装了 Limit 和 Offset
	if err := s.Where("Age > 10").Page(1, 2).Find(&tmps); err == nil {
		log.Info("分页查询到数据", tmps)
	}

	// 删除
	if _, err := s.Where("Age = ?", 18).Limit(1).Delete(); err != nil {
		log.Error(err)
	}

	// 更新
	s.Where("Name = ?", "tomygin").Update("Age", 18)

	// 查看更新
	s.Where("Name = ?", "tomygin").First(&tmp)
	log.Info(tmp)

	// 排序查找最小年龄
	s.OrderBy("Age DESC").First(&tmp)
	log.Info(tmp)

	// 执行原生SQL
	s.Raw("INSERT INTO User (`Name`)  VALUES (?) ", "RAW").Exec()

	// 一键事务，失败自动回滚
	r, err := engine.Transaction(func(s *session.Session) (interface{}, error) {
		// s 是新的会话，先前对外部会话的设置对此会话无效，如有需要请重新设置
		s.Model(&User{})
		s.CreateTable()
		s.Insert(&User{Name: "tomygin"})
		t := User{}
		err := s.Where("Name = ?", "tomygin").First(&t)
		return t, err
	})
	log.Info(r, err)

	// session的sql历史记录
	history := s.History()
	log.Info(history)

	// 日志分级
	log.SetLevel(log.ErrorLevel)

}

// 钩子函数
func (u *User) BeforeQuery(s *session.Session) error {
	log.Info("钩子函数运行成功")

	// 不希望最后执行sql
	s.Abort = true

	return nil
}

```

```go
// 可用的钩子函数
BeforeQuery  
AfterQuery   
BeforeUpdate 
AfterUpdate  
BeforeDelete 
AfterDelete  
BeforeInsert 
AfterInsert  
```

## 必要说明

1. 历史记录默认关闭，如果需要打开请在你的代码里面添加` s.EnableHistory = true`
2. 钩子函数默认关闭，如果需要打开请在你的代码里面添加` s.EnableHook = true`

## 未来计划

- [x] 支持钩子函数
- [x] 事务提交
- [x] 选项初始化
- [x] 分页
- [x] 钩子函数终止后续操作
- [x] 自动记录执行的sql语句
- [x] 异步插入
- [x] 爬虫数据缓冲保存
- [ ] ~~从新实现注册回调函数~~
- [ ] ~~支持mysql~~

## borm日志

- 2023年11月6日 由`box`更名为`borm` ，去除cache。
- 2023年11月20日 去除不必要的选项卡初始化
- 2023年12月1日 暂时归档

## License

borm learn from [GEEKTUTU](https://geektutu.com/post/geeorm.html)
released under the [MIT-License](./LICENSE)



