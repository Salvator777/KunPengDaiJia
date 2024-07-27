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
	return &pb.LoginResp{}, nil
}
