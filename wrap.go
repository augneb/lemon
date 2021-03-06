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

func (o *Orm) Get(obj interface{}, to ...*map[string]interface{}) (find bool, err error) {
	return o.NewSession().Get(obj, to...)
}

func (o *Orm) Find(obj interface{}) (find bool, err error) {
	return o.NewSession().Find(obj)
}

func (o *Orm) Id(id interface{}, obj interface{}, to ...*map[string]interface{}) (find bool, err error) {
	return o.NewSession().Where("id", id).Get(obj, to...)
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

func (o *Orm) Where(column string, value interface{}, operator ...string) *Session {
	return o.NewSession().Where(column, value, operator...)
}

func (o *Orm) OrderBy(column string, order ...string) *Session {
	return o.NewSession().OrderBy(column, order...)
}

func (o *Orm) GroupBy(group string) *Session {
	return o.NewSession().GroupBy(group)
}

func (o *Orm) Limit(limit int) *Session {
	return o.NewSession().Limit(limit)
}

func (o *Orm) Having(column string, value interface{}, operator ...string) *Session {
	return o.NewSession().Having(column, value, operator...)
}

func (o *Orm) WhereBracket(call func(*Session), connector ...string) *Session {
	return o.NewSession().WhereBracket(call, connector...)
}

func (o *Orm) HavingBracket(call func(*Session), connector ...string) *Session {
	return o.NewSession().HavingBracket(call, connector...)
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
