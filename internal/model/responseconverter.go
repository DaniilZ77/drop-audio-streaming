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
		Genres      []int64   `json:"genres"`
		Tags        []int64   `json:"tags"`
		Moods       []int64   `json:"moods"`
		Note        Note      `json:"note"`
		Bpm         int       `json:"bpm"`
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
		Bpm         int       `json:"bpm"`
		Genres      []int64   `json:"genres"`
		Tags        []int64   `json:"tags"`
		Moods       []int64   `json:"moods"`
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
		genres := toResponseGenres(beats[i])
		tags := toResponseTags(beats[i])
		moods := toResponseMoods(beats[i])

		beatPagination := BeatPagination{
			ID:          beats[i].Beat.ID,
			Name:        beats[i].Beat.Name,
			Description: beats[i].Beat.Description,
			CreatedAt:   beats[i].Beat.CreatedAt,
			Genres:      genres,
			Tags:        tags,
			Moods:       moods,
			Note: Note{
				NoteID: beats[i].Note.NoteID,
				Scale:  beats[i].Note.Scale,
			},
			Bpm: beats[i].Beat.Bpm,
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

func toResponseGenres(beat core.BeatParams) []int64 {
	var genres []int64
	for _, genre := range beat.Genres {
		genres = append(genres, int64(genre.GenreID))
	}
	return genres
}

func toResponseTags(beat core.BeatParams) []int64 {
	var tags []int64
	for _, tag := range beat.Tags {
		tags = append(tags, int64(tag.TagID))
	}
	return tags
}

func toResponseMoods(beat core.BeatParams) []int64 {
	var moods []int64
	for _, mood := range beat.Moods {
		moods = append(moods, int64(mood.MoodID))
	}
	return moods
}

func ToGetBeatResponse(beat *core.BeatParams, beatmaker *userv1.GetUserResponse) *audiov1.GetBeatResponse {
	genres := toResponseGenres(*beat)
	tags := toResponseTags(*beat)
	moods := toResponseMoods(*beat)

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
	genres := toResponseGenres(*beat)
	tags := toResponseTags(*beat)
	moods := toResponseMoods(*beat)

	return Beat{
		ID:          beat.Beat.ID,
		Beatmaker:   beatmaker,
		Name:        beat.Beat.Name,
		Description: beat.Beat.Description,
		Image:       image,
		Genres:      genres,
		Tags:        tags,
		Moods:       moods,
		Note: Note{
			NoteID: beat.Note.NoteID,
			Scale:  beat.Note.Scale,
		},
		Bpm:       beat.Beat.Bpm,
		CreatedAt: beat.Beat.CreatedAt,
	}
}

func ToUploadResponse(fileURL, imageURL string) *audiov1.UploadResponse {
	return &audiov1.UploadResponse{
		FileUploadUrl:  fileURL,
		ImageUploadUrl: imageURL,
	}
}

func toGetFiltersGenre(genre core.Genre) *audiov1.Genre {
	return &audiov1.Genre{
		GenreId: int64(genre.ID),
		Name:    genre.Name,
	}
}

func toGetFiltersTag(tag core.Tag) *audiov1.Tag {
	return &audiov1.Tag{
		TagId: int64(tag.ID),
		Name:  tag.Name,
	}
}

func toGetFiltersMood(mood core.Mood) *audiov1.Mood {
	return &audiov1.Mood{
		MoodId: int64(mood.ID),
		Name:   mood.Name,
	}
}

func toGetFiltersNote(note core.Note) *audiov1.Note {
	return &audiov1.Note{
		NoteId: int64(note.ID),
		Name:   note.Name,
	}
}

func ToGetFiltersResponse(filters core.Filters) *audiov1.GetFiltersResponse {
	var genres []*audiov1.Genre
	var tags []*audiov1.Tag
	var moods []*audiov1.Mood
	var notes []*audiov1.Note

	for i := range filters.Genres {
		genres = append(genres, toGetFiltersGenre(filters.Genres[i]))
	}
	for i := range filters.Tags {
		tags = append(tags, toGetFiltersTag(filters.Tags[i]))
	}
	for i := range filters.Moods {
		moods = append(moods, toGetFiltersMood(filters.Moods[i]))
	}
	for i := range filters.Note {
		notes = append(notes, toGetFiltersNote(filters.Note[i]))
	}

	return &audiov1.GetFiltersResponse{
		Genres: genres,
		Tags:   tags,
		Moods:  moods,
		Notes:  notes,
	}
}
