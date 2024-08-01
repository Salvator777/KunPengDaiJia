package data

import (
	"context"
	pb "customer/api/customer"
	"customer/api/verifyCode"
	"customer/internal/biz"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	consul "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/random"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/consul/api"
	"gorm.io/gorm"
)

// customer相关数据操作相关的代码
type CustomerData struct {
	data *Data
}

func NewCustomerData(data *Data) *CustomerData {
	return &CustomerData{data: data}
}

// Redis存储手机验证码
func (CData CustomerData) SetVerifyCode(PhoneNum, code string, life int) error {
	status := CData.data.Rdb.Set(context.Background(), "CVC:"+PhoneNum, code, time.Second*time.Duration(life))
	if _, err := status.Result(); err != nil {
		return err
	}
	return nil
}

// grpc调用获取随机6位验证码服务
func (CData CustomerData) GetVerifyCode(PhoneNum string) (*pb.GetVerifyCodeResp, error) {
	// 1.正则校验手机号
	pattern := `^(13\d|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18\d|19[0-35-9])\d{8}$`
	// 使用 regexp.MustCompile 方法将正则表达式编译成一个可以使用的模式对象。
	reqexpPattern := regexp.MustCompile(pattern)
	// 使用 MatchString 方法检查
	if !reqexpPattern.MatchString(PhoneNum) {
		return &pb.GetVerifyCodeResp{
			Code:    201,
			Message: "手机号格式错误",
		}, nil
	}

	// 2.使用服务发现调用验证码服务
	// 拿到服务注册管理器
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "localhost:8500"
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Fatal(err)
	}
	discover := consul.New(consulClient)
	// 设置负载均衡算法
	selector.SetGlobalSelector(random.NewBuilder())
	// selector.GlobalSelector(wrr.NewBuilder())
	// selector.GlobalSelector(p2c.NewBuilder())

	endpoint := "discovery:///VerifyCode"
	conn, err := grpc.DialInsecure(
		context.Background(),
		// grpc.WithEndpoint("localhost:9111"),//验证码的grpc地址
		grpc.WithEndpoint(endpoint),  // 目标服务的名字
		grpc.WithDiscovery(discover), // 使用服务注册管理器来找服务
	)
	if err != nil {
		return &pb.GetVerifyCodeResp{
			Code:    201,
			Message: "验证码服务不可用",
		}, err
	}
	defer conn.Close()

	// 3.发送获取验证码的请求
	client := verifyCode.NewVerifyCodeClient(conn)
	rep, err := client.GetVerifyCode(
		context.Background(),
		&verifyCode.GetVerifyCodeRequest{
			Length: 6,
			Type:   1,
		},
	)
	if err != nil {
		return &pb.GetVerifyCodeResp{
			Code:    201,
			Message: "验证码获取失败",
		}, err
	}
	return &pb.GetVerifyCodeResp{
		VerifyCode: rep.Code,
	}, nil
}

func (CData CustomerData) IsVerifyOK(PhoneNum, verify_code string) bool {
	value := CData.data.Rdb.Get(context.Background(), "CVC:"+PhoneNum)
	if value.String() == "" || value.Val() != verify_code {
		return false
	}
	return true
}

// 通过手机号找用户
func (CData CustomerData) GetCustomerByPhoneNum(PhoneNum string) (*biz.Customer, error) {
	customer := new(biz.Customer)
	// 用手机号去查
	result := CData.data.Mdb.Where("phone_num=?", PhoneNum).First(customer)

	// 查询成功
	if result.Error == nil && customer.ID > 0 {
		return customer, nil
	}

	// 没找到
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// 创建记录并返回
		customer.PhoneNum = PhoneNum
		if result := CData.data.Mdb.Create(customer); result.Error != nil {
			return nil, result.Error
		} else {
			return customer, nil
		}
	}

	// 有错误，但是不是记录不存在的错误
	return nil, result.Error
}

// 生成token并存储
func (CData CustomerData) GenerateTokenAndSave(c *biz.Customer, life time.Duration, secret []byte) (string, error) {
	// 1.处理token的载荷数据
	//标准的JWT的payload
	claims := jwt.RegisteredClaims{
		// 签发机构
		Issuer: "KunPengDJ",
		// 说明
		Subject: "给customer登录使用",
		//签发给谁用
		Audience: []string{"customer", "others"},
		// 有效期至
		ExpiresAt: &jwt.NumericDate{time.Now().Add(life)},
		// 何时启用
		NotBefore: nil,
		// 签发时间
		IssuedAt: &jwt.NumericDate{time.Now()},
		// 存什么都行，这里用来存custmer的id
		ID: fmt.Sprintf("%d", c.ID),
	}

	// 2.生成token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	// 3.签名成功，进行存储
	c.Token = signedToken
	c.TokenCreatedAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	if result := CData.data.Mdb.Save(c); result.Error != nil {
		return "", result.Error
	}

	return signedToken, nil
}

func (CData CustomerData) GetToken(id any) (string, error) {
	c := &biz.Customer{}
	if res := CData.data.Mdb.First(c, id); res.Error != nil {
		return "", res.Error
	}
	return c.Token, nil
}

func (CData CustomerData) DelToken(id any) error {
	// 找到customer
	c := &biz.Customer{}
	if res := CData.data.Mdb.First(c, id); res.Error != nil {
		return res.Error
	}

	// 删除token保存
	c.Token = ""
	c.TokenCreatedAt = sql.NullTime{Valid: false}
	CData.data.Mdb.Save(c)
	return nil
}
