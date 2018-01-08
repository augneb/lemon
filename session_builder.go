package lemon

import (
	"fmt"
	"strings"
	"reflect"
)

func (s *Session) buildSelectString() string {
	str := "SELECT "

	if opt := s.buildOptionString(); opt != "" {
		str += opt + " "
	}

	if len(s.columns) > 0 {
		columns := []string{}
		for _, v := range s.columns {
			if strings.Contains(v, "(") || strings.Contains(v, " ") {
				columns = append(columns, v)
			} else {
				columns = append(columns, "`"+v+"`")
			}
		}

		return str + strings.Join(columns, ",")
	}

	return str + "*"
}

func (s *Session) buildFromString(tableOnly bool) string {
	str := ""
	if !tableOnly {
		str += "FROM "
	}

	return str + "`" + s.table.Name + "`"
}

func (s *Session) buildWhereString() string {
	str := s.buildCriteriaString(&s.where)

	if str != "" {
		return "WHERE " + str
	}

	return ""
}

func (s *Session) buildOrderByString() string {
	if len(s.orderBy) > 0 {
		return "ORDER BY " + strings.Join(s.orderBy, ", ")
	}

	return ""
}

func (s *Session) buildGroupByString() string {
	if len(s.groupBy) > 0 {
		return "GROUP BY " + strings.Join(s.groupBy, ", ")
	}

	return ""
}

func (s *Session) buildLimitString() string {
	str := ""

	if s.offset > 0 {
		str = fmt.Sprintf("LIMIT %d OFFSET %d", s.limit, s.offset)
	} else if s.limit > 0 {
		str = fmt.Sprintf("LIMIT %d", s.limit)
	}

	return str
}

func (s *Session) buildHavingString() string {
	str := s.buildCriteriaString(&s.having)

	if str != "" {
		return "HAVING " + str
	}

	return ""
}

func (s *Session) buildOptionString() string {
	if len(s.options) > 0 {
		return strings.Join(s.options, ", ")
	}

	return ""
}

func (s *Session) buildSetString() string {
	arr := []string{}
	for k, v := range s.set {
		if reflect.TypeOf(v).Name() == "rawStore" {
			arr = append(arr, "`"+k+"`"+" = "+v.(rawStore).value)

			continue
		}

		arr = append(arr, "`"+k+"`"+" = ?")
	}

	return "SET " + strings.Join(arr, ", ")
}

func (s *Session) buildValuesString() string {
	if s.fields == nil || len(s.fields) == 0 {
		return ""
	}

	str := "(`"
	str += strings.Join(s.fields, "`,`")
	str += "`) VALUES "

	l := len(s.fields)

	val := make([]string, l)
	for i := 0; i < l; i++ {
		val[i] = "?"
	}

	valStr := "(" + strings.Join(val, ",") + ")"

	l = len(s.args) / l

	val = make([]string, l)
	for i := 0; i < l; i++ {
		val[i] = valStr
	}

	return str + strings.Join(val, ",")
}

func (s *Session) buildCriteriaString(store *[]conditionStore) string {
	if len(*store) == 0 {
		return ""
	}

	statement := ""
	useConnector := false
	for _, item := range *store {
		if item.bracket != "" {
			if item.bracket == bracketOpen {
				if useConnector {
					statement += " " + item.connector + " "
				}

				useConnector = false
			} else {
				useConnector = true
			}

			statement += item.bracket
			continue
		}

		if useConnector {
			statement += " " + item.connector + " "
		}

		useConnector = true

		value := "?"
		if item.operator == "IN" || item.operator == "NOT IN" {
			vals := []string{}
			for range item.value.([]interface{}) {
				vals = append(vals, "?")
			}

			value = bracketOpen + strings.Join(vals, ",") + bracketClose
		} else if item.operator == "IS" || item.operator == "IS NOT" {
			if item.value == nil {
				value = "NULL"
			} else {
				value = item.value.(string)
			}
		}

		if strings.Contains(item.column, "(") || strings.Contains(item.column, " ") {
			statement += item.column
		} else {
			statement += "`" + item.column + "`"
		}

		statement += " " + item.operator + " " + value
	}

	return statement
}

func (s *Session) buildSelect() {
	parts := sliceFilter([]string{
		s.buildSelectString(),
		s.buildFromString(false),
		s.buildWhereString(),
		s.buildGroupByString(),
		s.buildHavingString(),
		s.buildOrderByString(),
		s.buildLimitString(),
	})

	s.sql = strings.Join(parts, " ")

	s.args = []interface{}{}
	s.getCriteriaValues(&s.where)
	s.getCriteriaValues(&s.having)
}

func (s *Session) GetSelectSql() (string, []interface{}) {
	if s.sql == "" {
		s.buildSelect()
	}

	return s.sql, s.args
}

func (s *Session) getCriteriaValues(store *[]conditionStore) {
	if len(*store) == 0 {
		return
	}

	for _, item := range *store {
		if item.bracket != "" {
			continue
		}

		if item.operator == "IN" || item.operator == "NOT IN" {
			for _, v := range item.value.([]interface{}) {
				s.args = append(s.args, v)
			}

			continue
		}

		if item.operator == "IS" || item.operator == "IS NOT" {
			continue
		}

		s.args = append(s.args, item.value)
	}
}

func (s *Session) buildUpdate() {
	parts := sliceFilter([]string{
		"UPDATE",
		s.buildFromString(true),
		s.buildSetString(),
		s.buildWhereString(),
		s.buildOrderByString(),
		s.buildLimitString(),
	})

	s.sql = strings.Join(parts, " ")

	s.args = []interface{}{}
	for _, v := range s.set {
		if reflect.TypeOf(v).Name() == "rawStore" {
			continue
		}

		s.args = append(s.args, v)
	}

	s.getCriteriaValues(&s.where)
}

func (s *Session) GetUpdateSql() (string, []interface{}) {
	if s.sql == "" {
		s.buildUpdate()
	}

	return s.sql, s.args
}

func (s *Session) buildDelete() {
	parts := sliceFilter([]string{
		"DELETE",
		s.buildFromString(false),
		s.buildWhereString(),
		s.buildOrderByString(),
		s.buildLimitString(),
	})

	s.sql = strings.Join(parts, " ")

	s.args = []interface{}{}
	s.getCriteriaValues(&s.where)
}

func (s *Session) GetDeleteSql() (string, []interface{}) {
	if s.sql == "" {
		s.buildDelete()
	}

	return s.sql, s.args
}

func (s *Session) buildInsert() {
	s.sql = strings.Join([]string{
		"INSERT INTO",
		s.buildFromString(true),
		s.buildValuesString(),
	}, " ")
}

func (s *Session) GetInsertSql() (string, []interface{}) {
	if s.sql == "" {
		s.buildInsert()
	}

	return s.sql, s.args
}
