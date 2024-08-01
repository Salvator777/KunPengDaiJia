package service

import (
	"context"
	"time"

	pb "customer/api/customer"
	"customer/internal/biz"
	"customer/internal/data"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwt2 "github.com/golang-jwt/jwt/v5"
)

// 所有与customer相关的代码放这里
type CustomerService struct {
	pb.UnimplementedCustomerServer
	// 与customer的数据相关
	CData *data.CustomerData
	CB    *biz.CustomerBiz
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

	// 3.生成token并存储
	// 有效期两个月
	token, err := s.CData.GenerateTokenAndSave(customer, time.Second*biz.CustomerTokenLife, []byte(biz.CustomerSecret))
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
		TokenLift: biz.CustomerTokenLife,
	}, nil
}

func (s *CustomerService) Logout(ctx context.Context, req *pb.LogoutReq) (*pb.LogoutResp, error) {
	// 1.获取用户id
	claims, _ := jwt.FromContext(ctx)
	claimsMap := claims.(jwt2.MapClaims)
	id := claimsMap["jti"]

	// 2.删除用户token
	err := s.CData.DelToken(id)
	if err != nil {
		return &pb.LogoutResp{
			Code:    201,
			Message: "token删除失败",
		}, nil
	}

	// 3.响应
	return &pb.LogoutResp{
		Code:    200,
		Message: "logout success",
	}, nil
}

func (s *CustomerService) EstimatePrice(ctx context.Context, req *pb.EstimatePriceReq) (*pb.EstimatePriceResp, error) {
	res, err := s.CB.GetEstimatePrice(req.Origin, req.Destination)
	if err != nil {
		return nil, errors.New(200, "PRICE ERROR", "cal price error")
	}
	return res, nil
}
