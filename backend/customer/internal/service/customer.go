package service

import (
	"context"
	"time"

	pb "customer/api/customer"
	"customer/internal/data"
)

// 所有与customer相关的代码放这里
type CustomerService struct {
	pb.UnimplementedCustomerServer
	// 与customer的数据相关
	CData *data.CustomerData
}

func NewCustomerService(CData *data.CustomerData) *CustomerService {
	return &CustomerService{CData: CData}
}

func (s *CustomerService) GetVerifyCode(ctx context.Context, req *pb.GetVerifyCodeReq) (*pb.GetVerifyCodeResp, error) {
	// 1.获取验证码
	rep, err := s.CData.GetVerifyCode(req.PhoneNum)
	if err != nil {
		return &pb.GetVerifyCodeResp{
			Code:    201,
			Message: "验证码获取失败",
		}, err
	}

	const life = 60
	// 2.redis临时存储
	if err := s.CData.SetVerifyCode(req.PhoneNum, rep.VerifyCode, life); err != nil {
		return &pb.GetVerifyCodeResp{
			Code:    201,
			Message: "验证码存储失败(Redis的set操作失败)",
		}, err
	}
	// 返回响应
	return &pb.GetVerifyCodeResp{
		Code:           200,
		VerifyCode:     rep.VerifyCode,
		VerifyCodeTime: time.Now().Unix(),
		VerifyCodeLift: int64(life),
	}, nil
}

func (s *CustomerService) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	// 1.校验手机号和验证码
	if !s.CData.IsVerifyOK(req.PhoneNum, req.VerifyCode) {
		return &pb.LoginResp{
			Code:    201,
			Message: "验证码错误",
		}, nil
	}

	// 2.判断手机号是否已注册
	// 返回手机号对应的用户，如果没有就插入数据后再返回
	customer, err := s.CData.GetCustomerByPhoneNum(req.PhoneNum)
	if err != nil {
		return &pb.LoginResp{
			Code:    201,
			Message: "顾客信息获取错误",
		}, nil
	}

	// 3.设置token，jwt
	const secret = "MySecretKey"
	const life = 3600 * 24 * 30 * 2
	// 有效期两个月
	token, err := s.CData.GenerateTokenAndSave(customer, time.Second*life, []byte(secret))
	if err != nil {
		return &pb.LoginResp{
			Code:    201,
			Message: "token生成错误",
		}, nil
	}

	// 4.响应token
	return &pb.LoginResp{
		Code:      200,
		Message:   "login success",
		Token:     token,
		TokenTime: time.Now().Unix(),
		TokenLift: life,
	}, nil
}
