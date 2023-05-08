<img src="logo.png" style="zoom:15%;" />

## box 介绍

这是一款轻量级的数据持久化库，用于爬虫后期工作，还在递归更新中，相信你能3分钟内上手，默认使用sqlite3数据库

### 更新或下载

```go
go get -u github.com/tomygin/box@latest
```

## 重要说明

今天2023年5月5日我用box进行开发某个项目，box在简单易用方面表现不错，灵活的代价就是繁琐

所以还是用XORM或GORM这种成熟ORM更加明智，当然如果你怀着一腔热血用box我是不会阻拦的

也许某一天也能更加成熟吧😁



## 快速上手

### session介绍

适用于经典关系型数据库的增删改查

```go
package main

import (
	"github.com/tomygin/box"
	"github.com/tomygin/box/log"
	"github.com/tomygin/box/session"
)

type User struct {
	Name string `box:"PRIMARY KEY"`
	Age  int
}

func main() {
	engine, _ := box.NewEngine("test.db")
	defer engine.Close()

	s := engine.NewSession().Model(&User{})

	// 开启钩子函数
	s.Options(session.OpenHook())

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
	s.Abort()

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

### cache介绍

适用于 kv 数据，但这里的 key 和 value 目前仅仅支持string类型，同时会自动创建一个Item表用于持久化缓存数据

```go
package main

import (
	"fmt"

	"github.com/tomygin/box"
)

func main() {
	engine, _ := box.NewEngine("test.db")
	defer engine.Close()

	// 设置缓存大小为 8 byte
	c := engine.NewCache(1 << 3)
	go c.Add("h0", "imok")
	go c.Add("h1", "imok")
	go c.Add("h2", "imok")

	c.Add("h3", "imok")
	c.Add("h4", "imok")
	c.Add("h5", "imok")
	c.Add("h6", "imok")

	//查看缓存区有多少条数据
	fmt.Println(c.Len())

	if data, isok := c.Get("h6"); isok {
		//data 是从缓存里面找到的数据
		fmt.Println(data)
	}

	// 清除缓存
	c.Flush()

	fmt.Println(c.Len())
	if data, isok := c.Get("h6"); isok {
		//data 是从数据库里面找到的数据
		fmt.Println(data)
	}

}

```





## 必要说明

1. 这个项目不定期更新，如果你愿意动手修复问题欢迎PR，我会积极合并
2. sqlite3的驱动包是C的底层，所以你需要确保你有gcc或者mingw，并且配置了环境变量，对于Windows可以直接点击[这里](https://github.com/tomygin/box/releases/tag/v1.0.0)下载
3. 历史记录默认开启，如果需要关闭请在你的代码里面添加` s.Options(session.CloseHistory())`
4. 钩子函数默认关闭，如果需要打开请在你的代码里面添加` s.Options(session.OpenHook())`

## box进度

### 未来计划

- [x] 支持钩子函数
- [x] 事务提交
- [x] 选项初始化
- [x] 分页
- [x] 钩子函数终止后续操作
- [x] 自动记录执行的sql语句
- [x] 异步插入
- [x] 爬虫数据缓冲保存

### 已修复

- 代码逻辑结构不清晰✔
- 偶然情况钩子函数失效✔
- 终端打印颜色跨平台失效✔

## License

box released under the [MIT-License](./LICENSE)



