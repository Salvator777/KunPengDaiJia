package data

import (
	"valuation/internal/biz"
	"valuation/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewPriceRuleData)

// Data .
type Data struct {
	// TODO wrapped database client
	Mdb *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	data := new(Data)

	// 初始化Mdb
	DSN := c.Database.Source
	db, err := gorm.Open(mysql.Open(DSN), &gorm.Config{})
	if err != nil {
		data.Mdb = nil
		log.Fatal("Mdb初始化失败，", err)
	}
	data.Mdb = db

	// 开发阶段，自动迁移表
	migrateTable(db)
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return data, cleanup, nil
}

func migrateTable(db *gorm.DB) {
	if err := db.AutoMigrate(&biz.PriceRule{}); err != nil {
		log.Fatal("price_rule table migrate error, err:", err)
	}
	// 插入一些riceRule的测试数据
	rules := []biz.PriceRule{
		{
			Model: gorm.Model{ID: 1},
			PriceRuleWork: biz.PriceRuleWork{
				CityID:      1,
				StartFee:    300,
				DistanceFee: 35,
				DurationFee: 10, // 5m
				StartAt:     7,
				EndedAt:     23,
			},
		},
		{
			Model: gorm.Model{ID: 2},
			PriceRuleWork: biz.PriceRuleWork{
				CityID:      1,
				StartFee:    350,
				DistanceFee: 35,
				DurationFee: 10, // 5m
				StartAt:     23,
				EndedAt:     24,
			},
		},
		{
			Model: gorm.Model{ID: 3},
			PriceRuleWork: biz.PriceRuleWork{
				CityID:      1,
				StartFee:    400,
				DistanceFee: 35,
				DurationFee: 10, // 5m
				StartAt:     0,
				EndedAt:     7,
			},
		},
	}
	// 使用 GORM 的 Clauses 方法来处理数据库中的冲突情况
	// 插入数据如果遇到主键冲突，则更新所有字段，保证插入不会失败
	db.Clauses(clause.OnConflict{UpdateAll: true}).Create(rules)
}
