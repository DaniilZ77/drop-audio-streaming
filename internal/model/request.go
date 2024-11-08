package model

import (
	"net/url"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
)

func ToCoreBeat(req *audiov1.UploadRequest) core.Beat {
	return core.Beat{
		ID:          int(req.GetBeatId()),
		BeatmakerID: int(req.GetBeatmakerId()),
		Name:        req.GetName(),
		Description: req.GetDescription(),
	}
}

func ToCoreBeatGenre(req *audiov1.UploadRequest) []core.BeatGenre {
	var beatGenre []core.BeatGenre
	for _, genre := range req.GetBeatGenre() {
		beatGenre = append(beatGenre, core.BeatGenre{
			BeatID: int(req.GetBeatId()),
			Genre:  genre,
		})
	}
	return beatGenre
}

func ToCoreBeatParams(params url.Values) core.BeatParams {
	return core.BeatParams{
		Genre: params.Get("genre"),
	}
}
