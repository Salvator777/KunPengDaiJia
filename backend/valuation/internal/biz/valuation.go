package biz

import (
	"context"
	"log"
	"strconv"
	"valuation/api/mymap"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
	"gorm.io/gorm"
)

// 计价规则的表
type PriceRule struct {
	gorm.Model
	PriceRuleWork
}

type PriceRuleWork struct {
	CityID      uint
	StartFee    int64
	DistanceFee int64
	DurationFee int64
	StartAt     int // 0 [0
	EndedAt     int // 7 0)
}

// 定义操作priceRule的接口
type PriceRuleInterface interface {
	// 根据城市和当前时间获取规则
	GetRule(cityid uint, curr int) (*PriceRule, error)
}

type ValuationBiz struct {
	pri PriceRuleInterface
}

func NewValuationBiz(pri PriceRuleInterface) *ValuationBiz {
	return &ValuationBiz{
		pri: pri,
	}
}

func (v *ValuationBiz) GetDrivingInfoMyMap(origin, destination string) (string, string, error) {
	// 1.拿到服务注册器
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "localhost:8500"
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Fatal(err)
	}
	discover := consul.New(consulClient)

	endpoint := "discovery:///mymap"
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(endpoint),  // 目标服务的名字
		grpc.WithDiscovery(discover), // 使用服务注册管理器来找服务
	)
	if err != nil {
		return "", "", err
	}
	defer conn.Close()

	// 3.发送获取验证码的请求
	client := mymap.NewMymapClient(conn)
	rep, err := client.GetDrivingInfo(
		context.Background(),
		&mymap.GetDrivingInfoReq{
			Origin:      origin,
			Destination: destination,
		},
	)
	if err != nil {
		return "", "", err
	}

	return rep.Distance, rep.Duration, nil
}

// 获取价格
func (vb *ValuationBiz) GetPrice(ctx context.Context, distance, duration string, cityId uint, curr int) (int64, error) {

	// 一，获取规则
	rule, err := vb.pri.GetRule(cityId, curr)
	if err != nil {
		return 0, err
	}
	// 二，距离和时长是string类型，需要转换为int64
	distancInt64, err := strconv.ParseInt(distance, 10, 64)
	if err != nil {
		return 0, err
	}
	durationInt64, err := strconv.ParseInt(duration, 10, 64)
	if err != nil {
		return 0, err
	}

	// 三，基于rule计算
	// 起步价包含的距离
	var startDistance int64 = 5
	total := rule.StartFee +
		rule.DistanceFee*(distancInt64/1000-startDistance) +
		rule.DurationFee*durationInt64/60

	return total, nil
}
