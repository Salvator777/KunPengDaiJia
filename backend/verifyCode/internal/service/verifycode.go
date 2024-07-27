package service

import (
	"context"
	"math/rand"

	pb "verifyCode/api/verifyCode"
)

type VerifyCodeService struct {
	pb.UnimplementedVerifyCodeServer
}

func NewVerifyCodeService() *VerifyCodeService {
	return &VerifyCodeService{}
}

func (s *VerifyCodeService) GetVerifyCode(ctx context.Context, req *pb.GetVerifyCodeRequest) (*pb.GetVerifyCodeReply, error) {
	return &pb.GetVerifyCodeReply{
		Code: RandomCode(int(req.Length), req.Type),
	}, nil
}

func RandomCode(l int, t pb.TYPE) string {
	switch t {
	case pb.TYPE_DEFAULT:
		fallthrough
	case pb.TYPE_DIGIT:
		chars := "0123456789"
		return randCode(chars, l)
	case pb.TYPE_LETTER:
		chars := "abcdefghijklmnopqrstuvwxyz"
		return randCode(chars, l)

	case pb.TYPE_MIXED:
		chars := "0123456789abcdefghijklmnopqrstuvwxyz"
		return randCode(chars, l)
	default:

	}
	return ""
}

func randCode(chars string, l int) string {
	buf := make([]byte, l)
	Len := len(chars)
	for i := 0; i < l; i++ {
		// Intn函数随机返回0~Len的整数
		index := rand.Intn(Len)
		buf[i] = chars[index]
	}
	return string(buf)
}
