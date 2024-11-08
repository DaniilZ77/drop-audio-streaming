package model

import (
	"time"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	userv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/user"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
)

type (
	Beat struct {
		ID          int       `json:"id"`
		Beatmaker   Beatmaker `json:"beatmaker"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Genre       string    `json:"genre"`
		CreatedAt   time.Time `json:"created_at"`
	}

	Beatmaker struct {
		ID        int    `json:"id"`
		Username  string `json:"username"`
		Pseudonym string `json:"pseudonym"`
	}

	BeatPagination struct {
		ID          int       `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Genre       []string  `json:"genre"`
		CreatedAt   time.Time `json:"created_at"`
	}

	BeatsPagination struct {
		Meta  Meta             `json:"meta"`
		Beats []BeatPagination `json:"beats"`
	}

	Meta struct {
		Page    int `json:"page"`
		PerPage int `json:"per_page"`
		Pages   int `json:"pages"`
		Total   int `json:"total"`
	}
)

func ToBeatmaker(beatmaker *userv1.GetUserResponse) Beatmaker {
	return Beatmaker{
		ID:        int(beatmaker.GetUserId()),
		Username:  beatmaker.GetUsername(),
		Pseudonym: beatmaker.GetPseudonym(),
	}
}

func ToGetBeatmakerBeatsResponse(beats []core.Beat, genres [][]core.BeatGenre, p core.Pagination, total int) BeatsPagination {
	var (
		page    = (p.Offset / p.Limit) + 1
		perPage = p.Limit
		pages   = (total + perPage - 1) / perPage
	)

	var beatsPagination []BeatPagination
	for i := 0; i < len(beats); i++ {
		var beatGenres []string
		for _, genre := range genres[i] {
			beatGenres = append(beatGenres, genre.Genre)
		}

		beatPagination := BeatPagination{
			ID:          beats[i].ID,
			Name:        beats[i].Name,
			Description: beats[i].Description,
			CreatedAt:   beats[i].CreatedAt,
			Genre:       beatGenres,
		}

		beatsPagination = append(beatsPagination, beatPagination)
	}

	return BeatsPagination{
		Meta: Meta{
			PerPage: p.Limit,
			Total:   total,
			Pages:   pages,
			Page:    page,
		},
		Beats: beatsPagination,
	}
}

func ToGetBeatMetaResponse(beat *core.Beat, beatmaker *userv1.GetUserResponse, beatGenres []core.BeatGenre) *audiov1.GetBeatMetaResponse {
	genres := make([]string, 0)
	for _, genre := range beatGenres {
		genres = append(genres, genre.Genre)
	}

	return &audiov1.GetBeatMetaResponse{
		Id:          int64(beat.ID),
		Name:        beat.Name,
		Description: beat.Description,
		Beatmaker: &audiov1.Beatmaker{
			Id:        beatmaker.GetUserId(),
			Username:  beatmaker.GetUsername(),
			Pseudonym: beatmaker.GetPseudonym(),
		},
		BeatGenre: genres,
	}
}

func ToBeat(beat *core.Beat, beatmaker Beatmaker, genre string) Beat {
	return Beat{
		ID:          beat.ID,
		Beatmaker:   beatmaker,
		Name:        beat.Name,
		Description: beat.Description,
		Genre:       genre,
		CreatedAt:   beat.CreatedAt,
	}
}
