package session

import (
	"github.com/tomygin/borm/clause"
)

// Insert 插入一条或多条记录
func (s *Session) Insert(values ...interface{}) (int64, error) {
	s.callMethod(beforeInsert, nil)
	defer s.callMethod(afterInsert, nil)

	recordValues := make([]interface{}, 0)
	for _, value := range values {
		table := s.Model(value).RefTable()
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames)
		recordValues = append(recordValues, table.RecordValues(value))
	}

	s.clause.Set(clause.VALUES, recordValues...)
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	resout, err := s.Raw(sql, vars...).exec()
	if err != nil {
		return 0, err
	}
	return resout.RowsAffected()
}

// Update 更新匹配行
func (s *Session) Update(kv ...interface{}) (int64, error) {
	s.callMethod(beforeUpdate, nil)
	defer s.callMethod(afterUpdate, nil)

	m, ok := kv[0].(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
	}
	s.clause.Set(clause.UPDATE, s.RefTable().Name, m)
	sql, vars := s.clause.Build(clause.UPDATE, clause.WHERE)
	result, err := s.Raw(sql, vars...).exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Delete 删除匹配行
func (s *Session) Delete() (int64, error) {
	s.callMethod(beforeDelete, nil)
	defer s.callMethod(afterDelete, nil)

	s.clause.Set(clause.DELETE, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sql, vars...).exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Count 返回满足条件的行数
func (s *Session) Count() (int64, error) {
	s.clause.Set(clause.COUNT, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(sql, vars...).queryRow()

	var tmp int64
	err := row.Scan(&tmp)
	return tmp, err
}

// Limit 限制返回条数
func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	return s
}

// Offset 跳过多少条数据
func (s *Session) Offset(num int) *Session {
	s.clause.Set(clause.OFFSET, num)
	return s
}

// Where 设置 WHERE 子句
func (s *Session) Where(desc string, args ...interface{}) *Session {
	var vars []interface{}
	s.clause.Set(clause.WHERE, append(append(vars, desc), args...)...)
	return s
}

// OrderBy 设置 ORDER BY 子句
func (s *Session) OrderBy(desc string) *Session {
	s.clause.Set(clause.ORDERBY, desc)
	return s
}

// buildSelect 根据当前 Session 的 Model 与 clause 构造 SELECT 语句
// 返回 (sql, vars, table)，table 为对应 Schema，便于按字段顺序 Scan
func (s *Session) buildSelect(model interface{}) (string, []interface{}) {
	if model != nil {
		s.Model(model)
	}
	table := s.RefTable()
	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sqlStr, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT, clause.OFFSET)
	return sqlStr, vars
}

// hasRaw 判断当前 Session 的 SQL 缓冲区是否已经有内容（代表处于 Raw 模式）
func (s *Session) hasRaw() bool {
	return s.sql.Len() > 0
}
