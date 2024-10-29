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
}

func NewRouter(app *runtime.ServeMux, beatService core.BeatService, chunkSize int) {
	r := &Router{
		app:         app,
		beatService: beatService,
		chunkSize:   chunkSize,
	}

	r.initRoutes()
}

func (r *Router) initRoutes() {
	r.app.HandlePath(http.MethodGet, "/v1/audio/{id}", r.stream)
}
