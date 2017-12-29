package lemon

import (
	"sync"
	"time"
	"runtime"
	"math/rand"
	"database/sql"
)

const emptyCacheString = "nil"

type Orm struct {
	db         *sql.DB
	dbSlave    []*sql.DB
	dbSlaveLen int

	driverName string

	// 开启 debug 会打印查询日志
	debug bool

	// log handler
	log Logger

	// 缓存
	enableCache  bool
	cacheTime    int
	cacheEmpty   bool
	cacheHandler CacheHandler

	// uri 包含连接相关信息
	uri *Uri

	// 慢查询时间阙值
	longQueryTime float64

	// 慢查询事件回掉
	longQueryEventCall EventCall

	// 表结构缓存
	structCache sync.Map

	cachePrefix     string
	primaryCacheKey string
	uniqueCacheKey  string
}

type CacheHandler interface {
	Get(key string) []byte
	Set(key string, val []byte, expire int) error
	Del(key string)
}

//  实例化一个 ORM
func Open(driverName, dataSource string) (*Orm, error) {
	db, err := sql.Open(driverName, dataSource)
	if err != nil {
		return nil, err
	}

	o := &Orm{
		db:         db,
		driverName: driverName,
		log:        newLogger(),
	}

	o.uri = o.Parse(driverName, dataSource)

	// 结束的时候执行，关闭数据库连接
	runtime.SetFinalizer(o, func(o *Orm) {
		o.Close()
	})

	// 随机种子
	rand.Seed(time.Now().UnixNano())

	return o, nil
}

// 添加 Slave
func (o *Orm) AddSlave(dataSource string) error {
	db, err := sql.Open(o.driverName, dataSource)
	if err != nil {
		return err
	}

	o.dbSlave = append(o.dbSlave, db)
	o.dbSlaveLen++

	return nil
}

// 设置数据库最大空闲连接数
func (o *Orm) SetMaxIdleConns(n int) *Orm {
	o.db.SetMaxIdleConns(n)

	return o
}

// 设置数据库最大连接数
func (o *Orm) SetMaxOpenConns(n int) *Orm {
	o.db.SetMaxOpenConns(n)

	return o
}

// 设置连接最大存活时间
func (o *Orm) SetConnMaxLifetime(d time.Duration) *Orm {
	o.db.SetConnMaxLifetime(d)

	return o
}

// 设为 Debug Model
func (o *Orm) SetDebug(debug bool) *Orm {
	o.debug = debug

	return o
}

// 设置慢查询阙值
func (o *Orm) SetLongQueryTime(time float64) *Orm {
	o.longQueryTime = time

	return o
}

// 设置慢查询回掉函数
func (o *Orm) SetEventLongQuery(fn EventCall) *Orm {
	o.longQueryEventCall = fn

	return o
}

// 设置是否启用缓存
func (o *Orm) SetEnableCache(cache bool, prefix ...string) *Orm {
	o.enableCache = cache

	if cache && o.cachePrefix == "" {
		if len(prefix) > 0 {
			o.cachePrefix = prefix[0]
		} else {
			o.cachePrefix = "lemon"
		}

		o.primaryCacheKey = o.cachePrefix + ":primary:%s:%v"
		o.uniqueCacheKey = o.cachePrefix + ":unique:%s:%s"
	}

	return o
}

// 设置缓存时间
func (o *Orm) SetCacheTime(second int) *Orm {
	o.cacheTime = second

	return o
}

func (o *Orm) SetLogHandler(log Logger) *Orm {
	o.log = log

	return o
}

// 设置缓存器
func (o *Orm) SetCacheHandler(ch CacheHandler) *Orm {
	o.cacheHandler = ch

	return o
}

// 设置缓存空值
func (o *Orm) SetCacheEmpty(cache bool) *Orm {
	o.cacheEmpty = cache

	return o
}

// 手动关闭数据库连接
func (o *Orm) Close() error {
	return o.db.Close()
}

// 开启一个新的查询
func (o *Orm) NewSession() *Session {
	return &Session{orm: o, enableCache: o.enableCache}
}

// 随机获取一个从库连接
func (o *Orm) getSlave() *sql.DB {
	return o.dbSlave[rand.Intn(o.dbSlaveLen)]
}
