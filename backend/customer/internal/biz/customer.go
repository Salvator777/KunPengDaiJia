package biz

import (
	"context"
	pb "customer/api/customer"
	"customer/api/valuation"
	"database/sql"
	"log"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
	"gorm.io/gorm"
)

// gorm的模型
type Customer struct {
	// 嵌入4个基础字段
	gorm.Model
	// 业务逻辑
	CustomerWork
	// token部分
	CustomerToken
}

// 业务逻辑部分
type CustomerWork struct {
	PhoneNum string         `gorm:"type: varchar(15);uniqueIndex" json:"phone_num,omitempty"`
	Name     sql.NullString `gorm:"type: varchar(15);uniqueIndex" json:"name,omitempty"`
	Email    sql.NullString `gorm:"type: varchar(255);uniqueIndex" json:"email,omitempty"`
	Wechat   sql.NullString `gorm:"type: varchar(255);uniqueIndex" json:"wechat,omitempty"`
	CityID   uint           `gorm:"index;" json:"cityid,omitempty"`
}

type CustomerToken struct {
	Token          string       `gorm:"type: varchar(4095);" json:"token,omitempty"`
	TokenCreatedAt sql.NullTime `gorm:"" json:"token_created_at,omitempty"`
}

const CustomerSecret = "MySecretKey"
const CustomerTokenLife = 3600 * 24 * 30 * 2

type CustomerBiz struct {
}

// 只要定义了这种封装的结构体，在其他地方要注入
// 就要写New函数并且加入provider，养成好习惯
func NewCustomerBiz() *CustomerBiz {
	return &CustomerBiz{}
}

func (cb *CustomerBiz) GetEstimatePrice(origin, destination string) (*pb.EstimatePriceResp, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "localhost:8500"
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Fatal(err)
	}
	discover := consul.New(consulClient)

	endpoint := "discovery:///Valuation"
	conn, err := grpc.DialInsecure(
		context.Background(),
		// grpc.WithEndpoint("localhost:9111"),//验证码的grpc地址
		grpc.WithEndpoint(endpoint),  // 目标服务的名字
		grpc.WithDiscovery(discover), // 使用服务注册管理器来找服务
	)
	if err != nil {
		return &pb.EstimatePriceResp{
			Code:    201,
			Message: "价格估计服务不可用",
		}, err
	}
	defer conn.Close()

	// 3.发送获取验证码的请求
	client := valuation.NewValuationClient(conn)
	rep, err := client.GetEstimatePrice(
		context.Background(),
		&valuation.GetEstimatePriceReq{
			Origin:      origin,
			Destination: destination,
		},
	)
	if err != nil {
		return &pb.EstimatePriceResp{
			Code:    201,
			Message: "估计价格获取失败",
		}, err
	}
	return &pb.EstimatePriceResp{
		Code:        200,
		Message:     "get estimate price success",
		Origin:      rep.Origin,
		Destination: destination,
		Price:       rep.Price,
	}, nil
}
