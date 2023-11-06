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

	// 多条查询
	tmps := []User{}
	if err := s.Where("Age > 10").Find(&tmps); err == nil {
		log.Info("拿到数据", tmps)
	}

}

// 钩子函数
func (u *User) BeforeQuery(s *session.Session) error {
	log.Info("钩子函数运行成功")

	// 不希望最后执行sql
	// s.Abort()

	return nil
}
