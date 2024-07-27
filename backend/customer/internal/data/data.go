package data

import (
	"customer/internal/biz"
	"customer/internal/conf"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewCustomerData)

// Data .
type Data struct {
	// 可以将所有数据操作相关组件，都封装在这个Data里面
	Rdb *redis.Client
	Mdb *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	data := new(Data)
	// 1.初始化Rdb
	redisUrl := fmt.Sprintf("redis://%s/1?dial_timeout=%d", c.Redis.Addr, 1)
	// 先建立一个配置option
	option, err := redis.ParseURL(redisUrl)
	// 没有用户密码就不写
	// redis://user:password@localhost:6789/3?dial_timeout=3&db=1&read_timeout=6s&max_retries=2
	if err != nil {
		data.Rdb = nil
		log.Fatal(err)
	}
	// newclient不会立即连接，而是执行命令时才去连接
	data.Rdb = redis.NewClient(option)

	// 2.初始化Mdb
	DSN := c.Database.Source
	db, err := gorm.Open(mysql.Open(DSN), &gorm.Config{})
	if err != nil {
		data.Mdb = nil
		log.Fatal("Mdb初始化失败，", err)
	}
	data.Mdb = db

	// 3.开发阶段，自动迁移表
	migrateTable(db)

	cleanup := func() {
		// 释放连接
		_ = data.Rdb.Close()
		log.NewHelper(logger).Info("closing the data resources")
	}
	return data, cleanup, nil
}

func migrateTable(db *gorm.DB) {
	if err := db.AutoMigrate(&biz.Customer{}); err != nil {
		log.Fatal(err)
	}
}
