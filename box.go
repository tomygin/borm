// Copyright 2023 TomyGin
//
// Licensed under the MIT License

// Package borm implements a ORM framework
package borm

import "github.com/tomygin/borm/session"

// 事务的回调函数
type TxFunc func(*session.Session) (interface{}, error)

// Transaction一键事务提交，如果失败自动回滚
func (e *Engine) Transaction(f TxFunc) (result interface{}, err error) {
	s := e.NewSession()
	if err := s.Begin(); err != nil {
		return nil, err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = s.RollBack()
		} else if err != nil {
			_ = s.RollBack()
		} else {
			err = s.Commit()
		}
	}()
	return f(s)
}
