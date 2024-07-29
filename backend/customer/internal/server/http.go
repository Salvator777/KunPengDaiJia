package server

import (
	"context"
	"customer/api/customer"
	v1 "customer/api/helloworld/v1"
	"customer/internal/biz"
	"customer/internal/conf"
	"customer/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwt2 "github.com/golang-jwt/jwt/v5"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server,
	greeter *service.GreeterService,
	customerService *service.CustomerService,
	logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		// 这里配置http的一些选项，下面的Middleware里就是加入的中间件
		http.Middleware(
			recovery.Recovery(),
			// 自己设置的中间件
			// token *jwt2.Token这个参数是
			selector.Server(jwt.Server(func(token *jwt2.Token) (interface{}, error) {
				// 传一下校验的秘钥
				return []byte(biz.CustomerSecret), nil
			}), customerJWT(customerService)).Match(func(ctx context.Context, operation string) bool {
				// operation就是api请求的路由
				// 根据operation来设置那些业务用中间件校验，哪些不用
				noTokenCheck := map[string]struct{}{
					"/api.customer.Customer/GetVerifyCode": {},
					"/api.customer.Customer/Login":         {}}
				log.Info(operation)
				if _, exist := noTokenCheck[operation]; exist {
					return false
				}
				return true
			}).Build(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterGreeterHTTPServer(srv, greeter)
	customer.RegisterCustomerHTTPServer(srv, customerService)
	return srv
}
