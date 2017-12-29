package lemon

import (
	"strings"
	"reflect"
	"regexp"
)

func (s *Session) Columns(columns ...string) *Session {
	if s.columns == nil {
		s.columns = []string{}
	}

	s.columns = append(s.columns, columns...)

	return s
}

func (s *Session) Option(option string) *Session {
	if s.options == nil {
		s.options = []string{}
	}

	s.options = append(s.options, option)

	return s
}

func (s *Session) Table(table interface{}) *Session {
	switch table.(type) {
	case string:
		s.table = s.orm.GetTableInfo(table.(string))
	default:
		s.table = s.orm.GetTableInfo(table)
	}

	return s
}

func (s *Session) Where(column string, value interface{}, operators ...string) *Session {
	operator := "="
	if len(operators) > 0 {
		operator = operators[0]
	}

	s.criteria(&s.where, column, operator, value, logicalAnd)

	return s
}

func (s *Session) OrWhere(column string, value interface{}, operators ...string) *Session {
	operator := "="
	if len(operators) > 0 {
		operator = operators[0]
	}

	s.criteria(&s.where, column, operator, value, logicalOr)

	return s
}

func (s *Session) OrderBy(column string, orders ...string) *Session {
	if s.orderBy == nil {
		s.orderBy = []string{}
	}

	order := "ASC"
	if len(orders) > 0 {
		order = orders[0]
	}

	s.orderBy = append(s.orderBy, column+" "+order)

	return s
}

func (s *Session) GroupBy(group string) *Session {
	if s.groupBy == nil {
		s.groupBy = []string{}
	}

	s.groupBy = append(s.groupBy, group)

	return s
}

func (s *Session) Limit(limit int) *Session {
	s.limit = limit

	return s
}

func (s *Session) Offset(offset int) *Session {
	s.offset = offset

	return s
}

func (s *Session) Having(column string, value interface{}, operators ...string) *Session {
	operator := "="
	if len(operators) > 0 {
		operator = operators[0]
	}

	s.criteria(&s.having, column, operator, value, logicalAnd)

	return s
}

func (s *Session) OrHaving(column string, value interface{}, operators ...string) *Session {
	operator := "="
	if len(operators) > 0 {
		operator = operators[0]
	}

	s.criteria(&s.having, column, operator, value, logicalOr)

	return s
}

func (s *Session) WhereBracket(call func(*Session), connector ...string) *Session {
	s.bracket(&s.where, call, connector)

	return s
}

func (s *Session) HavingBracket(call func(*Session), connector ...string) *Session {
	s.bracket(&s.having, call, connector)

	return s
}

func (s *Session) Set(column interface{}, value ...interface{}) *Session {
	if s.set == nil {
		s.set = map[string]interface{}{}
	}

	switch column.(type) {
	case string:
		var val interface{}
		if len(value) > 0 {
			val = value[0]
		}

		s.set[column.(string)] = val

	case map[string]interface{}:
		s.set = column.(map[string]interface{})

	case *map[string]interface{}:
		s.set = *(column.(*map[string]interface{}))

	default:
		// reset
		s.set = map[string]interface{}{}

		if len(value) == 0 {
			break
		}

		fields := map[string]bool{}
		for _, v := range value {
			fields[v.(string)] = true
		}

		t := s.orm.GetTableInfo(column)
		i := reflect.Indirect(reflect.ValueOf(column))
		for idx, field := range t.Fields {
			if _, ok := fields[field]; ok {
				s.set[field] = i.Field(idx).Interface()
			}
		}

		if s.table == nil {
			s.Table(t.Name)
		}
	}

	return s
}

func (s *Session) SetRaw(column string, value string) *Session {
	if s.set == nil {
		s.set = map[string]interface{}{}
	}

	s.set[column] = rawStore{value: value}

	return s
}

func (s *Session) Values(value interface{}) *Session {
	var noFields bool
	if s.fields == nil {
		noFields = true
		s.fields = []string{}
	}

	if s.args == nil {
		s.args = []interface{}{}
	}

	var tmp map[string]interface{}
	switch value.(type) {
	case map[string]interface{}:
		tmp = value.(map[string]interface{})
	case *map[string]interface{}:
		tmp = *(value.(*map[string]interface{}))
	}

	if tmp != nil {
		if noFields {
			for k := range tmp {
				s.fields = append(s.fields, k)
			}
		}

		for _, k := range s.fields {
			if v, ok := tmp[k]; ok {
				s.args = append(s.args, v)
			} else {
				s.args = append(s.args, nil)
			}
		}
	} else {
		t := s.orm.GetTableInfo(value)
		a := t.AutoIncrement
		i := reflect.Indirect(reflect.ValueOf(value))

		if s.table == nil {
			s.Table(t.Name)
		}

		for idx, field := range t.Fields {
			if _, ok := a[idx]; ok {
				continue
			}

			if noFields {
				s.fields = append(s.fields, field)
			}

			s.args = append(s.args, i.Field(idx).Interface())
		}
	}

	return s
}

func (s *Session) criteria(store *[]conditionStore, column string, operator string, value interface{}, connector string) {
	if *store == nil {
		*store = []conditionStore{}
	}

	if matched, _ := regexp.MatchString("[!=<>]", operator); !matched {
		operator = strings.ToUpper(operator)

		if operator == "IN" || operator == "NOT IN" {
			var v reflect.Value
			if v = reflect.ValueOf(value); v.Kind() == reflect.Ptr {
				v = v.Elem()
			}

			l := v.Len()

			ret := make([]interface{}, l)
			for i := 0; i < l; i++ {
				ret[i] = v.Index(i).Interface()
			}

			value = ret
		}
	}

	column = strings.Trim(column, " ")

	*store = append(*store, conditionStore{
		column:    column,
		value:     value,
		operator:  operator,
		connector: connector,
	})
}

func (s *Session) bracket(store *[]conditionStore, call func(*Session), connectors []string) {
	if *store == nil {
		*store = []conditionStore{}
	}

	connector := logicalAnd
	if len(connectors) > 0 {
		connector = connectors[0]
	}

	*store = append(*store, conditionStore{
		bracket:   bracketOpen,
		connector: connector,
	})

	call(s)

	*store = append(*store, conditionStore{
		bracket:   bracketClose,
		connector: connector,
	})
}
