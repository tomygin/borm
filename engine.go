// 我相信box未来一定会有很多人用的
package box

import (
	"database/sql"
	"errors"

	"github.com/tomygin/box/cache"
	"github.com/tomygin/box/dialect"
	"github.com/tomygin/box/log"
	"github.com/tomygin/box/session"
)

// Eingie是引擎对象
// db用于调用go的database/sql连接后的对象
// dialect用于对不同的数据库的类型适配为go的数据类型
type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

// NewEngine用于生成一个Engine实例
// 在没有指定驱动的时候默认认为sqlite3
func NewEngine(info ...string) (e *Engine, err error) {
	driver := "sqlite"
	source := "box.db"

	switch len(info) {
	case 0:
	case 1:
		source = info[0]
	case 2:
		driver, source = info[0], info[1]
	default:
		return nil, errors.New("invalid param")
	}

	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}

	//测试连接
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}

	//获取sql方言
	dial, ok := dialect.GetDialect(driver)
	if !ok {
		log.Error("dialect %s Not Found ", driver)
		return
	}

	e = &Engine{db: db, dialect: dial}

	log.Infof("Connect %s success \n", source)
	return
}

func (e *Engine) Close() {
	if err := e.db.Close(); err != nil {
		log.Error("Failed to close database ")
	}
	log.Info("Close database success ")
}

func (e *Engine) NewSession() *session.Session {
	s := session.New(e.db, e.dialect)
	return s
}

func (e *Engine) NewCache(maxBytes int64) *cache.Cache {
	c := cache.New(maxBytes, nil, e.NewSession())
	return c
}
