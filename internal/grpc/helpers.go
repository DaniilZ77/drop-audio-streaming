package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model"
	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthMiddleware(secret string, requireAdmin map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if !requireAdmin[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok || len(md.Get("authorization")) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "%s: %s", model.ErrUnauthorized.Error(), "token not provided")
		}

		data := strings.Fields(md.Get("authorization")[0])
		if len(data) < 2 || strings.ToLower(data[0]) != "bearer" {
			return nil, status.Errorf(codes.Unauthenticated, "%s: %s", model.ErrUnauthorized.Error(), "invalid header format")
		}

		token := data[1]
		admin, err := validateToken(token, secret)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "%s: %s", model.ErrUnauthorized.Error(), err.Error())
		}

		if model.AdminScale(*admin) != model.AdminScaleMinor && model.AdminScale(*admin) != model.AdminScaleMajor {
			return nil, status.Errorf(codes.PermissionDenied, "%s: %s", model.ErrUnauthorized, "must be admin")
		}

		return handler(ctx, req)
	}
}

func validateToken(token, secret string) (*string, error) {
	data, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %s", model.ErrUnauthorized, "unexpected signing method")
		}

		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", model.ErrUnauthorized, err)
	}

	if claims, ok := data.Claims.(jwt.MapClaims); ok && data.Valid {
		admin, _ := claims["admin"].(string)
		return &admin, nil
	}

	return nil, model.ErrUnauthorized
}
