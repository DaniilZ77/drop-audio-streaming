package model

import (
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
)

type Beat struct {
	ID          int       `json:"id"`
	Artist      string    `json:"artist"`
	Genre       string    `json:"genre"`
	BeatmakerID int       `json:"beatmaker_id"`
	CreatedAt   time.Time `json:"created_at"`
}

func ToBeat(beat *core.Beat) Beat {
	return Beat{
		ID:          beat.ExternalID,
		Artist:      beat.Artist,
		Genre:       beat.Genre,
		BeatmakerID: beat.BeatmakerID,
		CreatedAt:   beat.CreatedAt,
	}
}
