package lemon

func (o *Orm) Master() *Session {
	return o.NewSession().Master()
}

func (o *Orm) Slave() *Session {
	return o.NewSession().Slave()
}

func (o *Orm) EnableCache(cache bool) *Session {
	return o.NewSession().EnableCache(cache)
}

func (o *Orm) Columns(columns ...string) *Session {
	return o.NewSession().Columns(columns...)
}

func (o *Orm) Option(option string) *Session {
	return o.NewSession().Option(option)
}

func (o *Orm) Table(table interface{}) *Session {
	return o.NewSession().Table(table)
}

func (o *Orm) Where(column string, operator string, value interface{}) *Session {
	return o.NewSession().Where(column, operator, value)
}

func (o *Orm) OrderBy(column string, order string) *Session {
	return o.NewSession().OrderBy(column, order)
}

func (o *Orm) GroupBy(group string) *Session {
	return o.NewSession().GroupBy(group)
}

func (o *Orm) Limit(limit int64) *Session {
	return o.NewSession().Limit(limit)
}

func (o *Orm) Having(column string, operator string, value interface{}) *Session {
	return o.NewSession().Having(column, operator, value)
}

func (o *Orm) WhereBracket(call func(*Session), connector string) *Session {
	return o.NewSession().WhereBracket(call, connector)
}

func (o *Orm) HavingBracket(call func(*Session), connector string) *Session {
	return o.NewSession().HavingBracket(call, connector)
}

func (o *Orm) Set(column interface{}, value ...interface{}) *Session {
	return o.NewSession().Set(column, value...)
}

func (o *Orm) SetRaw(column string, value string) *Session {
	return o.NewSession().SetRaw(column, value)
}

func (o *Orm) Values(value interface{}) *Session {
	return o.NewSession().Values(value)
}

func (o *Orm) Begin() (*Session, error) {
	s := o.NewSession()

	return s, s.Begin()
}
