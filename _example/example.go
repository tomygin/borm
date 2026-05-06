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
	// 2) 表结构：Model / RefTable / CreateTable / IsExistTable / DropTable / Sync
	// ======================================================
	title("Model / CreateTable / IsExistTable / DropTable")

	s.Model(&User{}) // 绑定模型

	fmt.Println("RefTable Name:", s.RefTable().Name)       // RefTable
	fmt.Println("IsExistTable(before):", s.IsExistTable()) // false

	_ = s.CreateTable() // CreateTable
	fmt.Println("IsExistTable(after create):", s.IsExistTable())

	_ = s.DropTable() // DropTable
	fmt.Println("IsExistTable(after drop):", s.IsExistTable())

	title("Sync (自动建表/加列/改类型/删除多余列)")

	// 先手动创建一个"旧"表：只有 Name 列、还多了一个 OldCol
	_ = s.Raw("CREATE TABLE User (Name text, OldCol text)").Run()

	// 用新 Model Sync 一下，预期列变为 [Name, Age]（OldCol 被删，Age 被加）
	if err := s.Model(&User{}).Sync(); err != nil {
		panic(err)
	}
	var cols []string
	_ = s.Raw("SELECT name FROM pragma_table_info('User')").All(&cols)
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
	// 4) ORM 查询：Where / OrderBy / Limit / Offset / Get / All / Count
	// ======================================================
	title("Get / All / Where / OrderBy / Limit / Offset / Count")

	var u User
	_ = s.Where("Name = ?", "tomygin").Get(&u) // Get 单条
	fmt.Println("Get:", u)

	var all []User
	_ = s.OrderBy("Age DESC").Limit(10).Offset(0).All(&all) // All 多条
	fmt.Println("All:", all)

	total, _ := s.Where("Age > ?", 1).Count() // Count
	fmt.Println("Count:", total)

	// ======================================================
	// 5) Raw 系列：Raw / Run / Get / All / Exec / QueryRow / QueryRows
	// ======================================================
	title("Raw + Run / Get / All")

	// Run：写操作
	_ = s.Raw("INSERT INTO User (Name, Age) VALUES (?, ?)", "raw_a", 100).Run()

	// Raw + Get 到结构体
	var u2 User
	_ = s.Raw("SELECT Name, Age FROM User WHERE Name = ?", "tomygin").Get(&u2)
	fmt.Println("Raw Get struct:", u2)

	// Raw + Get 到基础类型
	var cnt int
	_ = s.Raw("SELECT COUNT(*) FROM User").Get(&cnt)
	fmt.Println("Raw Get int:", cnt)

	// Raw + All 到结构体切片
	var users []User
	_ = s.Raw("SELECT Name, Age FROM User ORDER BY Age DESC").All(&users)
	fmt.Println("Raw All []struct:", users)

	// Raw + All 到基础类型切片
	var names []string
	_ = s.Raw("SELECT Name FROM User").All(&names)
	fmt.Println("Raw All []string:", names)

	// 多段 Raw 链式拼接（自动以空格连接）
	var filtered []User
	_ = s.Raw("SELECT Name, Age FROM User").
		Raw("WHERE Age BETWEEN ? AND ?", 10, 30).
		Raw("ORDER BY Age DESC").
		All(&filtered)
	fmt.Println("Raw 链式 All:", filtered)

	title("Raw + Exec / QueryRow / QueryRows（底层 API）")

	// Exec：返回 sql.Result
	res, _ := s.Raw("UPDATE User SET Age = Age + 1 WHERE Name = ?", "raw_a").Exec()
	if res != nil {
		n, _ := res.RowsAffected()
		fmt.Println("Exec RowsAffected:", n)
	}

	// QueryRow：单行 *sql.Row
	row := s.Raw("SELECT Age FROM User WHERE Name = ?", "raw_a").QueryRow()
	var age int
	if row != nil {
		_ = row.Scan(&age)
	}
	fmt.Println("QueryRow Age:", age)

	// QueryRows：多行 *sql.Rows（需要自己关）
	rows, _ := s.Raw("SELECT Name, Age FROM User").QueryRows()
	if rows != nil {
		for rows.Next() {
			var n string
			var a int
			_ = rows.Scan(&n, &a)
			fmt.Println("  row:", n, a)
		}
		_ = rows.Close()
	}

	// ======================================================
	// 6) Session 杂项：DB / Clear / CallMethod
	// ======================================================
	title("DB / Clear / CallMethod")

	// DB 返回底层 *sql.DB (或事务里的 *sql.Tx)
	_ = s.DB() // session.CommonDB

	// Clear 手动清空缓冲区（一般不用调，执行类方法会自动调）
	s.Raw("SELECT 1")
	s.Clear()

	// CallMethod 手动调用钩子（底层 API，一般不用）
	s.Model(&User{}).CallMethod(session.BeforeQuery, nil)
	// session 里可用的钩子常量：
	_ = []string{
		session.BeforeQuery, session.AfterQuery,
		session.BeforeInsert, session.AfterInsert,
		session.BeforeUpdate, session.AfterUpdate,
		session.BeforeDelete, session.AfterDelete,
	}

	// ======================================================
	// 7) 事务：Engine.Transaction（高层）或 Begin/Commit/RollBack（底层）
	// ======================================================
	title("Engine.Transaction")

	r, err := engine.Transaction(func(ts *session.Session) (interface{}, error) {
		ts.Model(&User{})
		_ = ts.Sync()
		_, _ = ts.Insert(&User{Name: "tx_ok", Age: 66})
		var t User
		return t, ts.Where("Name = ?", "tx_ok").Get(&t)
	})
	fmt.Println("Transaction ok:", r, err)

	// 失败回滚示例
	r2, err2 := engine.Transaction(func(ts *session.Session) (interface{}, error) {
		ts.Model(&User{})
		_, _ = ts.Insert(&User{Name: "tx_bad", Age: 1})
		return nil, errors.New("manual rollback")
	})
	fmt.Println("Transaction rollback:", r2, err2)

	// 验证 tx_bad 已回滚：Get 返回 sql.ErrNoRows 表示找不到
	var miss User
	err3 := s.Where("Name = ?", "tx_bad").Get(&miss)
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
	// 8) session.New（底层构造）& CommonDB 接口
	// ======================================================
	title("session.New / CommonDB")

	rawDB, _ := sql.Open("sqlite", dbFile)
	// 需要一个 dialect，可以从 engine 那边"借"出来，这里演示最直白的用法
	// 直接从另一个 NewSession 的 s.DB() 得到 CommonDB 已经够用，
	// 真实项目不建议绕过 borm.NewEngine 使用。
	_ = rawDB

	// session.CommonDB 本身就是接口，*sql.DB / *sql.Tx 都满足
	var _ session.CommonDB = engine.NewSession().DB()

	title("展示完毕")
}
