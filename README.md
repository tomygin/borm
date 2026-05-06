<img src="logo.png" style="zoom:15%;" />

## Borm 介绍

**一个用一条 DSN 就能跑起来的轻量 Go ORM，自动同步表结构、钩子默认开启、ORM 与 Raw 接口完全统一。**

## 安装最新版

```bash
go get -u github.com/tomygin/borm@latest
```

## 快速上手

```go
package main

import (
	"fmt"

	"github.com/tomygin/borm"
	"github.com/tomygin/borm/session"
)

type User struct {
	Name string `borm:"PRIMARY KEY"`
	Age  int
}

func main() {

// 只需要 DSN，自动识别数据库类型
	// sqlite:   "test.db"  或  "sqlite://test.db"
	// mysql:    "root:root@tcp(127.0.0.1:3306)/test"
	// postgres: "postgres://user:pass@127.0.0.1:5432/test?sslmode=disable"

engine, err := borm.NewEngine("test.db")
	if err != nil {
		panic(err)
	}
	defer engine.Close()

	s := engine.NewSession().Model(&User{})

	// 自动同步表结构（建表 / 加列 / 改类型 / 删除多余列）
	_ = s.Sync()

	// 插入
	_, _ = s.Insert(
		&User{Name: "tomygin", Age: 20},
		&User{Name: "ice", Age: 19},
	)


	var u User
	_ = s.Where("Name = ?", "tomygin").Get(&u)

	var list []User
	_ = s.Where("Age > ?", 10).Limit(10).Offset(0).All(&list)

	// Raw 风格 取一条 / 取多条 / 单值 / 基础类型切片
	var u2 User
	_ = s.Raw("SELECT Name, Age FROM User WHERE Name = ?", "tomygin").Get(&u2)

	var users []User
	_ = s.Raw("SELECT Name, Age FROM User WHERE Age > ?", 10).All(&users)

	var count int
	_ = s.Raw("SELECT COUNT(*) FROM User").Get(&count)

	var names []string
	_ = s.Raw("SELECT Name FROM User").All(&names)
	fmt.Println(names)

	// 写操作
	_ = s.Raw("UPDATE User SET Age = Age + 1 WHERE Name = ?", "tomygin").Run()

	// ORM 更新 / 删除
	_, _ = s.Where("Name = ?", "tomygin").Update("Age", 21)
	_, _ = s.Where("Age = ?", 19).Limit(1).Delete()

	// 事务：失败自动回滚
	r, err := engine.Transaction(func(s *session.Session) (interface{}, error) {
		s.Model(&User{})
		_ = s.Sync()
		_, _ = s.Insert(&User{Name: "tx_user", Age: 1})
		t := User{}
		return t, s.Where("Name = ?", "tx_user").Get(&t)
	})
	fmt.Println(r, err)
}

// 钩子函数默认开启
// 只要返回的 error 不为 nil，就会自动终止后续 SQL 执行
func (u *User) BeforeQuery(s *session.Session) error {
	return nil
}
```

### 可用的钩子函数

```go
BeforeQuery  / AfterQuery
BeforeUpdate / AfterUpdate
BeforeDelete / AfterDelete
BeforeInsert / AfterInsert
```

### DSN 识别规则

| DSN 示例                                            | 解析出的驱动 |
| --------------------------------------------------- | ------------ |
| `postgres://user:pass@host:5432/db?sslmode=disable` | `pgx`        |
| `postgresql://user:pass@host:5432/db`               | `pgx`        |
| `host=127.0.0.1 user=xx password=xx dbname=xx`      | `pgx`        |
| `mysql://user:pass@tcp(host:3306)/db`               | `mysql`      |
| `user:pass@tcp(host:3306)/db`                       | `mysql`      |
| `sqlite://path/to/file.db` / `sqlite3://file.db`    | `sqlite`     |
| `file:xxx.db`                                       | `sqlite`     |
| `path/to/file.db` / `*.sqlite` / `*.sqlite3`        | `sqlite`     |

### 统一的 Get / All / Run

ORM 和 Raw 共享同一套终结方法；根据是否调用过 `Raw(...)` 自动选择模式：

| 方法                     | ORM 模式             | Raw 模式                                     |
| ------------------------ | -------------------- | -------------------------------------------- |
| `s.Where(...).Get(&x)`   | 取一条到 `*Struct`   | ——                                           |
| `s.Where(...).All(&xs)`  | 取多条到 `*[]Struct` | ——                                           |
| `s.Raw(sql...).Get(&x)`  | ——                   | 取一条，`*Struct` 或 `*int / *string / ...`  |
| `s.Raw(sql...).All(&xs)` | ——                   | 取多条，`*[]Struct` 或 `*[]int / *[]string`  |
| `s.Raw(sql...).Run()`    | ——                   | 执行写操作（INSERT / UPDATE / DELETE / DDL） |

结构体按列名匹配字段，忽略大小写。多段 `Raw` 可链式拼接，之间自动以空格连接：

```go
s.Raw("SELECT Name, Age FROM User").
  Raw("WHERE Age BETWEEN ? AND ?", 10, 30).
  Raw("ORDER BY Age DESC").
  All(&users)
```

### Sync 自动同步表结构

`Sync()` 会对比 Model 与数据库中的真实表结构：

| 情况                   | 处理方式                                       |
| ---------------------- | ---------------------------------------------- |
| 表不存在               | 自动 `CREATE TABLE`                            |
| Model 新增的列         | 自动 `ALTER TABLE ADD COLUMN`                  |
| 列类型发生变更         | 自动 `ALTER TABLE MODIFY/ALTER COLUMN`（见下） |
| Model 中已经不存在的列 | 自动 `ALTER TABLE DROP COLUMN`                 |

> 说明：SQLite 仅支持加列/删列（3.35+），不支持直接修改列类型-元素亲和，`Sync` 会跳过类型变更。MySQL / PostgreSQL 三种变更都支持。

## 未来计划

- [x] 支持钩子函数（默认开启，返回 error 自动终止）
- [x] 事务提交
- [x] 支持 mysql / postgres / sqlite
- [x] 内置驱动，DSN 自动识别
- [x] `Sync` 自动同步表结构（建表 + 加列 + 改类型 + 删列）
- [x] 简洁且统一的 Get / All / Run 接口
