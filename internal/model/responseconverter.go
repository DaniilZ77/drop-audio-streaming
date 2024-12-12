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
		Image       string    `json:"image"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Genres      []int     `json:"genres"`
		Tags        []int     `json:"tags"`
		Moods       []int     `json:"moods"`
		Note        Note      `json:"note"`
		CreatedAt   time.Time `json:"created_at"`
	}

	Note struct {
		NoteID int    `json:"note_id"`
		Scale  string `json:"scale"`
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
		Genres      []int     `json:"genres"`
		Tags        []int     `json:"tags"`
		Moods       []int     `json:"moods"`
		Note        Note      `json:"note"`
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

func ToGetBeatmakerBeatsResponse(beats []core.BeatParams, p core.GetBeatsParams, total int) BeatsPagination {
	var (
		page    = (p.Offset / p.Limit) + 1
		perPage = p.Limit
		pages   = (total + perPage - 1) / perPage
	)

	var beatsPagination []BeatPagination
	for i := 0; i < len(beats); i++ {
		var beatGenres []int
		for _, genre := range beats[i].Genres {
			beatGenres = append(beatGenres, genre.GenreID)
		}
		var beatTags []int
		for _, tag := range beats[i].Tags {
			beatTags = append(beatTags, tag.TagID)
		}
		var beatMoods []int
		for _, mood := range beats[i].Moods {
			beatMoods = append(beatMoods, mood.MoodID)
		}

		beatPagination := BeatPagination{
			ID:          beats[i].Beat.ID,
			Name:        beats[i].Beat.Name,
			Description: beats[i].Beat.Description,
			CreatedAt:   beats[i].Beat.CreatedAt,
			Genres:      beatGenres,
			Tags:        beatTags,
			Moods:       beatMoods,
			Note: Note{
				NoteID: beats[i].Note.NoteID,
				Scale:  beats[i].Note.Scale,
			},
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

func ToGetBeatResponse(beat *core.BeatParams, beatmaker *userv1.GetUserResponse) *audiov1.GetBeatResponse {
	var genres []int64
	for _, genre := range beat.Genres {
		genres = append(genres, int64(genre.GenreID))
	}
	var tags []int64
	for _, tag := range beat.Tags {
		tags = append(tags, int64(tag.TagID))
	}
	var moods []int64
	for _, mood := range beat.Moods {
		moods = append(moods, int64(mood.MoodID))
	}

	return &audiov1.GetBeatResponse{
		Id:          int64(beat.Beat.ID),
		Name:        beat.Beat.Name,
		Description: beat.Beat.Description,
		Beatmaker: &audiov1.Beatmaker{
			Id:        beatmaker.GetUserId(),
			Username:  beatmaker.GetUsername(),
			Pseudonym: beatmaker.GetPseudonym(),
		},
		BeatGenre: genres,
		BeatTag:   tags,
		BeatMood:  moods,
		Note: &audiov1.NoteUpload{
			NoteId: int64(beat.Note.NoteID),
			Scale:  beat.Note.Scale,
		},
		Bpm: int64(beat.Beat.Bpm),
	}
}

func ToBeat(beat *core.BeatParams, beatmaker Beatmaker, image string) Beat {
	var genresInt []int
	var tagsInt []int
	var moodsInt []int
	for _, genre := range beat.Genres {
		genresInt = append(genresInt, genre.GenreID)
	}
	for _, tag := range beat.Tags {
		tagsInt = append(tagsInt, tag.TagID)
	}
	for _, mood := range beat.Moods {
		moodsInt = append(moodsInt, mood.MoodID)
	}

	return Beat{
		ID:          beat.Beat.ID,
		Beatmaker:   beatmaker,
		Name:        beat.Beat.Name,
		Description: beat.Beat.Description,
		Image:       image,
		Genres:      genresInt,
		Tags:        tagsInt,
		Moods:       moodsInt,
		Note: Note{
			NoteID: beat.Note.NoteID,
			Scale:  beat.Note.Scale,
		},
		CreatedAt: beat.Beat.CreatedAt,
	}
}

func ToUploadResponse(fileURL, imageURL string) *audiov1.UploadResponse {
	return &audiov1.UploadResponse{
		FileUploadUrl:  fileURL,
		ImageUploadUrl: imageURL,
	}
}
