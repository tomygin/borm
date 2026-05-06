package main

import (
	"fmt"
	"os"

	"github.com/tomygin/borm"
	"github.com/tomygin/borm/session"
)

type User struct {
	Name string `borm:"PRIMARY KEY"`
	Age  int
}

func main() {
	_ = os.Remove("test.db")

	// DSN 自动识别数据库类型
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
	if err := s.Sync(); err != nil {
		panic(err)
	}

	// 插入
	_, _ = s.Insert(
		&User{Name: "tomygin", Age: 20},
		&User{Name: "ice", Age: 19},
		&User{Name: "test", Age: 18},
	)

	// ===== 统一的 Get / All =====

	// 取一条（ORM 风格）
	var u User
	_ = s.Where("Name = ?", "tomygin").Get(&u)
	fmt.Println("orm get:", u)

	// 取多条（ORM 风格）
	var users []User
	_ = s.Where("Age > ?", 10).Limit(10).Offset(0).All(&users)
	fmt.Println("orm all:", users)

	// 取多条（Raw 风格，和 ORM 同名）
	var users2 []User
	_ = s.Raw("SELECT Name, Age FROM User WHERE Age > ? ORDER BY Age DESC", 10).All(&users2)
	fmt.Println("raw all:", users2)

	// 取一条（Raw 风格 + 基础类型）
	var count int
	_ = s.Raw("SELECT COUNT(*) FROM User").Get(&count)
	fmt.Println("count:", count)

	// 取多条基础类型切片
	var names []string
	_ = s.Raw("SELECT Name FROM User").All(&names)
	fmt.Println("names:", names)

	// 写操作
	_ = s.Raw("UPDATE User SET Age = Age + 1 WHERE Name = ?", "tomygin").Run()

	// 更新 / 删除（ORM）
	_, _ = s.Where("Age = ?", 18).Limit(1).Delete()
	_, _ = s.Where("Name = ?", "tomygin").Update("Age", 18)

	// 排序
	_ = s.OrderBy("Age DESC").Get(&u)

	// 事务：失败自动回滚
	r, err := engine.Transaction(func(s *session.Session) (interface{}, error) {
		s.Model(&User{})
		_ = s.Sync()
		_, _ = s.Insert(&User{Name: "in_tx", Age: 99})
		t := User{}
		return t, s.Where("Name = ?", "in_tx").Get(&t)
	})
	fmt.Println("tx:", r, err)

	// 清理
	_ = s.DropTable()
	_ = os.Remove("test.db")
}

// 钩子函数默认开启；返回非 nil 的 error 会自动终止后续 SQL 执行
func (u *User) BeforeQuery(s *session.Session) error {
	return nil
}
