// example.go 展示 borm 所有可导出的 API 用法
//
// 运行方式：
//
//	go run example.go
//
// 使用 sqlite 演示，运行结束会自动清理 db 文件。
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/tomygin/borm"
	"github.com/tomygin/borm/session"
)

// User 既做表模型也用来演示钩子函数
type User struct {
	Name string `borm:"PRIMARY KEY"`
	Age  int
}

// ---- 钩子：Before*/After* 全部可选实现 ----

// BeforeInsert 钩子：返回非 nil error 会终止后续 SQL
func (u *User) BeforeInsert(s *session.Session) error {
	if u.Name == "bad" {
		return errors.New("forbidden name")
	}
	return nil
}

// AfterInsert 钩子
func (u *User) AfterInsert(s *session.Session) error { return nil }

// BeforeQuery 钩子
func (u *User) BeforeQuery(s *session.Session) error { return nil }

// AfterQuery 钩子
func (u *User) AfterQuery(s *session.Session) error { return nil }

// BeforeUpdate 钩子
func (u *User) BeforeUpdate(s *session.Session) error { return nil }

// AfterUpdate 钩子
func (u *User) AfterUpdate(s *session.Session) error { return nil }

// BeforeDelete 钩子
func (u *User) BeforeDelete(s *session.Session) error { return nil }

// AfterDelete 钩子
func (u *User) AfterDelete(s *session.Session) error { return nil }

func main() {
	const dbFile = "example.db"
	_ = os.Remove(dbFile)
	defer os.Remove(dbFile)

	title := func(s string) { fmt.Println("\n==== " + s + " ====") }

	// ======================================================
	// 1) 引擎相关：borm.NewEngine / Engine.Close / Engine.NewSession
	// ======================================================
	title("NewEngine / NewSession")

	engine, err := borm.NewEngine(dbFile) // 也支持 "sqlite://xxx.db"、mysql/pg 的 DSN
	if err != nil {
		panic(err)
	}
	defer engine.Close() // Engine.Close

	s := engine.NewSession()
	fmt.Println("engine 初始化完成")

	// ======================================================
	// 2) 表结构：Model / RefTable / DropTable / Sync
	// ======================================================
	title("Model / RefTable / DropTable")

	s.Model(&User{}) // 绑定模型

	fmt.Println("RefTable Name:", s.RefTable().Name) // RefTable

	_ = s.DropTable() // DropTable 删除表（若存在）

	title("Sync (自动建表/加列/改类型/删除多余列)")

	// 先手动创建一个"旧"表：只有 Name 列、还多了一个 OldCol
	_ = s.Raw("CREATE TABLE User (Name text, OldCol text)").Query()

	// 用新 Model Sync 一下，预期列变为 [Name, Age]（OldCol 被删，Age 被加）
	if err := s.Model(&User{}).Sync(); err != nil {
		panic(err)
	}
	var cols []string
	_ = s.Raw("SELECT name FROM pragma_table_info('User')").Query(&cols)
	fmt.Println("sync 后列:", cols)

	// ======================================================
	// 3) ORM 写操作：Insert / Update / Delete
	// ======================================================
	title("Insert / Update / Delete")

	affect, _ := s.Insert(
		&User{Name: "tomygin", Age: 20},
		&User{Name: "ice", Age: 19},
		&User{Name: "test", Age: 18},
	)
	fmt.Println("Insert affect:", affect)

	// Update 两种形式：
	// 1) "k1", v1, "k2", v2 …
	n, _ := s.Where("Name = ?", "tomygin").Update("Age", 21)
	fmt.Println("Update (kv) affect:", n)

	// 2) map
	n2, _ := s.Where("Name = ?", "ice").Update(map[string]interface{}{
		"Age": 30,
	})
	fmt.Println("Update (map) affect:", n2)

	n3, _ := s.Where("Age = ?", 18).Delete()
	fmt.Println("Delete affect:", n3)

	// ======================================================
	// 4) ORM 查询：Where / OrderBy / Limit / Offset / Query / Count
	//    Query 根据 dest 的反射类型自动判别取一条还是取多条
	// ======================================================
	title("Query / Where / OrderBy / Limit / Offset / Count")

	var u User
	_ = s.Where("Name = ?", "tomygin").Query(&u) // dest 为 *Struct → 取一条
	fmt.Println("Query one:", u)

	var all []User
	_ = s.OrderBy("Age DESC").Limit(10).Offset(0).Query(&all) // dest 为 *[]Struct → 取多条
	fmt.Println("Query all:", all)

	total, _ := s.Where("Age > ?", 1).Count() // Count
	fmt.Println("Count:", total)

	// ======================================================
	// 5) Raw 系列：Raw / Query
	//    Query 不带参数 = 执行写操作；带参数按反射类型取一条/取多条
	// ======================================================
	title("Raw + Query")

	// Query 不带参数：写操作
	_ = s.Raw("INSERT INTO User (Name, Age) VALUES (?, ?)", "raw_a", 100).Query()

	// Raw + Query 到结构体（*Struct → 取一条）
	var u2 User
	_ = s.Raw("SELECT Name, Age FROM User WHERE Name = ?", "tomygin").Query(&u2)
	fmt.Println("Raw Query struct:", u2)

	// Raw + Query 到基础类型（*int → 取一条）
	var cnt int
	_ = s.Raw("SELECT COUNT(*) FROM User").Query(&cnt)
	fmt.Println("Raw Query int:", cnt)

	// Raw + Query 到结构体切片（*[]Struct → 取多条）
	var users []User
	_ = s.Raw("SELECT Name, Age FROM User ORDER BY Age DESC").Query(&users)
	fmt.Println("Raw Query []struct:", users)

	// Raw + Query 到基础类型切片（*[]string → 取多条）
	var names []string
	_ = s.Raw("SELECT Name FROM User").Query(&names)
	fmt.Println("Raw Query []string:", names)

	// 多段 Raw 链式拼接（自动以空格连接）
	var filtered []User
	_ = s.Raw("SELECT Name, Age FROM User").
		Raw("WHERE Age BETWEEN ? AND ?", 10, 30).
		Raw("ORDER BY Age DESC").
		Query(&filtered)
	fmt.Println("Raw 链式 Query:", filtered)

	// ======================================================
	// 6) Session 杂项：Clear
	// ======================================================
	title("Clear")

	// Clear 手动清空 SQL 缓冲区（一般无需手动调用，终结方法执行后会自动清理）
	s.Raw("SELECT 1")
	s.Clear()
	fmt.Println("Clear 完成，缓冲区已重置")

	// ======================================================
	// 7) 事务：Engine.Transaction（高层）或 Begin/Commit/RollBack（底层）
	// ======================================================
	title("Engine.Transaction")

	r, err := engine.Transaction(func(ts *session.Session) (interface{}, error) {
		ts.Model(&User{})
		_ = ts.Sync()
		_, _ = ts.Insert(&User{Name: "tx_ok", Age: 66})
		var t User
		return t, ts.Where("Name = ?", "tx_ok").Query(&t)
	})
	fmt.Println("Transaction ok:", r, err)

	// 失败回滚示例
	r2, err2 := engine.Transaction(func(ts *session.Session) (interface{}, error) {
		ts.Model(&User{})
		_, _ = ts.Insert(&User{Name: "tx_bad", Age: 1})
		return nil, errors.New("manual rollback")
	})
	fmt.Println("Transaction rollback:", r2, err2)

	// 验证 tx_bad 已回滚：Query 取一条时查不到会返回 sql.ErrNoRows
	var miss User
	err3 := s.Where("Name = ?", "tx_bad").Query(&miss)
	fmt.Println("tx_bad 已回滚（查不到）:", errors.Is(err3, sql.ErrNoRows))

	title("Begin / Commit / RollBack（手动事务）")

	ts := engine.NewSession().Model(&User{})
	if err := ts.Begin(); err == nil {
		if _, err := ts.Insert(&User{Name: "manual_tx", Age: 77}); err != nil {
			_ = ts.RollBack()
		} else {
			_ = ts.Commit()
		}
	}

	// ======================================================
	// 8) 底层接口：Session.DB / CommonDB
	// ======================================================
	title("DB / CommonDB")

	// DB() 返回底层执行器：无事务时是 *sql.DB，事务中是 *sql.Tx，
	// 二者都满足 session.CommonDB 接口。需要时可绕过 borm 直接跑原生 database/sql 调用。
	var db session.CommonDB = s.DB()
	var raw int
	_ = db.QueryRow("SELECT COUNT(*) FROM User").Scan(&raw)
	fmt.Println("via CommonDB COUNT:", raw)

	title("展示完毕")
}
