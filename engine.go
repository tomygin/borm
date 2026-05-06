// Copyright 2023 tomygin
//
// Licensed under the MIT License

// Package borm implements a ORM framework
package borm

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/tomygin/borm/dialect"
	"github.com/tomygin/borm/session"
)

// Engine 是引擎对象
// db 用于调用 go 的 database/sql 连接后的对象
// dialect 用于对不同的数据库的类型适配为 go 的数据类型
type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

// detectDriver 根据 DSN 自动识别数据库驱动类型
// 支持 mysql / postgres / sqlite
func detectDriver(dsn string) (driver string, source string, err error) {
	lower := strings.ToLower(strings.TrimSpace(dsn))

	switch {
	// postgres://user:pass@host:port/dbname?sslmode=disable
	case strings.HasPrefix(lower, "postgres://"),
		strings.HasPrefix(lower, "postgresql://"):
		return "pgx", dsn, nil

	// mysql://user:pass@tcp(host:port)/dbname
	case strings.HasPrefix(lower, "mysql://"):
		return "mysql", strings.TrimPrefix(dsn, "mysql://"), nil

	// sqlite://path/to/file.db  或  file:xxx.db
	case strings.HasPrefix(lower, "sqlite://"):
		return "sqlite", strings.TrimPrefix(dsn, "sqlite://"), nil
	case strings.HasPrefix(lower, "sqlite3://"):
		return "sqlite", strings.TrimPrefix(dsn, "sqlite3://"), nil
	case strings.HasPrefix(lower, "file:"):
		return "sqlite", dsn, nil
	}

	// 兼容 mysql 常见 DSN: user:pass@tcp(host:port)/dbname
	if strings.Contains(dsn, "@tcp(") || strings.Contains(dsn, "@unix(") {
		return "mysql", dsn, nil
	}

	// postgres 的 kv 形式: host=xxx user=xxx dbname=xxx
	if strings.Contains(lower, "host=") && strings.Contains(lower, "dbname=") {
		return "pgx", dsn, nil
	}

	// 以 .db / .sqlite 结尾或者不含协议的本地文件默认作 sqlite
	if strings.HasSuffix(lower, ".db") ||
		strings.HasSuffix(lower, ".sqlite") ||
		strings.HasSuffix(lower, ".sqlite3") {
		return "sqlite", dsn, nil
	}

	return "", "", fmt.Errorf("cannot detect database driver from dsn: %s", dsn)
}

// NewEngine 用于生成一个 Engine 实例
// 仅接受一个 DSN 字符串，自动识别数据库类型
//
// 支持的 DSN 形式：
//   - postgres://user:pass@host:port/dbname?sslmode=disable
//   - mysql://user:pass@tcp(host:3306)/dbname
//   - user:pass@tcp(host:3306)/dbname                (mysql)
//   - sqlite://path/to/file.db
//   - path/to/file.db                                (sqlite)
func NewEngine(dsn string) (e *Engine, err error) {
	if dsn == "" {
		return nil, errors.New("dsn is empty")
	}

	driver, source, err := detectDriver(dsn)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(driver, source)
	if err != nil {
		return nil, err
	}

	// 测试连接
	if err = db.Ping(); err != nil {
		return nil, err
	}

	// 获取 sql 方言
	dial, ok := dialect.GetDialect(driver)
	if !ok {
		return nil, fmt.Errorf("dialect %s not found", driver)
	}

	e = &Engine{db: db, dialect: dial}
	return
}

// Close 关闭底层数据库连接
func (e *Engine) Close() error {
	return e.db.Close()
}

// NewSession 创建一个新的会话
func (e *Engine) NewSession() *session.Session {
	return session.New(e.db, e.dialect)
}
