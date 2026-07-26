<img src="logo.png" style="zoom:15%;" />

## Borm 介绍

**一个用一条 DSN 就能跑起来的轻量 Go ORM。** 自动识别数据库、自动同步表结构、钩子默认开启，
查询与写入统一到一个 `Query` 方法，ORM 与手写 SQL 无缝混用。

## 特性一览

- **一条 DSN 启动**：`borm.NewEngine(dsn)` 自动识别 MySQL / PostgreSQL / SQLite，内置驱动，无需手动导入
- **唯一的终结方法 `Query`**：不带参数即执行写操作，带参数按反射类型自动判别取一条 / 取多条
- **ORM 与 Raw 混用**：链式条件（`Where` / `OrderBy` / `Limit` / `Offset`）与手写 `Raw(...)` 共享同一套终结方法
- **`Sync` 自动同步表结构**：建表、加列、改类型、删除多余列一步到位
- **钩子默认开启**：`Before*` / `After*` 返回 error 即自动终止后续 SQL
- **一键事务**：`Transaction` 失败自动回滚

## 安装

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

// User 是表模型；结构体字段名即列名，borm tag 追加列约束
type User struct {
	Name string `borm:"PRIMARY KEY"`
	Age  int
}

func main() {
	// 只需要 DSN，自动识别数据库类型：
	//   sqlite:   "test.db"  或  "sqlite://test.db"
	//   mysql:    "root:root@tcp(127.0.0.1:3306)/test"
	//   postgres: "postgres://user:pass@127.0.0.1:5432/test?sslmode=disable"
	engine, err := borm.NewEngine("test.db")
	if err != nil {
		panic(err)
	}
	defer engine.Close()

	s := engine.NewSession().Model(&User{})

	// 自动同步表结构（建表 / 加列 / 改类型 / 删除多余列）
	_ = s.Sync()

	// 插入一条或多条
	_, _ = s.Insert(
		&User{Name: "tomygin", Age: 20},
		&User{Name: "ice", Age: 19},
	)

	// 查询：Query 按 dest 反射类型自动判别取一条 / 取多条
	var u User
	_ = s.Where("Name = ?", "tomygin").Query(&u) // *Struct → 取一条

	var list []User
	_ = s.Where("Age > ?", 10).Limit(10).Offset(0).Query(&list) // *[]Struct → 取多条

	// 手写 SQL 同样用 Query，支持结构体 / 基础类型 / 各自的切片
	var u2 User
	_ = s.Raw("SELECT Name, Age FROM User WHERE Name = ?", "tomygin").Query(&u2)

	var count int
	_ = s.Raw("SELECT COUNT(*) FROM User").Query(&count)

	var names []string
	_ = s.Raw("SELECT Name FROM User").Query(&names)
	fmt.Println(names)

	// 写操作：Query 不带参数即执行（INSERT / UPDATE / DELETE / DDL）
	_ = s.Raw("UPDATE User SET Age = Age + 1 WHERE Name = ?", "tomygin").Query()

	// ORM 写操作：Update / Delete 返回受影响行数
	_, _ = s.Where("Name = ?", "tomygin").Update("Age", 21)
	_, _ = s.Where("Age = ?", 19).Limit(1).Delete()

	// 事务：回调返回 error 自动回滚，否则自动提交
	r, err := engine.Transaction(func(s *session.Session) (interface{}, error) {
		s.Model(&User{})
		_, _ = s.Insert(&User{Name: "tx_user", Age: 1})
		var t User
		return t, s.Where("Name = ?", "tx_user").Query(&t)
	})
	fmt.Println(r, err)
}

// 钩子默认开启：只要返回的 error 不为 nil，就会自动终止后续 SQL 执行
func (u *User) BeforeQuery(s *session.Session) error {
	return nil
}
```

> 更完整的、覆盖每个导出 API 的可运行示例见 [`_example/example.go`](_example/example.go)（`cd _example && go run example.go`）。

## 核心用法

### 结构体与 tag

字段名即列名（匹配时忽略大小写），`borm` tag 用于追加列约束，会原样拼进建表语句：

```go
type User struct {
	Name string `borm:"PRIMARY KEY"`
	Age  int    // 无 tag 时仅生成 "Age <类型>"
}
```

字段类型由方言（dialect）映射为对应数据库的列类型。

### 唯一的终结方法 Query

查询不再区分 `Get` / `First` / `All`，写操作也不再单列 `Run`，全部统一为一个 `Query` 方法，
根据**传入参数**自动决定行为：

- **不传参数** → 执行写操作（INSERT / UPDATE / DELETE / DDL），只返回 error
- **传入切片指针 `*[]T`** → 取多条
- **传入其它指针 `*Struct` / `*基础类型`** → 取一条

查询时 ORM 与 Raw 共享这一套方法；根据是否调用过 `Raw(...)` 自动选择模式：

| 用法                       | 行为                                         |
| -------------------------- | -------------------------------------------- |
| `s.Where(...).Query(&x)`   | ORM 取一条到 `*Struct`                       |
| `s.Where(...).Query(&xs)`  | ORM 取多条到 `*[]Struct`                     |
| `s.Raw(sql...).Query(&x)`  | 取一条，`*Struct` 或 `*int / *string / ...`  |
| `s.Raw(sql...).Query(&xs)` | 取多条，`*[]Struct` 或 `*[]int / *[]string`  |
| `s.Raw(sql...).Query()`    | 执行写操作（INSERT / UPDATE / DELETE / DDL） |

结构体按列名匹配字段（忽略大小写）。多段 `Raw` 可链式拼接，片段之间自动以空格连接：

```go
var users []User
s.Raw("SELECT Name, Age FROM User").
	Raw("WHERE Age BETWEEN ? AND ?", 10, 30).
	Raw("ORDER BY Age DESC").
	Query(&users)
```

### ORM 写操作

链式条件搭配 `Insert` / `Update` / `Delete`，三者都返回 `(受影响行数 int64, error)`：

```go
// Insert：一次插入一条或多条
n, _ := s.Insert(&User{Name: "a", Age: 1})                          // 单条
n, _ = s.Insert(&User{Name: "b", Age: 2}, &User{Name: "c", Age: 3}) // 多条

// Update 形式一：键值对 "col1", v1, "col2", v2 …
n, _ = s.Where("Name = ?", "a").Update("Age", 10)

// Update 形式二：map[string]interface{}，适合字段不固定的场景
n, _ = s.Where("Name = ?", "b").Update(map[string]interface{}{"Age": 20})

// Delete：配合 Where / Limit 删除匹配行
n, _ = s.Where("Age < ?", 5).Limit(1).Delete()

// Count：统计匹配行数
total, _ := s.Where("Age > ?", 1).Count()
```

> 写操作默认触发对应钩子（`BeforeInsert` / `AfterInsert` 等）；钩子返回 error 会终止本次 SQL。

### 钩子函数

在模型上实现对应方法即可，默认开启；返回非 nil error 会终止后续 SQL：

```go
func (u *User) BeforeInsert(s *session.Session) error {
	if u.Name == "" {
		return errors.New("name required")
	}
	return nil
}
```

可用的钩子：

```
BeforeQuery  / AfterQuery
BeforeInsert / AfterInsert
BeforeUpdate / AfterUpdate
BeforeDelete / AfterDelete
```

### 事务

高层 `Transaction`（推荐）失败自动回滚、成功自动提交：

```go
r, err := engine.Transaction(func(s *session.Session) (interface{}, error) {
	s.Model(&User{})
	_, _ = s.Insert(&User{Name: "tx", Age: 1})
	return nil, nil // 返回 error（或 panic）则整体回滚
})
```

也可用底层的 `Begin` / `Commit` / `RollBack` 手动控制。

### Sync 自动同步表结构

`Sync()` 会对比 Model 与数据库中的真实表结构并自动对齐：

| 情况                   | 处理方式                                       |
| ---------------------- | ---------------------------------------------- |
| 表不存在               | 自动 `CREATE TABLE`                            |
| Model 新增的列         | 自动 `ALTER TABLE ADD COLUMN`                  |
| 列类型发生变更         | 自动 `ALTER TABLE MODIFY/ALTER COLUMN`（见下） |
| Model 中已经不存在的列 | 自动 `ALTER TABLE DROP COLUMN`                 |

> 说明：SQLite 仅支持加列 / 删列（3.35+），不支持直接修改列类型，`Sync` 会跳过类型变更；MySQL / PostgreSQL 三种变更都支持。

### DSN 识别规则

`NewEngine` 根据 DSN 形态自动选择驱动：

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

## 未来计划

- [x] 支持钩子函数（默认开启，返回 error 自动终止）
- [x] 事务提交
- [x] 支持 mysql / postgres / sqlite
- [x] 内置驱动，DSN 自动识别
- [x] `Sync` 自动同步表结构（建表 + 加列 + 改类型 + 删列）
- [x] 唯一的终结方法 Query（不带参数即写操作；带参数按反射类型自动判别取一条 / 取多条）
