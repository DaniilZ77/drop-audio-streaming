package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/golang-jwt/jwt"
)

func validToken(ctx context.Context, tokenString, secret string) (*int, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Log().Error(ctx, "unexpected signing method")
			return nil, core.ErrUnauthorized
		}

		return []byte(secret), nil
	})
	if err != nil {
		logger.Log().Debug(ctx, err.Error())
		return nil, core.ErrUnauthorized
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id, ok := claims["id"].(float64)
		if !ok {
			return nil, core.ErrUnauthorized
		}

		idInt := int(id)
		return &idInt, nil
	}

	return nil, core.ErrUnauthorized
}

func getUserIDFromContext(ctx context.Context) (int, error) {
	id, ok := ctx.Value(userIDContextKey).(int)
	if !ok {
		logger.Log().Debug(ctx, "user id is not provided")
		return 0, core.ErrUnauthorized
	}

	return id, nil
}

func parseRangeHeader(ctx context.Context, req *http.Request) (start, end int64, err error) {
	val := strings.TrimPrefix(req.Header.Get("Range"), "bytes=")
	if val == "" {
		return 0, -1, nil
	}

	tmp := strings.Split(val, "-")
	if len(tmp) != 2 {
		logger.Log().Error(ctx, "invalid range header")
		return 0, 0, core.ErrInvalidRange
	}

	start, err = strconv.ParseInt(tmp[0], 10, 64)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return 0, 0, core.ErrInvalidRange
	}

	if tmp[1] == "" {
		return start, start + 1024*1024, nil
	}
	end, err = strconv.ParseInt(tmp[1], 10, 64)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return 0, 0, core.ErrInvalidRange
	}

	if start < 0 || end < start {
		logger.Log().Error(ctx, "invalid range header")
		return 0, 0, core.ErrInvalidRange
	}

	return start, end, nil
}

func toJSON(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return b, nil
}
