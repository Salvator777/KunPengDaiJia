package server

import (
	v1 "verifyCode/api/helloworld/v1"
	"verifyCode/api/verifyCode"
	"verifyCode/internal/conf"
	"verifyCode/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server,
	greeter *service.GreeterService,
	// 增加业务需要的服务参数
	VerifyCodeService *service.VerifyCodeService,
	logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterGreeterServer(srv, greeter)
	// 把服务注册到集中的grpc服务器中
	verifyCode.RegisterVerifyCodeServer(srv, VerifyCodeService)
	return srv
}
