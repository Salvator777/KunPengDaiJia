package server

import (
	"context"
	"customer/internal/service"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport"
	jwt2 "github.com/golang-jwt/jwt/v5"
)

func customerJWT(customerService *service.CustomerService) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 1.获取jwt中的顾客id
			claims, ok := jwt.FromContext(ctx)
			if !ok {
				return nil, errors.Unauthorized("Unauthorized", "claims not found")
			}
			claimsMap := claims.(jwt2.MapClaims)
			id := claimsMap["jti"]

			// 2.拿到顾客信息
			Token, err := customerService.CData.GetToken(id)
			if err != nil {
				return nil, errors.Unauthorized("Unauthorized", "customer not found")
			}

			// 3.比对数据表的token和请求的token是否一致
			// 获取请求的token（抄jwt的代码）
			header, _ := transport.FromServerContext(ctx)
			auths := strings.SplitN(header.RequestHeader().Get("Authorization"), " ", 2)
			if len(auths) != 2 || !strings.EqualFold(auths[0], "Bearer") {
				return nil, errors.Unauthorized("UNAUTHORIZED", "JWT token is missing")
			}
			jwtToken := auths[1]
			// 如果不一致
			if Token != jwtToken {
				return nil, errors.Unauthorized("UNAUTHORIZED", "token was updated")
			}

			// 4.所有校验都通过，交由下一个handler处理
			return handler(ctx, req)
		}
	}
}
