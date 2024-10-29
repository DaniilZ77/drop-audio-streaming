package grpc

import (
	audiostreamingv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audiostreaming"
	"google.golang.org/grpc"
)

type server struct {
	audiostreamingv1.UnimplementedAudioStreamingServiceServer
}

func Register(gRPCServer *grpc.Server) {
	audiostreamingv1.RegisterAudioStreamingServiceServer(gRPCServer, &server{})
}
