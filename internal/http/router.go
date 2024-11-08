package http

import (
	"net/http"

	userclient "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/client/user/grpc"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type Router struct {
	app         *runtime.ServeMux
	beatService core.BeatService
	chunkSize   int
	jwtSecret   string
	userClient  *userclient.Client
}

func NewRouter(
	app *runtime.ServeMux,
	beatService core.BeatService,
	chunkSize int,
	jwtSecret string,
	userClient *userclient.Client) {
	r := &Router{
		app:         app,
		beatService: beatService,
		chunkSize:   chunkSize,
		jwtSecret:   jwtSecret,
		userClient:  userClient,
	}

	r.initRoutes()
}

func (r *Router) initRoutes() {
	r.app.HandlePath(http.MethodGet, "/v1/audio/{id}", r.stream)
	r.app.HandlePath(http.MethodGet, "/v1/audio/play", r.ensureValidToken(r.getBeat))
}
