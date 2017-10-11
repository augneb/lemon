package lemon

import (
	"time"
	"fmt"
	"database/sql"
)

const logicalAnd   = "AND"
const logicalOr    = "OR"
const bracketOpen  = "("
const bracketClose = ")"

type conditionStore struct {
	column    string
	value     interface{}
	operator  string
	connector string
	bracket   string
}

type rawStore struct {
	value string
}

type Session struct {
	orm *Orm

	enableCache bool

	tx *sql.Tx

	sql string
	args []interface{}

	queryStart time.Time
	queryTime  float64

	options []string
	columns []string
	orderBy []string
	groupBy []string
	where   []conditionStore
	having  []conditionStore
	table   *structCache
	limit   int64
	offset  int64

	// for update
	set     map[string]interface{}
	// for insert
	fields  []string
	// for select
	colIdx  []int
}

// 设置是否启用缓存
func (s *Session) SetEnableCache(cache bool) *Session {
	s.enableCache = cache

	return s
}

// 查询后执行相关操作，各种事件回掉，打印 SQL 等
func (s *Session) after(status bool) {
	s.queryTime = time.Since(s.queryStart).Seconds()

	// TODO: open a goroutine?
	if s.orm.longQueryTime > 0 && s.queryTime >= s.orm.longQueryTime {
		s.orm.eventLongQuery(s)
	}

	if !s.orm.debug {
		return
	}

	str := "\033[42"
	if !status {
		str = "\033[41"
	}

	fmt.Printf("[%s] " + str + ";37;1mOrm\033[0m %s %v [%fs]\n", date("m-d H:i:s", s.queryStart), s.sql, s.args, s.queryTime)
}

// 重设清理
func (s *Session) reset() {
	s.limit   = 0
	s.offset  = 0
	s.table   = nil
	s.options = nil
	s.columns = nil
	s.orderBy = nil
	s.groupBy = nil
	s.where   = nil
	s.having  = nil
	s.set     = nil
	s.fields  = nil
	s.colIdx  = nil
}
