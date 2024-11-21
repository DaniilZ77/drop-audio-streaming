package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func (r *Router) ensureValidToken(next runtime.HandlerFunc) runtime.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request, params map[string]string) {
		ctx := req.Context()

		tokenString := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer")
		tokenString = strings.TrimSpace(tokenString)

		id, err := validToken(ctx, tokenString, r.jwtSecret)
		if err != nil {
			logger.Log().Debug(ctx, err.Error())
			http.Error(w, core.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(ctx, userIDContextKey, *id)

		next(w, req.WithContext(ctx), params)
	}
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Log().Info(r.Context(), "%s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)
	})
}
