# lemon

lemon 是一个简单Go语言ORM库。

## 驱动支持

目前只支持 MySQL  

## 安装

	go get -u github.com/augeb/lemon

# 快速开始

* 第一步

```Go
db, err := lemon.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8")
```

* 定义表结构体

```Go
type User struct {
    Id int64 `db:"pk ai"` // primary key auto_increment
    Name string `db:"unique   "` //unique key
    NickName string `db:"nick_name"` // 指定字段名
}
```

* 查询

```Go
// SELECT * FROM user LIMIT 1
has, err := db.Get(&user)  

// SELECT name FROM user WHERE id = ? LIMIT 1
has, err := db.Columns("name").Where("id", id).Get(&user)  

// SELECT * FROM user WHERE id < ?
users := []User{}
has, err := db.Where("id", id, "<").Find(&users)  

// SELECT * FROM user WHERE id IN (?, ?, ?)
users := []User{}
has, err := db.Where("id", []int{1,2,3}, "IN").Find(&users)

// SELECT * FROM user WHERE id = ? OR id = ?
users := []User{}
has, err := db.Where("id", 1).OrWhere("id", 2).Find(&users)  

// SELECT * FROM user WHERE (id = ? OR id = ?) AND name LIKE ?
users := []User{}
has, err := db.WhereBracket(func (s *lemon.Session) {
	s.Where("id", 1).OrWhere("id", 2)
}).Where("name", "%test%", "like").Find(&users)
```

* 插入

```Go
affected, err := db.Values(&user).Insert()
affected, err := db.Values(&user1).Values(&user2).Insert()
```

* 更新

```Go
// UPDATE user SET name = ? Where id = ?
affected, err := db.Where("id", 1).Set(&user, "name").Update()

// UPDATE user SET name = ? Where id = ?
affected, err := db.Table(&user).Where("id", 1).Set("name", "xxx").Update()

// UPDATE user SET name = ? Where id = ?
affected, err := db.Table(&user).Where("id", 1).Set(map[string]interface{}{"name": "xxx"}).Update()

// UPDATE user SET xx = xx + 1 Where id = ?
affected, err := db.Table(&user).Where("id", 1).SetRaw("xx", "xx + 1").Update()
```

* 删除

```Go
// DELETE FROM user Where id = ?
affected, err := db.Table(&user).Where("id", 1).Delete()
```
