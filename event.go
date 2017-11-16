package lemon

import (
	"time"
)

type EventParams struct {
	Sql string
	Args []interface{}
	StartTime time.Time
	QueryTime float64

	Uri *Uri
}

type EventCall func(params *EventParams)

// 慢查询事件回掉
func (o *Orm) eventLongQuery(s *Session) {
	// 执行回掉
	o.longQueryEventCall(&EventParams{
		Sql:       s.sql,
		Args:      s.args,
		StartTime: s.queryStart,
		QueryTime: s.queryTime,
		Uri:       s.orm.uri,
	})
}
