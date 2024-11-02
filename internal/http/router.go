package http

import (
	"net/http"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type Router struct {
	app         *runtime.ServeMux
	beatService core.BeatService
	chunkSize   int
	jwtSecret   string
}

func NewRouter(app *runtime.ServeMux, beatService core.BeatService, chunkSize int, jwtSecret string) {
	r := &Router{
		app:         app,
		beatService: beatService,
		chunkSize:   chunkSize,
		jwtSecret:   jwtSecret,
	}

	r.initRoutes()
}

func (r *Router) initRoutes() {
	r.app.HandlePath(http.MethodGet, "/v1/audio/{id}", r.stream)
	r.app.HandlePath(http.MethodGet, "/v1/audio", r.ensureValidToken(r.getBeat))
}
