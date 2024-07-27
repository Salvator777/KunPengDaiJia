package data

import (
	"context"
	pb "customer/api/customer"
	"customer/api/verifyCode"
	"regexp"
	"time"

	"github.com/go-kratos/kratos/v2/transport/grpc"
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

// 获取验证码
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

	// 2.连接验证码服务
	conn, err := grpc.DialInsecure(context.Background(), grpc.WithEndpoint("localhost:9000")) //验证码的grpc地址
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
