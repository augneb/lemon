package lemon

import (
	"regexp"
	"strings"
	"strconv"
)

// code is copy from xorm

type Uri struct {
	Type    string
	Host    string
	Port    int
	Schema  string
	Charset string
}

func (o *Orm) Parse(driverName, dataSourceName string) *Uri {
	dsnPattern := regexp.MustCompile(
		`^(?:(?P<user>.*?)(?::(?P<passwd>.*))?@)?` + // [user[:password]@]
			`(?:(?P<net>[^\(]*)(?:\((?P<addr>[^\)]*)\))?)?` + // [net[(addr)]]
			`\/(?P<schema>.*?)` + // /schema
			`(?:\?(?P<params>[^\?]*))?$`) // [?param1=value1&paramN=valueN]
	matches := dsnPattern.FindStringSubmatch(dataSourceName)
	names := dsnPattern.SubexpNames()

	uri := &Uri{Type: driverName}

	for i, match := range matches {
		switch names[i] {
		case "dbname":
			uri.Schema = match
		case "addr":
			uri.Port = 3306
			uri.Host = match
			if strings.Contains(match, ":") {
				ms := strings.Split(match, ":")
				uri.Host = ms[0]
				uri.Port, _ = strconv.Atoi(ms[1])
			}
		case "params":
			if len(match) > 0 {
				kvs := strings.Split(match, "&")
				for _, kv := range kvs {
					splits := strings.Split(kv, "=")
					if len(splits) == 2 {
						switch splits[0] {
						case "charset":
							uri.Charset = splits[1]
						}
					}
				}
			}

		}
	}

	return uri
}
