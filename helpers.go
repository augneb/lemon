package lemon

import (
	"strings"
	"regexp"
	"time"
	"fmt"
)

// 去除空行
func reduceEmptyElements(items []string) []string {
	result := []string{}
	for _, text := range items {
		if strings.Trim(text, " ") != "" {
			result = append(result, text)
		}
	}

	return result
}

// 驼峰转换为下划线
func toSnakeCase(str string, toLower bool) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchAllCap.ReplaceAllString(matchFirstCap.ReplaceAllString(str, "${1}_${2}"), "${1}_${2}")

	if toLower {
		return strings.ToLower(snake)
	}

	return snake
}

// 日期
func date(format string, ts ...time.Time) string {
	patterns := []string{
		"Y", "2006", // 4 位数字完整表示的年份
		"m", "01",   // 数字表示的月份，有前导零
		"d", "02",   // 月份中的第几天，有前导零的 2 位数字
		"H", "15",   // 小时，24 小时格式，有前导零
		"i", "04",   // 有前导零的分钟数
		"s", "05",   // 秒数，有前导零
	}

	format = strings.NewReplacer(patterns...).Replace(format)

	t := time.Now()
	if len(ts) > 0 {
		t = ts[0]
	}

	return t.Format(format)
}

// 带颜色的输出
func echo(v ...interface{}) {
	l := v[len(v)-1]

	pref := ""
	switch val := l.(type) {
	case string:
		switch val {
		case "blue":
			pref = "\033[44;37;1m"
		case "green":
			pref = "\033[42;37;1m"
		case "red":
			pref = "\033[41;37;1m"
		case "yellow":
			pref = "\033[43;37;1m"
		}
	}

	str := []string{date("[m-d H:i:s]")}

	n := len(v)
	if pref != "" {
		n--
		v = v[:n]

		str = append(str, pref)
	}

	for i := 0; i<n; i++ {
		str = append(str, "%v")
	}

	if pref != "" {
		str = append(str, "\033[0m")
	}

	str = append(str, "\n")

	fmt.Printf(strings.Join(str, " "), v...)
}

