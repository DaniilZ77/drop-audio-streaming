package model

import (
	"time"

	userv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/user"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
)

type Beat struct {
	ID        int       `json:"id"`
	Beatmaker Beatmaker `json:"beatmaker"`
	Genre     *string   `json:"genre,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Beatmaker struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Pseudonym string `json:"pseudonym"`
}

func ToBeatmaker(beatmaker *userv1.GetUserResponse) Beatmaker {
	return Beatmaker{
		ID:        int(beatmaker.GetUserId()),
		Username:  beatmaker.GetUsername(),
		Pseudonym: beatmaker.GetPseudonym(),
	}
}

func ToBeat(beat *core.Beat, beatmaker Beatmaker, genre *string) Beat {
	return Beat{
		ID:        beat.ID,
		Beatmaker: beatmaker,
		Genre:     genre,
		CreatedAt: beat.CreatedAt,
	}
}
