package model

import (
	"net/url"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
)

func ToCoreBeat(req *audiov1.UploadRequest) core.Beat {
	return core.Beat{
		ExternalID:  int(req.GetBeatId()),
		BeatmakerID: int(req.GetBeatmakerId()),
		Artist:      req.GetBeatArtist(),
		Genre:       req.GetBeatGenre(),
	}
}

func ToCoreBeatParams(params url.Values) core.BeatParams {
	return core.BeatParams{
		Artist: params.Get("artist"),
		Genre:  params.Get("genre"),
	}
}
