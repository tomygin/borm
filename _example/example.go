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
